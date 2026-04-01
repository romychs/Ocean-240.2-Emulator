package okean240

import (
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/breakpoint"
	"okemu/gval"
	"okemu/okean240/fdc"
	"okemu/okean240/pic"
	"okemu/okean240/pit"
	"okemu/okean240/usart"
	"okemu/z80"
	"os"
	"strconv"
	"sync/atomic"

	"okemu/z80/c99"

	log "github.com/sirupsen/logrus"
)

const DefaultCPUFrequency = 2_500_000
const CPUFrequencyLow float64 = 2.47
const CPUFrequencyHi float64 = 2.53
const TimerFrequencyLow float64 = 1.47
const TimerFrequencyHi float64 = 1.53

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
	vRAM           *MemoryBlock
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
	debugger     *debug.Debugger
	config       *config.OkEmuConfig
	kbAck        atomic.Bool
	fullSpeed    atomic.Bool
	pendingReset atomic.Bool
}

type Snapshot struct {
	CPU    *z80.CPU `json:"cpu,omitempty"`
	Memory string   `json:"memory,omitempty"`
}

const VRAMBlock0 = 3
const VRAMBlock1 = 7
const VidVsuBit = 0x80
const VidColorBit = 0x40

//type ComputerInterface interface {
//	Run()
//	Reset()
//	GetPixel(x uint16, y uint16) color.RGBA
//	Do() uint64
//	TimerClk()
//	PutKey(key *fyne.KeyEvent)
//	PutRune(key rune)
//	PutCtrlKey(shortcut fyne.Shortcut)
//	SaveFloppy()
//	LoadFloppy()
//	CPUState() *z80.CPU
//	SetCPUState(state *z80.CPU)
//}

func (c *ComputerType) CPUState() *z80.CPU {
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
	c.config = cfg
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
	c.kbAck.Store(false)
	c.usart = usart.New()
	c.pic = pic.NewI8259()
	c.fdc = fdc.NewFDC(cfg)
	c.cpuFrequency = DefaultCPUFrequency
	c.debugger = deb
	c.fullSpeed.Store(false)
	c.pendingReset.Store(false)
	return &c
}

func (c *ComputerType) Reset() {
	c.cpu.Reset()
	c.cycles = 0
	c.tstatesPartial = 0
}

