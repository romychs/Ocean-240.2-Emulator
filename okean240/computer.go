package okean240

import (
	"image/color"
	"okemu/config"
	"okemu/z80em"

	log "github.com/sirupsen/logrus"
)

type ComputerType struct {
	cpu           z80em.Z80Type
	memory        Memory
	ioPorts       [256]byte
	cycles        uint64
	dd17EnableOut bool
	colorMode     bool
	screenWidth   int
	screenHeight  int
	vRAM          *RamBlock
	palette       byte
	bgColor       byte
	dd70          *Timer8253
	dd72          *Sio8251
}

const VRAMBlock0 = 3
const VRAMBlock1 = 7
const VidVsuBit = 0x80
const VidColorBit = 0x40

type ComputerInterface interface {
	Run()
	Reset()
	GetPixel(x uint16, y uint16) color.RGBA
	Do() uint64
	TimerClk()
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
	switch port & 0x00ff {
	case PIC_DD75RS:
		v := c.ioPorts[PIC_DD75RS]
		c.ioPorts[PIC_DD75RS] = 0
		return v
	default:
		log.Debugf("IORead from port: %x", port)
	}
	return c.ioPorts[byte(port&0x00ff)]
}

func (c *ComputerType) IOWrite(port uint16, val byte) {
	bp := byte(port & 0x00ff)
	c.ioPorts[bp] = val
	//log.Debugf("OUT (%x), %x", bp, val)
	switch bp {
	case SYS_DD17PB:
		if c.dd17EnableOut {
			c.memory.Configure(val)
		}
	case SYS_DD17CTR:
		c.dd17EnableOut = val == 0x80
	case VID_DD67PB:
		if val&VidVsuBit == 0 {
			// video page 0
			c.vRAM = c.memory.allMemory[VRAMBlock0]
		} else {
			// video page 1
			c.vRAM = c.memory.allMemory[VRAMBlock1]
		}
		if val&VidColorBit != 0 {
			c.colorMode = true
			c.screenWidth = 256
		} else {
			c.colorMode = false
			c.screenWidth = 512
		}
		c.palette = val & 0x07
		c.bgColor = val & 0x38 >> 3
	case DD67CTR:

	case TMR_DD70CTR:
		// Timer VI63 config register
		c.dd70.Configure(val)
	case TMR_DD70C1:
		// Timer VI63 counter0 register
		c.dd70.Load(0, val)
	case TMR_DD70C2:
		// Timer VI63 counter1 register
		c.dd70.Load(1, val)
	case TMR_DD70C3:
		// Timer VI63 counter2 register
		c.dd70.Load(2, val)

	case KBD_DD78CTR:
	default:
		//log.Debugf("OUT to Unknown port (%x), %x", bp, val)

	}
}

// New Builds new computer
func New(cfg *config.OkEmuConfig) *ComputerType {
	c := ComputerType{}
	c.memory = Memory{}
	c.memory.Init(cfg.MonitorFile, cfg.CPMFile)

	c.cpu = *z80em.New(&c)

	c.cycles = 0
	c.dd17EnableOut = false
	c.screenWidth = 512
	c.screenHeight = 256
	c.vRAM = c.memory.allMemory[3]
	c.palette = 0
	c.bgColor = 0

	c.dd70 = NewTimer8253()
	c.dd72 = NewSio8251()

	return &c
}

func (c *ComputerType) Reset() {
	c.cpu.Reset()
	c.cycles = 0
}

func (c *ComputerType) Do() int {
	s := c.cpu.GetState()
	if s.PC == 0xe0db {
		log.Debugf("breakpoint")
	}
	ticks := uint64(c.cpu.RunInstruction())
	c.cycles += ticks
	return int(ticks)
}

func (c *ComputerType) GetPixel(x uint16, y uint16) color.RGBA {
	if y > 255 {
		return CWhite
	}

	var addr uint16
	var resColor color.RGBA

	if c.colorMode {
		if x > 255 {
			return CWhite
		}
		// Color 256x256 mode
		addr = ((x & 0xf8) << 6) | y
		var mask byte = 1 << (x & 0x07)
		pix1 := c.vRAM.memory[addr]&(mask) != 0
		pix2 := c.vRAM.memory[addr+0x100]&(mask) != 0
		var cl byte = 0
		if pix1 {
			cl |= 1
		}
		if pix2 {
			cl |= 2
		}
		if cl == 0 {
			resColor = BgColorPalette[c.bgColor]
		} else {
			resColor = ColorPalette[c.palette][cl]
		}
	} else {
		if x > 511 {
			return CWhite
		}
		// Mono 512x256 mode
		addr = ((x & 0xf8) << 5) | y
		pix := c.vRAM.memory[addr]&(1<<x) != 0
		if c.palette == 6 {
			if !pix {
				resColor = CBlack
			} else {
				resColor = CLGreen
			}
		} else {
			if !pix {
				resColor = BgColorPalette[c.bgColor]
			} else {
				resColor = MonoPalette[c.bgColor]
			}
		}
	}
	return resColor
}

func (c *ComputerType) ScreenWidth() int {
	return c.screenWidth
}

func (c *ComputerType) ScreenHeight() int {
	return c.screenHeight
}

func (c *ComputerType) Cycles() uint64 {
	return c.cycles
}

func (c *ComputerType) TimerClk() {
	// DD70 KR580VI53 CLK0, CKL1 @ 1.5MHz
	c.dd70.Tick(0)
	c.dd70.Tick(1)

	// IRQ from timer
	if c.dd70.Fired(0) {
		c.ioPorts[PIC_DD75RS] |= Rst4TmrFlag
	}
	// clock for SIO KR580VV51
	if c.dd70.Fired(1) {
		c.dd72.Tick()
	}
}
