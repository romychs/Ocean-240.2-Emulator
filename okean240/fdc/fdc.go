package fdc

/**
 * Floppy drive controller, based on
 * MB8877, К1818ВГ93
 *
 * By Romych, 2025.03.05
 */

import (
	"encoding/binary"
	"os"
	"slices"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// Floppy parameters
const (
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
	CmdRestore         byte = 0x0
	CmdSeek            byte = 0x1
	CmdStep            byte = 0x2
	CmdStepIn          byte = 0x5
	CmdStepOut         byte = 0x7
	CmdReadSector      byte = 0x8
	CmdReadSectorMulti byte = 0x9
	CmdWriteSector     byte = 0xa
	CmdWriteTrack      byte = 0xf
	CmdNoCommand       byte = 0xff
)

const (
	StatusTR0        = 0x04 // TR0 - Head at track 0
	StatusRNF        = 0x10 // RNF - Record not found
	StatusSeekError  = 0x10 // Sector out of disk
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
	// FloppyStorage
	sectors [SizeInSectors]SectorType
	data    byte
	status  byte
	lastCmd byte
	//curSector   *SectorType
	bytePtr     uint16
	trackBuffer []byte
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
	SaveFloppy()
	GetSectorNo() uint16
	Track() byte
	Sector() byte
}

func (f *FloppyDriveController) GetSectorNo() uint16 {
	return uint16(f.sideNo)*SectorsPerSide + uint16(f.trackNo)*SectorPerTrack + uint16(f.sectorNo) - 1
}

func (f *FloppyDriveController) SetFloppy(val byte) {
	// WR: 5-SSEL, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
	f.sideNo = val >> 5 & 0x01
	f.ddEn = val >> 4 & 0x01
	f.init = val >> 3 & 0x01
	f.drive = val >> 2 & 0x01
	f.mot1 = val >> 1 & 0x01
	f.mot0 = val & 0x01
}

func (f *FloppyDriveController) GetFloppy() byte {
	// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT
	floppy := f.intRq | (f.mot0 << 1) | (f.mot1 << 2) | (f.drive << 3) | (f.sideNo << 6) | (f.motSt << 7)
	return floppy
}

func (f *FloppyDriveController) SetCmd(value byte) {
	f.lastCmd = value >> 4
	switch f.lastCmd {
	case CmdRestore:
		log.Debug("CMD Restore (seek trackNo 0)")
		f.trackNo = 0
		f.status = StatusTR0 | StatusHeadLoaded // TR0 & Head loaded
	case CmdSeek:
		log.Debugf("CMD Seek %x", value&0xf)
		f.status = StatusHeadLoaded
		f.trackNo = f.data
	case CmdStep:
		log.Debugf("CMD Step %x", value&0xf)
		f.status = StatusHeadLoaded
		f.trackNo = f.data
	case CmdStepIn:
		log.Debugf("CMD StepIn (Next track) %x", value&0xf)
		f.status = StatusHeadLoaded
		if f.trackNo < TracksCount {
			f.trackNo++
		}
	case CmdStepOut:
		log.Debugf("CMD StepOut (Previous track) %x", value&0xf)
		f.status = StatusHeadLoaded
		if f.trackNo > 0 {
			f.trackNo--
		}
	case CmdReadSector:
		sectorNo := f.GetSectorNo()
		log.Debugf("CMD Read single sectorNo: %d", sectorNo)
		if sectorNo < SizeInSectors {
			f.trackBuffer = slices.Clone(f.sectors[sectorNo])
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
			f.trackBuffer = slices.Concat(f.trackBuffer, f.sectors[sectorNo])
			sectorNo++
		}
		f.drq = 1
		f.status = 0x0
	case CmdWriteSector:
		sectorNo := f.GetSectorNo()
		log.Debugf("CMD Write Sector %d", sectorNo)
		if sectorNo < SizeInSectors {
			f.bytePtr = 0
			f.drq = 1
			f.status = 0x0
			f.trackBuffer = []byte{}
		} else {
			f.drq = 0
			f.status = StatusRNF
		}
	case CmdWriteTrack:
		log.Debugf("CMD Write Track %x", f.trackNo)
		f.status = 0x00
		f.trackBuffer = []byte{}
		f.drq = 1
	default:
		log.Debugf("Unknown CMD: %x VAL: %x", f.lastCmd, value&0xf)
	}
}

func (f *FloppyDriveController) Status() byte {
	return f.status
}

func (f *FloppyDriveController) SetTrackNo(value byte) {
	//log.Debugf("FDC Track: %d", value)
	if value > TracksCount {
		f.status |= 0x10 /// RNF
		log.Error("Track not found!")
	} else {
		f.trackNo = value
	}
}

func (f *FloppyDriveController) SetSectorNo(value byte) {
	//log.Debugf("FDC Sector: %d", value)
	if value > SectorPerTrack {
		f.status |= 0x10
		log.Error("Record not found!")
	} else {
		f.sectorNo = value
	}
}

func (f *FloppyDriveController) SetData(value byte) {
	//log.Debugf("FCD Data: %d", value)
	if f.lastCmd == CmdWriteTrack {
		if len(f.trackBuffer) < TrackBufferSize {
			f.trackBuffer = append(f.trackBuffer, value)
			f.drq = 1
			f.status = 0x00
		} else {
			//f.dump()
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
			f.sectors[f.GetSectorNo()] = slices.Clone(f.trackBuffer)
			f.lastCmd = CmdNoCommand
		}
	}
	f.data = value
}

func (f *FloppyDriveController) Data() byte {
	switch f.lastCmd {
	case CmdReadSector, CmdReadSectorMulti:
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

func (f *FloppyDriveController) LoadFloppy() {
	log.Debug("Load Floppy content.")
	file, err := os.Open("floppy.okd")
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

	for sector := 0; sector < SizeInSectors; sector++ {
		var n int
		n, err = file.Read(f.sectors[sector])
		if n != SectorSize {
			log.Error("Load floppy error, sector size: %d <> %d", n, SectorSize)
		}
		//		err = binary.Read(file, binary.LittleEndian, f.sectors[sector])
		if err != nil {
			log.Error("Load floppy content failed:", err)
			break
		}

	}

}

func (f *FloppyDriveController) SaveFloppy() {
	log.Debug("Save Floppy content.")
	file, err := os.Create("floppy.okd")
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

	// Write the struct to the file in little-endian byte order
	for sector := 0; sector < SizeInSectors; sector++ {
		var n int
		n, err = file.Write(f.sectors[sector])
		if n != SectorSize {
			log.Errorf("Save floppy error, sector %d size: %d <> %d", sector, n, SectorSize)
		}
		if err != nil {
			log.Error("Save floppy content failed:", err)
			break
		}
	}

}

func New() *FloppyDriveController {
	sec := [SizeInSectors]SectorType{}
	for i := 0; i < SizeInSectors; i++ {
		sec[i] = make([]byte, SectorSize)
		for s := 0; s < SectorSize; s++ {
			sec[i][s] = 0xE5
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
	}
}

func (f *FloppyDriveController) dump() {
	log.Debug("Dump Buffer content.")
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

//
