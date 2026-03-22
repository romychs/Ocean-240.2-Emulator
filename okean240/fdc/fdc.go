package fdc

/**
 * Floppy drive controller, based on
 * MB8877, К1818ВГ93
 *
 * By Romych, 2025.03.05
 */

import (
	"bytes"
	"encoding/binary"
	"errors"
	"okemu/config"
	"os"
	"slices"
	"strconv"

	"github.com/howeyc/crc16"
	log "github.com/sirupsen/logrus"
)

// Floppy parameters
const (
	FloppyB = 0
	FloppyC = 1

	TotalDrives    = 2
	FloppySizeK    = 720
	SectorSize     = 512
	SideCount      = 2
	SectorPerTrack = 9

	SizeInSectors  = FloppySizeK * 1024 / SectorSize
	TracksCount    = SizeInSectors / SideCount / SectorPerTrack
	SectorsPerSide = SizeInSectors / SideCount

	TrackHeaderSize = 146
	TrackSectorSize = 626
	TrackFooterSize = 256 * 3
	TrackBufferSize = TrackHeaderSize + TrackSectorSize*SectorPerTrack + TrackFooterSize
)

// FDC Commands
const (
	CmdRestore         = 0x0
	CmdSeek            = 0x1
	CmdStep            = 0x2
	CmdStepIn          = 0x5
	CmdStepOut         = 0x7
	CmdReadSector      = 0x8
	CmdReadSectorMulti = 0x9
	CmdWriteSector     = 0xa
	CmdReadAddress     = 0xc
	CmdWriteTrack      = 0xf
	CmdNoCommand       = 0xff
)

var interleave = []byte{1, 8, 6, 4, 2, 9, 7, 5, 3}

const (
	StatusTR0 = 0x04 // TR0 - Head at track 0
	StatusRNF = 0x10 // RNF - Record not found
	//	StatusSeekError  = 0x10 // Sector out of disk
	StatusHeadLoaded = 0x20 // Head on disk
)

type SectorType []byte

type FloppyDriveController struct {
	// Floppy controller port
	sideNo   byte
	ddEn     byte
	init     byte
	drive    byte
	mot1     byte
	mot0     byte
	intRq    byte
	motSt    byte
	sectorNo byte
	trackNo  byte
	drq      byte
	// FloppyStorage B and C
	sectors [TotalDrives][SizeInSectors]SectorType
	data    byte
	status  byte
	lastCmd byte
	//curSector   *SectorType
	bytePtr     uint16
	trackBuffer []byte
	//	floppyFile  []string
	config *config.OkEmuConfig
}

type FloppyDriveControllerInterface interface {
	SetFloppy()
	Floppy() byte
	SetCmd(value byte)
	Status() byte
	SetTrack(value byte)
	SetSector(value byte)
	SetData(value byte)
	Data() byte
	Drq() byte
	SaveFloppy(drive byte)
	GetSectorNo() uint16
	Track() byte
	Sector() byte
}

//var slicer = []uint16{1, 8, 6, 4, 2, 9, 7, 5, 3}

func getSectorNo(side byte, track byte, sector byte) uint16 {
	//return (uint16(track)*18 + uint16(side)*9 + slicer[sector-1] - 1)
	return uint16(side)*SectorsPerSide + uint16(track)*SectorPerTrack + uint16(sector) - 1
}

func (f *FloppyDriveController) GetSectorNo() uint16 {
	return getSectorNo(f.sideNo, f.trackNo, f.sectorNo)
}

func (f *FloppyDriveController) SetFloppy(val byte) {
	// WR: 5-SSEL, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
	f.sideNo = val >> 5 & 0x01
	f.ddEn = val >> 4 & 0x01
	f.init = val >> 3 & 0x01
	f.drive = (^val) >> 2 & 0x01
	f.mot1 = val >> 1 & 0x01
	f.mot0 = val & 0x01
}

func (f *FloppyDriveController) GetFloppy() byte {
	// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT
	floppy := f.intRq | (f.mot0 << 1) | (f.mot1 << 2) | ((^f.drive & 1) << 3) | (f.sideNo << 6) | (f.motSt << 7)
	return floppy
}

var crcTable *crc16.Table

func init() {
	crcTable = crc16.MakeTable(0xffff)
}

