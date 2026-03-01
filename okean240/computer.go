package okean240

import (
	"z80em/config"
	"z80em/z80em"

	log "github.com/sirupsen/logrus"
)

type ComputerType struct {
	cpu           z80em.Z80Type
	memory        Memory
	ioPorts       [256]byte
	cycles        uint64
	dd17EnableOut bool
}

type ComputerInterface interface {
	//Init(rom0 string, rom1 string)
	Run()
}

func (c *ComputerType) M1MemRead(addr uint16) byte {
	return c.memory.M1MemRead(addr)
}

func (c *ComputerType) MemRead(addr uint16) byte {
	return c.memory.MemRead(addr)
}

func (c *ComputerType) MemWrite(addr uint16, val byte) {
	c.memory.MemWrite(addr, val)
}

func (c *ComputerType) IORead(port uint16) byte {
	return c.ioPorts[port]
}

func (c *ComputerType) IOWrite(port uint16, val byte) {
	c.ioPorts[byte(port&0x00ff)] = val
	switch byte(port & 0x00ff) {
	case SYS_DD17PB:

		if c.dd17EnableOut {
			c.memory.Configure(val)
		}
	case SYS_DD17CTR:
		c.dd17EnableOut = val == 0x80
	case KBD_DD78CTR:

	}
}

// New Builds new computer
func New(cfg config.OkEmuConfig) *ComputerType {
	c := ComputerType{}
	c.memory = Memory{}
	c.memory.Init(cfg.MonitorFile, cfg.CPMFile)

	c.cpu = *z80em.New(&c)

	c.cycles = 0
	c.dd17EnableOut = false
	return &c
}

func (c *ComputerType) Run() {
	c.cpu.Reset()
	for {
		state := c.cpu.GetState()
		log.Infof("%d - [%x]: %x\n", c.cycles, state.PC, c.MemRead(state.PC))
		c.cycles += uint64(c.cpu.RunInstruction())

	}
}