func (c *ComputerType) getContext() map[string]interface{} {
	ctx := make(map[string]interface{})
	s := c.cpu.GetState()
	ctx["A"] = s.A
	ctx["B"] = s.B
	ctx["C"] = s.C
	ctx["D"] = s.D
	ctx["E"] = s.E
	ctx["H"] = s.H
	ctx["L"] = s.L
	ctx["A'"] = s.AAlt
	ctx["B'"] = s.BAlt
	ctx["C'"] = s.CAlt
	ctx["D'"] = s.DAlt
	ctx["E'"] = s.EAlt
	ctx["H'"] = s.HAlt
	ctx["L'"] = s.LAlt
	ctx["PC"] = s.PC
	ctx["SP"] = s.SP
	ctx["IX"] = s.IX
	ctx["IY"] = s.IY
	ctx["ZF"] = s.Flags.Z
	ctx["SF"] = s.Flags.S
	ctx["NF"] = s.Flags.N
	ctx["PF"] = s.Flags.P
	ctx["HF"] = s.Flags.H
	ctx["YF"] = s.Flags.Y
	ctx["XF"] = s.Flags.X
	ctx["CF"] = s.Flags.C
	ctx["BC"] = uint16(s.B)<<8 | uint16(s.C)
	ctx["DE"] = uint16(s.D)<<8 | uint16(s.E)
	ctx["HL"] = uint16(s.H)<<8 | uint16(s.L)
	ctx["AF"] = uint16(s.A)<<8 | uint16(s.Flags.GetFlags())
	ctx["MEM"] = c.memory.MemRead
	ctx["IO"] = c.IORead
	return ctx
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

		a1 := (addr - offset) & 0x3fff
		a2 := (a1 + 0x100) & 0x3fff

		cl := (c.vRAM.memory[a1] >> (x & 0x07)) & 1
		cl |= ((c.vRAM.memory[a2] >> (x & 0x07)) & 1) << 1
		if cl == 0 {
			//resColor = BgColorPalette[c.bgColor]
			resColor = ColorPalette[c.palette][cl]
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
	err = os.WriteFile(fn, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *ComputerType) LoadSnapshot(fn string) error {
	// read snapshot file
	b, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	// unmarshal from JSON
	var result Snapshot
	err = json.Unmarshal(b, &result)
	if err != nil {
		return err
	}
	c.cpu.SetState(result.CPU)
	return c.restoreMemoryFromHex(result.Memory)
}

func (c *ComputerType) restoreMemoryFromHex(memory string) error {
	for addr := 0; addr <= 65535; addr++ {
		b, e := strconv.ParseUint(memory[addr*2:addr*2+2], 16, 8)
		if e != nil {
			log.Error(e)
			return e
		}
		c.memory.MemWrite(uint16(addr), byte(b))
	}
	return nil
}

func (c *ComputerType) AutoSaveFloppy() {
	for drv := byte(0); drv < fdc.TotalDrives; drv++ {
		if c.config.FDC[drv].AutoSave {
			e := c.fdc.SaveFloppy(drv)
			if e != nil {
				log.Error(e)
			}
		}
	}
}

func (c *ComputerType) AutoLoadFloppy() {
	for drv := byte(0); drv < fdc.TotalDrives; drv++ {
		if c.config.FDC[drv].AutoLoad {
			e := c.fdc.LoadFloppy(drv)
			if e != nil {
				log.Error(e)
			}
		}
	}
}

func (c *ComputerType) ClearCodeCoverage() {
	c.cpu.ClearCodeCoverage()
}

func (c *ComputerType) SetCodeCoverage(enabled bool) {
	c.cpu.SetCodeCoverage(enabled)
}

func (c *ComputerType) CodeCoverage() map[uint16]bool {
	return c.cpu.CodeCoverage()
}

func (c *ComputerType) SetExtendedStack(enabled bool) {
	c.cpu.SetExtendedStack(enabled)
}

func (c *ComputerType) ExtendedStack() ([]byte, error) {
	return c.cpu.ExtendedStack()
}

var language gval.Language

func init() {
	language = gval.NewLanguage(gval.Base(), gval.Arithmetic(), gval.Bitmask(), gval.PropositionalLogic(),
		gval.Function("PEEKW", breakpoint.CfPeekW),
		gval.Function("PEEK", breakpoint.CfPeek),
		gval.Function("BYTE", breakpoint.CfByte),
		gval.Function("WORD", breakpoint.CfWord),
		gval.Function("ABS", breakpoint.CfAbs),
		gval.Function("IN", breakpoint.CfIn),
	)
}

func (c *ComputerType) Evaluate(expression string) (string, error) {
	params := c.getContext()
	bc := context.WithValue(context.Background(), "MEM", params["MEM"])
	bc = context.WithValue(bc, "IO", params["IO"])
	eval, err := language.NewEvaluable(expression)
	if err != nil {
		return "", fmt.Errorf("error: %s", err.Error())
	}
	value, err := eval.EvalUint(bc, params)
	if err != nil {
		return "", fmt.Errorf("error: %s", err.Error())
	}
	return strconv.FormatUint(uint64(value), 10), nil
}

func (c *ComputerType) MemoryPages() []byte {
	return c.memory.MemoryWindows()
}

func (c *ComputerType) SetFullSpeed(full bool) {
	c.fullSpeed.Store(full)
}

func (c *ComputerType) FullSpeed() bool {
	return c.fullSpeed.Load()
}

func (c *ComputerType) SetPendingReset(pending bool) {
	c.pendingReset.Store(pending)
}

func (c *ComputerType) PendingReset() bool {
	return c.pendingReset.Load()
}
