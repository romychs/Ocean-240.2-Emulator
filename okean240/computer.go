package okean240

import (
	_ "embed"
	"encoding/binary"
	"image/color"
	"okemu/config"
	"okemu/okean240/fdc"
	"okemu/okean240/pic"
	"okemu/okean240/pit"
	"okemu/okean240/usart"
	"okemu/z80em"
	"os"

	"fyne.io/fyne/v2"
	log "github.com/sirupsen/logrus"
)

type ComputerType struct {
	cpu           *z80em.Z80Type
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
	pit           *pit.I8253
	usart         *usart.I8251
	pic           *pic.I8259
	fdc           *fdc.FloppyDriveController
	kbdBuffer     []byte
	vShift        byte
	hShift        byte
	//aOffset       uint16
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
	PutKey(key *fyne.KeyEvent)
	PutRune(key rune)
	PutCtrlKey(shortcut fyne.Shortcut)
	SaveFloppy()
	LoadFloppy()
	Dump(start uint16, length uint16)
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

	c.cpu = z80em.New(&c)

	c.cycles = 0
	c.dd17EnableOut = false
	c.screenWidth = 512
	c.screenHeight = 256
	c.vRAM = c.memory.allMemory[3]
	c.palette = 0
	c.bgColor = 0

	c.vShift = 0
	c.hShift = 0
	//c.aOffset = 0x100

	c.pit = pit.New()
	c.usart = usart.New()
	c.pic = pic.New()
	c.fdc = fdc.New()

	return &c
}

func (c *ComputerType) Reset(cfg *config.OkEmuConfig) {
	c.cpu.Reset()
	c.cycles = 0
	c.vShift = 0
	c.hShift = 0

	c.memory = Memory{}
	c.memory.Init(cfg.MonitorFile, cfg.CPMFile)

	c.cycles = 0
	c.dd17EnableOut = false
	c.screenWidth = 512
	c.screenHeight = 256
	c.vRAM = c.memory.allMemory[3]

}

func (c *ComputerType) Do() int {
	//s := c.cpu.GetState()
	//if s.PC == 0xe0db {
	//	log.Debugf("breakpoint")
	//}
	ticks := uint64(c.cpu.RunInstruction())
	c.cycles += ticks
	//if c.cpu.PC == 0xFF26 {
	//	log.Debugf("%4X: H:%X L:%X A:%X B: %X C: %X D: %X E: %X", c.cpu.PC, c.cpu.H, c.cpu.L, c.cpu.A, c.cpu.B, c.cpu.C, c.cpu.D, c.cpu.E)
	//}
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

		var offset uint16
		if (c.vShift != 0) && (y > 255-uint16(c.vShift)) {
			offset = 0x100
		} else {
			offset = 0
		}
		y += uint16(c.vShift) & 0x00ff
		x += uint16(c.hShift-7) & 0x00ff

		// Color 256x256 mode
		addr = ((x & 0xf8) << 6) | y

		cl := (c.vRAM.memory[(addr-offset)&0x3fff] >> (x & 0x07)) & 1
		cl |= ((c.vRAM.memory[(addr+0x100-offset)&0x3fff] >> (x & 0x07)) & 1) << 1
		if cl == 0 {
			resColor = BgColorPalette[c.bgColor]
		} else {
			resColor = ColorPalette[c.palette][cl]
		}
	} else {
		if x > 511 {
			return CWhite
		}

		var offset uint16
		if (c.vShift != 0) && (y > 255-uint16(c.vShift)) {
			offset = 0x100
		} else {
			offset = 0
		}

		// Shifts
		y += uint16(c.vShift) & 0x00ff
		x += uint16(c.hShift-7) & 0x001ff

		// Mono 512x256 mode
		addr = (((x & 0x1f8) << 5) + y) - offset
		pix := c.vRAM.memory[addr&0x3fff] >> (x & 0x07) & 1
		if c.palette == 6 {
			if pix == 0 {
				resColor = CBlack
			} else {
				resColor = CLGreen
			}
		} else {
			if pix == 0 {
				resColor = BgColorPalette[c.bgColor]
			} else {
				resColor = MonoPalette[c.palette]
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
	c.pit.Tick(0)
	c.pit.Tick(1)

	// IRQ from timer
	if c.pit.Fired(0) {
		c.pic.SetIRQ(RstTimerNo)
		//c.ioPorts[PIC_DD75RS] |= Rst4Mask
	}
	// clock for SIO KR580VV51
	if c.pit.Fired(1) {
		c.usart.Tick()
	}
}

func (c *ComputerType) LoadFloppy() {
	c.fdc.LoadFloppy()
}

func (c *ComputerType) SaveFloppy() {
	c.fdc.SaveFloppy()
}

func (c *ComputerType) SetSerialBytes(bytes []byte) {
	c.usart.SetRxBytes(bytes)
}

func (c *ComputerType) SetRamBytes(bytes []byte) {
	addr := 0x100
	for i := 0; i < len(bytes); i++ {
		c.memory.MemWrite(uint16(addr), bytes[i])
		addr++
	}
	pages := len(bytes) / 256
	if len(bytes)%256 != 0 {
		pages++
	}
	log.Debugf("Loaded bytes: %d; blocks: %d", len(bytes), pages)
	//c.cpu.SP = 0x100
	//c.cpu.PC = 0x100
}

func (c *ComputerType) Dump(start uint16, length uint16) {
	file, err := os.Create("dump.dat")
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

	var buffer []byte
	for addr := 0; addr < 65535; addr++ {
		buffer = append(buffer, c.memory.MemRead(uint16(addr)))
	}
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		log.Error("Save memory dump failed:", err)
	}
}
