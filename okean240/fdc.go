package okean240

import log "github.com/sirupsen/logrus"

const FloppySizeK = 360
const SectorSize = 128
const SectorPerTrack = 36
const FloppySizeInS = FloppySizeK * 1024 / SectorSize
const TracksCount = FloppySizeInS / SectorPerTrack

type SectorType []byte

type FDCType struct {
	// Floppy controller port
	ssen   byte
	deenN  byte
	init   byte
	drsel  byte
	mot1   byte
	mot0   byte
	intr   byte
	motst  byte
	sector byte
	track  byte
	// FloppyStorage
	sectors [FloppySizeInS]SectorType
	data    byte
}

type FDCTypeInterface interface {
	SetFloppy()
	GetFloppy() byte
	SetCmd(value byte)
	SetTrack(value byte)
	SetSector(value byte)
	SetData(value byte)
	Data() byte
}

func (f FDCType) SetFloppy(val byte) {
	// WR: 5-SSEN, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
	f.ssen = val >> 5 & 0x01
	f.deenN = val >> 4 & 0x01
	f.init = val >> 3 & 0x01
	f.drsel = val >> 2 & 0x01
	f.mot1 = val >> 1 & 0x01
	f.mot0 = val & 0x01
}

func (f FDCType) GetFloppy() byte {
	// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT
	floppy := f.intr | (f.mot0 << 1) | (f.mot1 << 2) | (f.drsel << 3) | (f.ssen << 6) | (f.motst << 7)
	return floppy
}

func (f FDCType) SetCmd(value byte) {
	log.Debugf("FCD CMD: %x", value)
}

func (f FDCType) SetTrack(value byte) {
	log.Debugf("FCD Track: %d", value)
	f.track = value
}

func (f FDCType) SetSector(value byte) {
	log.Debugf("FCD Sector: %d", value)
	f.sector = value
}

func (f FDCType) SetData(value byte) {
	log.Debugf("FCD Data: %d", value)
	f.data = value
}

func (f FDCType) GetData() byte {
	return f.data
}

func NewFDCType() *FDCType {
	sec := [FloppySizeInS]SectorType{}
	for i := 0; i < FloppySizeInS; i++ {
		sec[i] = make(SectorType, SectorSize)
		for s := 0; s < 128; s++ {
			sec[i][s] = 0
		}
	}
	return &FDCType{
		ssen:    0,
		deenN:   0,
		init:    0,
		drsel:   0,
		mot1:    0,
		mot0:    0,
		intr:    0,
		motst:   0,
		sectors: sec,
	}
}

//
