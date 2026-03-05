package okean240

import (
	"image/color"
	"okemu/config"
	fdc2 "okemu/okean240/fdc"
	"okemu/okean240/pit"
	"okemu/okean240/usart"
	"okemu/z80em"

	"fyne.io/fyne/v2"
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
	dd70          *pit.I8253
	dd72          *usart.I8251
	fdc           *fdc2.FloppyDriveController
	kbdBuffer     []byte
	vShift        byte
	hShift        byte
}

const VRAMBlock0 = 3
const VRAMBlock1 = 7
const VidVsuBit = 0x80
const VidColorBit = 0x40
const KbdBufferSize = 3

type ComputerInterface interface {
	Run()
	Reset()
	GetPixel(x uint16, y uint16) color.RGBA
	Do() uint64
	TimerClk()
	PutKey(key *fyne.KeyEvent)
	PutCtrlKey(shortcut fyne.Shortcut)
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

	c.vShift = 0
	c.hShift = 0

	c.dd70 = pit.NewI8253()
	c.dd72 = usart.NewI8251()
	c.fdc = fdc2.NewFDCType()

	return &c
}

func (c *ComputerType) Reset() {
	c.cpu.Reset()
	c.cycles = 0
	c.vShift = 0
	c.hShift = 0
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
		y += uint16(c.vShift)
		x += uint16(c.hShift)
		// Color 256x256 mode
		addr = ((x & 0xf8) << 6) | (y & 0xff)
		if c.vShift != 0 {
			addr -= 8
		}

		var cl byte = (c.vRAM.memory[addr&0x3fff] >> (x & 0x07)) & 1
		cl |= (c.vRAM.memory[(addr+0x100)&0x3fff] >> (x & 0x07)) & 1 << 1
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
		y += uint16(c.vShift)
		addr = ((x & 0xf8) << 5) | (y & 0xff)
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
