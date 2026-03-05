package fdc

/**
Floppy drive controller, based on
MB8877, К1818ВГ93

By Romych, 2025.03.05
*/

import (
	log "github.com/sirupsen/logrus"
)

const FloppySizeK = 360

const SectorSize = 128
const SideCount = 2
const SectorPerTrack = 36

const SizeInSectors = FloppySizeK * 1024 / SectorSize
const TracksCount = SizeInSectors / SideCount / SectorPerTrack
const SectorsPerSide = SizeInSectors / SideCount

type SectorType []byte

type FloppyDriveController struct {
	// Floppy controller port
	sideSel byte
	ddEn    byte
	init    byte
	drSel   byte
	mot1    byte
	mot0    byte
	intRq   byte
	motSt   byte
	sector  byte
	track   byte
	drq     byte
	// FloppyStorage
	sectors   [SizeInSectors]SectorType
	data      byte
	status    byte
	lastCmd   byte
	curSector *SectorType
	bytePtr   uint16
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
}

func (f *FloppyDriveController) SetFloppy(val byte) {
	// WR: 5-SSEN, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
	f.sideSel = val >> 5 & 0x01
	f.ddEn = val >> 4 & 0x01
	f.init = val >> 3 & 0x01
	f.drSel = val >> 2 & 0x01
	f.mot1 = val >> 1 & 0x01
	f.mot0 = val & 0x01
}

func (f *FloppyDriveController) GetFloppy() byte {
	// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT
	floppy := f.intRq | (f.mot0 << 1) | (f.mot1 << 2) | (f.drSel << 3) | (f.sideSel << 6) | (f.motSt << 7)
	return floppy
}

const (
	FdcCmdRestore    byte = 0
	FdcCmdSeek       byte = 1
	FdcCmdStep       byte = 2
	FdcCmdReadSector byte = 8
)

func (f *FloppyDriveController) SetCmd(value byte) {
	//log.Debugf("FCD CMD: %x", value)
	f.lastCmd = value >> 4
	switch f.lastCmd {
	case FdcCmdRestore:
		log.Debug("CMD Restore (seek track 0)")
		f.status = 0x24 // TR0 & Head loaded
		f.track = 0
	case FdcCmdSeek:
		log.Debugf("CMD Seek %x", value&0xf)
		f.status = 0x04 // Head loaded
		f.track = f.data
	case FdcCmdStep:
		log.Debugf("CMD Step %x", value&0xf)
		f.status = 0x04 // Head loaded
		f.track = f.data
	case FdcCmdReadSector:
		f.status = 0x04
		sectorNo := uint16(f.sideSel)*SectorsPerSide + uint16(f.track)*SectorPerTrack + uint16(f.sector)
		log.Debugf("CMD Read single sector: %d", sectorNo)
		if sectorNo >= SizeInSectors {
			f.status = 0x10 // RNF - Record not found
		} else {
			f.curSector = &f.sectors[sectorNo]
			f.bytePtr = 0
			f.drq = 1
			f.status = 0x00
		}
	default:
		log.Debugf("Unknown CMD: %x VAL: %x", f.lastCmd, value&0xf)
	}
}

func (f *FloppyDriveController) Status() byte {
	return f.status
}

func (f *FloppyDriveController) SetTrack(value byte) {
	log.Debugf("FCD Track: %d", value)
	f.track = value
}

func (f *FloppyDriveController) SetSector(value byte) {
	log.Debugf("FCD Sector: %d", value)
	f.sector = value
}

func (f *FloppyDriveController) SetData(value byte) {
	log.Debugf("FCD Data: %d", value)
	f.data = value
}

func (f *FloppyDriveController) Data() byte {
	if f.lastCmd == FdcCmdReadSector {
		if f.bytePtr < SectorSize {
			f.drq = 1
			f.data = (*f.curSector)[f.bytePtr]
			f.bytePtr++
		} else {
			f.drq = 0
			f.status = 0
		}
	}
	return f.data
}

func (f *FloppyDriveController) Drq() byte {
	return f.drq
}

func NewFDCType() *FloppyDriveController {
	sec := [SizeInSectors]SectorType{}
	for i := 0; i < int(SizeInSectors); i++ {
		sec[i] = make([]byte, SectorSize)
		for s := 0; s < 128; s++ {
			sec[i][s] = 0
		}
	}
	return &FloppyDriveController{
		sideSel: 0,
		ddEn:    0,
		init:    0,
		drSel:   0,
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

//
