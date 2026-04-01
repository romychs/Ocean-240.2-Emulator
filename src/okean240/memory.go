package okean240

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const RamBlockSize = 16 * 1024           // 16Kb
const RamSize = 512 * 1024               // 512kb (16xRU7) or 128k (16xRU5)
const RamBlocks = RamSize / RamBlockSize // 32 Ram blocks for 512k, 8 for 128k
const RamDefaultInitPattern = 0x3f
const RamWindows = 4

const RSTBit = 0x20
const ROMDisBit = 0x10
const AccessHiBit = 0x01
const ExtRamAddrBits = 0x0E

const (
	WindowNo0 = iota
	WindowNo1
	WindowNo2
	WindowNo3
)

type MemoryBlock struct {
	id     byte
	memory [RamBlockSize]byte
}

type Memory struct {
	allMemory    [RamBlocks]*MemoryBlock
	memoryWindow [RamWindows]*MemoryBlock
	rom0         MemoryBlock // monitor + monitor
	rom1         MemoryBlock // cpm + monitor
	config       byte
}

type MemoryInterface interface {
	// Init - Initialize memory at "computer started"
	Init(rom0 string, rom1 string)
	// Configure - Set memory configuration
	Configure(value byte)
	// M1MemRead Read byte from memoryWindow for specified address
	M1MemRead(addr uint16) byte
	// MemRead Read byte from memoryWindow for specified address
	MemRead(addr uint16) byte
	// MemWrite Write byte to memoryWindow to specified address
	MemWrite(addr uint16, val byte)
}

func (m *Memory) Init(monFile string, cmpFile string) {

	// empty RAM
	var id byte = 0
	for block := range m.allMemory {
		rb := MemoryBlock{}
		rb.id = id
		id++
		for addr := 0; addr < RamBlockSize; addr++ {
			rb.memory[addr] = RamDefaultInitPattern
		}
		m.allMemory[block] = &rb
	}

	// Command ROM files and init ROM0,1
	// Read the entire file into a byte slice
	rom0bin, err := os.ReadFile(monFile)
	if err != nil {
		log.Fatal(err)
	}
	rom1bin, err := os.ReadFile(cmpFile)
	if err != nil {
		log.Fatal(err)
	}
	m.rom0 = MemoryBlock{}
	m.rom0.id = 0xF0
	m.rom1 = MemoryBlock{}
	m.rom1.id = 0xF1
	half := RamBlockSize / 2
	for i := 0; i < half; i++ {
		// mon+mon
		m.rom0.memory[i] = rom0bin[i]
		m.rom0.memory[i+half] = rom0bin[i]
		// cp/m + mon
		m.rom1.memory[i] = rom1bin[i]
		m.rom1.memory[i+half] = rom0bin[i]
	}
	// Config mem with RST pin Hi
	m.Configure(RSTBit)
}

// Configure - Configure memoryWindow windows
func (m *Memory) Configure(value byte) {
	m.config = value
	if m.config&RSTBit != 0 {
		// RST bit set just after System RESET
		// All memoryWindow windows points to ROM0 (monitor)
		for i := 0; i < RamWindows; i++ {
			m.memoryWindow[i] = &m.rom0
		}
	} else {
		// Map RAM blocks to windows
		sp := (m.config & ExtRamAddrBits) << 1 // 0,4,8,12
		for i := byte(0); i < RamWindows; i++ {
			m.memoryWindow[i] = m.allMemory[sp+i]
		}
		// Map two hi windows to low windows in 32k flag set
		if m.config&AccessHiBit == 1 {
			m.memoryWindow[WindowNo0] = m.memoryWindow[WindowNo2]
			m.memoryWindow[WindowNo1] = m.memoryWindow[WindowNo3]
		}
		// If ROM enabled, map ROM to last window
		if m.config&ROMDisBit == 0 {
			// If ROM enabled, CP/M + Mon at window 3 [0xC000:0xFFFF]
			m.memoryWindow[WindowNo3] = &m.rom1
		}
	}
}

func (m *Memory) M1MemRead(addr uint16) byte {
	return m.memoryWindow[addr>>14].memory[addr&0x3fff]
}

func (m *Memory) MemRead(addr uint16) byte {
	return m.memoryWindow[addr>>14].memory[addr&0x3fff]
}

func (m *Memory) MemWrite(addr uint16, val byte) {
	window := addr >> 14
	offset := addr & 0x3fff
	if m.memoryWindow[window].id < 0xF0 {
		// write to RAM only
		m.memoryWindow[window].memory[offset] = val
	} else {
		log.Debugf("Attempting to write 0x%02X=>ROM[0x%04X]=", val, addr)
	}
}

// MemoryWindows Return memory pages, mapped to memory windows
func (m *Memory) MemoryWindows() []byte {
	var res []byte
	for w := 0; w < RamWindows; w++ {
		res = append(res, m.memoryWindow[w].id)
	}
	return res
}