func (f *FloppyDriveController) SetCmd(value byte) {
	f.lastCmd = value >> 4
	switch f.lastCmd {
	case CmdRestore:
		log.Trace("CMD Restore (seek trackNo 0)")
		f.trackNo = 0
		f.status = StatusTR0 | StatusHeadLoaded // TR0 & Head loaded
	case CmdSeek:
		log.Tracef("CMD Seek %x", value&0xf)
		f.status = StatusHeadLoaded
		f.trackNo = f.data
	case CmdStep:
		log.Tracef("CMD Step %x", value&0xf)
		f.status = StatusHeadLoaded
		f.trackNo = f.data
	case CmdStepIn:
		log.Tracef("CMD StepIn (Next track) %x", value&0xf)
		f.status = StatusHeadLoaded
		if f.trackNo < TracksCount {
			f.trackNo++
		}
	case CmdStepOut:
		log.Tracef("CMD StepOut (Previous track) %x", value&0xf)
		f.status = StatusHeadLoaded
		if f.trackNo > 0 {
			f.trackNo--
		}
	case CmdReadSector:
		sectorNo := f.GetSectorNo()
		log.Tracef("CMD Read single sectorNo: %d", sectorNo)
		if sectorNo < SizeInSectors {
			f.trackBuffer = slices.Clone(f.sectors[f.drive][sectorNo])
			f.drq = 1
			f.status = 0x00
		} else {
			f.drq = 0
			f.status = StatusRNF
		}
	case CmdReadSectorMulti:
		sectorNo := f.GetSectorNo()
		f.trackBuffer = []byte{}
		for c := 0; c < SectorPerTrack; c++ {
			f.trackBuffer = slices.Concat(f.trackBuffer, f.sectors[f.drive][sectorNo])
			sectorNo++
		}
		f.drq = 1
		f.status = 0x0
	case CmdWriteSector:
		sectorNo := f.GetSectorNo()
		log.Tracef("CMD Write Sector %d", sectorNo)
		if sectorNo < SizeInSectors {
			f.bytePtr = 0
			f.drq = 1
			f.status = 0x0
			f.trackBuffer = []byte{}
		} else {
			f.drq = 0
			f.status = StatusRNF
		}
	case CmdReadAddress:
		log.Tracef("CMD ReadAddress %d", value)
		f.trackBuffer = []byte{f.trackNo, f.sideNo, f.sectorNo, 2}

		checksum := crc16.Checksum(f.trackBuffer, crcTable)
		f.trackBuffer = append(f.trackBuffer, byte(checksum))
		f.trackBuffer = append(f.trackBuffer, byte(checksum>>8))

		f.drq = 1
		f.status = 0x0
	case CmdWriteTrack:
		log.Tracef("CMD Write Track %x", f.trackNo)
		f.drq = 1
		f.status = 0x00
		f.trackBuffer = []byte{}
	default:
		log.Errorf("Unknown CMD: %x VAL: %x", f.lastCmd, value&0xf)
	}
}

func (f *FloppyDriveController) Status() byte {
	return f.status
}

func (f *FloppyDriveController) SetTrackNo(value byte) {
	//log.Tracef("FDC Track: %d", value)
	if value > TracksCount {
		f.status |= 0x10 /// RNF
		log.Errorf("Track %d not found!", value)
	} else {
		f.trackNo = value
	}
}

func (f *FloppyDriveController) SetSectorNo(value byte) {
	//log.Tracef("FDC Sector: %d", value)
	if value > SectorPerTrack {
		f.status |= 0x10
		log.Errorf("Record not found %d!", value)
	} else {
		f.sectorNo = value
	}
}

func (f *FloppyDriveController) SetData(value byte) {
	//log.Tracef("FCD Data: %d", value)
	if f.lastCmd == CmdWriteTrack {
		if len(f.trackBuffer) < TrackBufferSize {
			f.trackBuffer = append(f.trackBuffer, value)
			f.drq = 1
			f.status = 0x00
		} else {
			//f.dump()
			f.writeTrack()
			f.drq = 0
			f.status = 0x00
			f.lastCmd = CmdNoCommand
		}
	} else if f.lastCmd == CmdWriteSector {
		if len(f.trackBuffer) < SectorSize {
			f.trackBuffer = append(f.trackBuffer, value)
			if len(f.trackBuffer) == SectorSize {
				f.drq = 0
			} else {
				f.drq = 1
			}
		}
		if len(f.trackBuffer) == SectorSize {
			f.drq = 0
			f.sectors[f.drive][f.GetSectorNo()] = slices.Clone(f.trackBuffer)
			f.lastCmd = CmdNoCommand
		}
	}
	f.data = value
}

const SectorInfoSize = 626
const SectorInfoOffset = 0x0092
const TrackNoOffset = 0x0010
const SideNoOffset = 0x0011
const SectorNoOffset = 0x0012
const SectorLengthOffset = 0x0013
const SectorDataOffset = 0x003b

var SectorLengths = []int{128, 256, 512, 1024}

