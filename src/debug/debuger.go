package debug

import (
	"fmt"
	"okemu/debug/breakpoint"
	"okemu/z80"
	"okemu/z80/dis"

	log "github.com/sirupsen/logrus"
)

const BPMemAccess = 65535

type Debugger struct {
	stepMode           bool
	doStep             bool
	runMode            bool
	runInst            uint64
	breakpointsEnabled bool
	breakpoints        map[uint16]*breakpoint.Breakpoint
	cpuFrequency       uint32
	disassembler       *dis.Disassembler
	cpuHistoryEnabled  bool
	cpuHistoryStarted  bool
	cpuHistoryMaxSize  int
	cpuHistory         []*z80.CPU
	memBreakpoints     [65536]byte
}

func NewDebugger() *Debugger {
	d := Debugger{
		stepMode:           false,
		doStep:             false,
		runMode:            false,
		runInst:            0,
		breakpointsEnabled: false,
		breakpoints:        map[uint16]*breakpoint.Breakpoint{},
		cpuHistoryEnabled:  false,
		cpuHistoryStarted:  false,
		cpuHistoryMaxSize:  0,
		cpuHistory:         []*z80.CPU{},
	}
	return &d
}

type DEZOG interface {
	SetupTcpHandler()
	BreakpointHit(number uint16, typ byte)
}

func (d *Debugger) SetStepMode(step bool) {
	d.SetRunMode(false)
	d.stepMode = step
}

func (d *Debugger) SetRunMode(run bool) {
	if run {
		d.runInst = 0
	}
	d.runMode = run
}

func (d *Debugger) RunMode() bool {
	return d.runMode
}

func (d *Debugger) DoStep() bool {
	if d.doStep {
		d.doStep = false
		return true
	}
	return false
}

func (d *Debugger) SetCpuHistoryEnabled(enable bool) {
	d.cpuHistoryEnabled = enable
}

func (d *Debugger) SetCpuHistoryMaxSize(size int) {
	if size < 0 || size > 1_000_000 {
		log.Error("CPU history max size must be positive and up to 1M")
	} else {
		d.cpuHistoryMaxSize = size
	}
}

func (d *Debugger) CpuHistoryClear() {
	d.cpuHistory = make([]*z80.CPU, 0)
}

func (d *Debugger) CpuHistorySize() int {
	return len(d.cpuHistory)
}

func (d *Debugger) CpuHistory(index int) *z80.CPU {
	if index >= 0 && index < len(d.cpuHistory) {
		return d.cpuHistory[index]
	}
	if len(d.cpuHistory) > 0 {
		log.Warnf("CPU history index %d out of range [0:%d]", index, len(d.cpuHistory)-1)
	} else {
		log.Warn("CPU history is empty")
	}
	return nil
}

func (d *Debugger) SetCpuHistoryStarted(started bool) {
	d.cpuHistoryStarted = started
}

func (d *Debugger) SaveHistory(state *z80.CPU) {
	if d.cpuHistoryEnabled && d.cpuHistoryMaxSize > 0 && d.cpuHistoryStarted {
		d.cpuHistory = append([]*z80.CPU{state}, d.cpuHistory...)
		if len(d.cpuHistory) > d.cpuHistoryMaxSize {
			d.cpuHistory = d.cpuHistory[0 : d.cpuHistoryMaxSize-1]
		}
	}
}

func (d *Debugger) CheckBreakpoints(ctx map[string]interface{}) (bool, uint16) {
	if d.breakpointsEnabled && d.runMode {
		for n, bp := range d.breakpoints {
			if bp != nil && bp.Hit(ctx) {
				// breakpoint hit
				if bp.Pass() >= bp.PassCount() {
					bp.SetPass(0)
					d.runMode = false
					return true, n
				}
				// increment breakpoint pass count
				bp.IncPass()
			}
		}
	}
	return false, 0
}

func (d *Debugger) SetBreakpointsEnabled(enabled bool) {
	d.breakpointsEnabled = enabled
}

func (d *Debugger) BreakpointsEnabled() bool {
	return d.breakpointsEnabled
}

