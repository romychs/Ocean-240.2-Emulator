package okean240

import (
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"okemu/config"
	"okemu/debug"
	"okemu/okean240/fdc"
	"okemu/okean240/pic"
	"okemu/okean240/pit"
	"okemu/okean240/usart"
	"okemu/z80"
	"os"

	//"okemu/z80"
	"okemu/z80/c99"
	//"okemu/z80/js"

	"fyne.io/fyne/v2"
	log "github.com/sirupsen/logrus"
)

const DefaultCPUFrequency = 2_500_000

type ComputerType struct {
	cpu            *c99.Z80
	memory         Memory
	ioPorts        [256]byte
	cycles         uint64
	tstatesPartial uint64
	dd17EnableOut  bool
	colorMode      bool
	screenWidth    int
	screenHeight   int
	vRAM           *RamBlock
	palette        byte
	bgColor        byte
	pit            *pit.I8253
	usart          *usart.I8251
	pic            *pic.I8259
	fdc            *fdc.FloppyDriveController
	kbdBuffer      []byte
	vShift         byte
	hShift         byte
	cpuFrequency   uint32
	//
	debugger *debug.Debugger
}

type Snapshot struct {
	CPU    *z80.CPU `json:"cpu,omitempty"`
	Memory string   `json:"memory,omitempty"`
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
	CPUState() *z80.CPU
	SetCPUState(state *z80.CPU)
}

func (c *ComputerType) GetCPUState() *z80.CPU {
	return c.cpu.GetState()
}

func (c *ComputerType) SetCPUState(state *z80.CPU) {
	c.cpu.SetState(state)
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

// NewComputer Builds new computer
func NewComputer(cfg *config.OkEmuConfig, deb *debug.Debugger) *ComputerType {
	c := ComputerType{}
	c.memory = Memory{}
	c.memory.Init(cfg.MonitorFile, cfg.CPMFile)

	c.cpu = c99.New(&c)

	c.cycles = 0
	c.tstatesPartial = 0
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
	c.fdc = fdc.NewFDC(cfg)
	c.cpuFrequency = DefaultCPUFrequency
	c.debugger = deb
	return &c
}

func (c *ComputerType) Reset() {
	c.cpu.Reset()
	c.cycles = 0
	c.tstatesPartial = 0
}

func (c *ComputerType) getContext() map[string]interface{} {
	context := make(map[string]interface{})
	s := c.cpu.GetState()
	context["A"] = s.A
	context["B"] = s.B
	context["C"] = s.C
	context["D"] = s.D
	context["E"] = s.E
	context["H"] = s.H
	context["L"] = s.L
	context["A'"] = s.AAlt
	context["B'"] = s.BAlt
	context["C'"] = s.CAlt
	context["D'"] = s.DAlt
	context["E'"] = s.EAlt
	context["H'"] = s.HAlt
	context["L'"] = s.LAlt
	context["PC"] = s.PC
	context["SP"] = s.SP
	context["IX"] = s.IX
	context["IY"] = s.IY
	context["ZF"] = s.Flags.Z
	context["SF"] = s.Flags.S
	context["NF"] = s.Flags.N
	context["PF"] = s.Flags.P
	context["HF"] = s.Flags.H
	context["YF"] = s.Flags.Y
	context["XF"] = s.Flags.X
	context["CF"] = s.Flags.C
	context["BC"] = uint16(s.B)<<8 | uint16(s.C)
	context["DE"] = uint16(s.D)<<8 | uint16(s.E)
	context["HL"] = uint16(s.H)<<8 | uint16(s.L)
	context["AF"] = uint16(s.A)<<8 | uint16(s.Flags.GetFlags())
	return context
}

func (c *ComputerType) Do() (uint32, uint16, byte) {
	ticks := uint32(0)
	var memAccess *map[uint16]byte
	if c.debugger.StepMode() {
		if c.debugger.RunMode() || c.debugger.DoStep() {
			if c.debugger.RunInst() > 0 {
				// skip first instruction after run-mode activated
				bpHit, bp := c.debugger.CheckBreakpoints(c.getContext())
				if bpHit {
					//c.debugger.SetRunMode(false)
					return 0, bp, 0
				}
			}
			c.debugger.SaveHistory(c.cpu.GetState())
			ticks, memAccess = c.cpu.RunInstruction()
			mHit, mAddr, mTyp := c.debugger.CheckMemBreakpoints(memAccess)
			if mHit {
				return ticks, mAddr, mTyp
			}
		}
	} else {
		ticks, memAccess = c.cpu.RunInstruction()
	}
	c.cycles += uint64(ticks)
	c.tstatesPartial += uint64(ticks)
	return ticks, 0, 0
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

func (c *ComputerType) ResetTStatesPartial() {
	c.tstatesPartial = 0
}

func (c *ComputerType) TStatesPartial() uint64 {
	return c.tstatesPartial
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

func (c *ComputerType) LoadFloppy(drive byte) error {
	return c.fdc.LoadFloppy(drive)
}

func (c *ComputerType) SaveFloppy(drive byte) error {
	return c.fdc.SaveFloppy(drive)
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
	for addr := start; addr < start+length; addr++ {
		buffer = append(buffer, c.memory.MemRead(addr))
	}
	_, err = file.Write(buffer)
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		log.Error("Save memory dump failed:", err)
	} else {
		log.Debug("Memory dump saved successfully")
	}
}

func (c *ComputerType) CPUFrequency() uint32 {
	return c.cpuFrequency
}

func (c *ComputerType) SetCPUFrequency(frequency uint32) {
	c.cpuFrequency = frequency
}

func (c *ComputerType) DebuggerState() string {
	if c.debugger.StepMode() {
		if c.debugger.RunMode() {
			return "Run"
		}
		return "Step"
	}
	return "Off"
}

func (c *ComputerType) memoryAsHexStr() string {
	res := ""
	for addr := 0; addr <= 65535; addr++ {
		res += fmt.Sprintf("%02X", c.memory.MemRead(uint16(addr)))
	}
	return res
}

func (c *ComputerType) SaveSnapshot(fn string) error {
	// create snapshot file
	file, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	// take snapshot
	s := Snapshot{
		CPU:    c.cpu.GetState(),
		Memory: c.memoryAsHexStr(),
	}
	// convert to JSON
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	// and save
	err = binary.Write(file, binary.LittleEndian, b)
	if err != nil {
		return err
	}
	return nil
}