func (f *FloppyDriveController) writeTrack() {
	// skip header
	ptr := SectorInfoOffset
	// repeat for every sector on track
	for sec := 0; sec < SectorPerTrack; sec++ {
		// get info from header
		trackNo := f.trackBuffer[ptr+TrackNoOffset]
		sideNo := f.trackBuffer[ptr+SideNoOffset]
		sectorNo := f.trackBuffer[ptr+SectorNoOffset]
		sectorLength := SectorLengths[f.trackBuffer[ptr+SectorLengthOffset]&0x03]
		// get sector data
		sectorData := f.trackBuffer[ptr+SectorDataOffset : ptr+SectorDataOffset+sectorLength]
		absSector := getSectorNo(sideNo, trackNo, sectorNo)
		log.Debugf("Write Drive: %d; side:%d; T: %d S: %d Len: %d  Data: [%X..%X]; Abs sector: %d", f.drive, sideNo, trackNo, sectorNo, len(sectorData), sectorData[0], sectorData[len(sectorData)-1], absSector)
		// write data to sector buffer
		f.sectors[f.drive][absSector] = slices.Clone(sectorData)
		// shift pointer to next sector info block
		ptr += SectorInfoSize
	}
}

func (f *FloppyDriveController) Data() byte {
	switch f.lastCmd {
	case CmdReadSector, CmdReadSectorMulti, CmdReadAddress:
		if len(f.trackBuffer) > 0 {
			f.drq = 1
			f.data = f.trackBuffer[0]
			f.trackBuffer = f.trackBuffer[1:]
		}
		if len(f.trackBuffer) == 0 {
			f.drq = 0
			f.status = 0
			f.lastCmd = CmdNoCommand
		}
	default:
		f.data = 0xff
	}
	return f.data
}

func (f *FloppyDriveController) Drq() byte {
	return f.drq
}

func (f *FloppyDriveController) LoadFloppy(drive byte) error {
	if drive < TotalDrives {
		return loadFloppy(&f.sectors[drive], f.config.FDC[drive].FloppyFile)
	}
	return errors.New("DriveNo " + strconv.Itoa(int(drive)) + " out of range")
}

func (f *FloppyDriveController) SaveFloppy(drive byte) error {
	if drive < TotalDrives {
		return saveFloppy(&f.sectors[drive], f.config.FDC[drive].FloppyFile)
	}
	return errors.New("DriveNo " + strconv.Itoa(int(drive)) + " out of range")
}

func NewFDC(conf *config.OkEmuConfig) *FloppyDriveController {
	sec := [2][SizeInSectors]SectorType{}
	// for each drive
	for d := 0; d < TotalDrives; d++ {
		// for each sector
		for i := 0; i < SizeInSectors; i++ {
			sec[d][i] = bytes.Repeat([]byte{0xe5}, SectorSize)
		}
	}

	return &FloppyDriveController{
		sideNo:  0,
		ddEn:    0,
		init:    0,
		drive:   0,
		mot1:    0,
		mot0:    0,
		intRq:   0,
		motSt:   0,
		drq:     0,
		lastCmd: 0xff,
		sectors: sec,
		bytePtr: 0xffff,
		//floppyConf: conf.FDC,
		config: conf,
	}
}

func (f *FloppyDriveController) dump() {
	log.Trace("Dump Buffer content.")
	file, err := os.Create("track-" + strconv.Itoa(int(f.trackNo)) + ".dat")
	if err != nil {
		log.Error(err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	err = binary.Write(file, binary.LittleEndian, f.trackBuffer)
	if err != nil {
		log.Error("Save track content failed:", err)
	}

}

func (f *FloppyDriveController) Track() byte {
	return f.trackNo
}

func (f *FloppyDriveController) Sector() byte {
	return f.sectorNo
}

// loadFloppy load floppy image to sector buffer from file
func loadFloppy(sectors *[SizeInSectors]SectorType, fileName string) error {
	log.Debugf("Load Floppy content from file %s.", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		log.Error(err)
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	for sector := 0; sector < SizeInSectors; sector++ {
		var n int
		n, err = file.Read(sectors[sector])
		if n != SectorSize {
			log.Error("Load floppy error, sector size: %d <> %d", n, SectorSize)
		}
		if err != nil {
			log.Error("Load floppy content failed:", err)
			return err
		}
	}
	return nil
}

// saveFloppy Save specified sectors to file with name fileName
func saveFloppy(sectors *[SizeInSectors]SectorType, fileName string) error {
	log.Debugf("Save Floppy to file %s.", fileName)
	file, err := os.Create(fileName)
	if err != nil {
		log.Error(err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	for sector := 0; sector < SizeInSectors; sector++ {
		var n int
		n, err = file.Write(sectors[sector])
		if n != SectorSize {
			log.Errorf("Save floppy error, sector %d size: %d <> %d", sector, n, SectorSize)
		}
		if err != nil {
			log.Error("Save floppy content failed:", err)
			return err
		}
	}
	return nil
}