// SetBreakpoint  Create new breakpoint with specified number
func (d *Debugger) SetBreakpoint(number uint16, exp string, mBank uint8) error {
	var err error
	bp, err := breakpoint.NewBreakpoint(exp, mBank)
	if err == nil && bp != nil {
		d.breakpoints[number] = bp
	}
	return err
}

func (d *Debugger) AddBreakpoint(exp string, mBank uint8) (uint16, error) {
	var err error
	bpNo := d.GetBreakpointNum()
	if bpNo < breakpoint.MaxBreakpoints {
		bp, err := breakpoint.NewBreakpoint(exp, mBank)
		if err == nil && bp != nil {
			bp.SetEnabled(true)
			d.breakpoints[bpNo] = bp
		}
	}
	return bpNo, err
}

func (d *Debugger) GetBreakpointNum() uint16 {
	num := uint16(1)
	for no, bp := range d.breakpoints {
		if bp != nil && no < breakpoint.MaxBreakpoints && num <= no {
			num = no + 1
		}
	}
	return num
}

func (d *Debugger) SetBreakpointPassCount(number uint16, count uint16) {
	bp, ok := d.breakpoints[number]
	if ok && bp != nil {
		bp.SetPass(0)
		bp.SetPassCount(count)
	}
}

func (d *Debugger) SetBreakpointEnabled(number uint16, enabled bool) {
	bp, ok := d.breakpoints[number]
	if ok && bp != nil {
		bp.SetEnabled(enabled)
	}
}

func (d *Debugger) BreakpointEnabled(number uint16) bool {
	bp, ok := d.breakpoints[number]
	if ok && bp != nil {
		return bp.Enabled()
	}
	return false
}

func (d *Debugger) BreakpointMBank(number uint16) uint8 {
	bp, ok := d.breakpoints[number]
	if ok && bp != nil {
		return bp.MBank()
	}
	return 1
}

func (d *Debugger) ClearMemBreakpoints() {
	for c := 0; c < 65536; c++ {
		d.memBreakpoints[c] = 0
	}
}

func (d *Debugger) StepMode() bool {
	return d.stepMode
}

func (d *Debugger) SetDoStep(on bool) {
	d.doStep = on
}

// BPExpression Return requested breakpoint
func (d *Debugger) BPExpression(number uint16) string {
	bp, ok := d.breakpoints[number]
	if ok && bp != nil {
		return bp.Expression()
	}
	return ""
}

// RunInst return and increment count of instructions executed
func (d *Debugger) RunInst() uint64 {
	v := d.runInst
	d.runInst++
	return v
}

func (d *Debugger) SetMemBreakpoint(address uint16, typ byte, size uint16) {
	var offset uint16
	for offset = address; offset < address+size; offset++ {
		d.memBreakpoints[offset] = typ
	}
}

func (d *Debugger) RemoveBreakpoint(number uint16) {
	delete(d.breakpoints, number)
}

func (d *Debugger) CheckMemBreakpoints(accessMap *map[uint16]byte) (bool, uint16, byte) {
	if !d.breakpointsEnabled {
		return false, 0, 0
	}
	for addr, typ := range *accessMap {
		bp := d.memBreakpoints[addr]
		if bp == 0 {
			return false, addr, 0
		}
		if (bp == 3) || bp == typ {
			d.SetRunMode(false)
			return true, addr, typ
		}
	}
	return false, 0, 0
}

func (d *Debugger) ClearBreakpoints() {
	clear(d.breakpoints)
}

type MemBP struct {
	addr uint16
	size uint16
}

func (d *Debugger) GetMemBreakpoints() []MemBP {
	var res []MemBP
	a := uint16(0)
	s := uint16(0)
	isBp := false
	for addr := 0; addr < 65536; addr++ {
		if d.memBreakpoints[addr] > 0 {
			if !isBp {
				isBp = true
				a = uint16(addr)
			}
			s++
			if addr == 65535 {
				res = append(res, MemBP{addr: a, size: s})
			}
		} else {
			if isBp {
				isBp = false
				res = append(res, MemBP{addr: a, size: s})
				s = 0
			}
		}
	}
	return res
}

func (m *MemBP) String() string {
	return fmt.Sprintf("%04XH : %d", m.addr, m.size)
}
