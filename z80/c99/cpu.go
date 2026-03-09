package c99

import "okemu/z80"

type MemIoRW interface {
	// M1MemRead Read byte from memory for specified address @ M1 cycle
	M1MemRead(addr uint16) byte
	// MemRead Read byte from memory for specified address
	MemRead(addr uint16) byte
	// MemWrite Write byte to memory to specified address
	MemWrite(addr uint16, val byte)
	// IORead Read byte from specified port
	IORead(port uint16) byte
	// IOWrite Write byte to specified port
	IOWrite(port uint16, val byte)
}

type CPUInterface interface {
	// Reset CPU to initial state
	Reset()
	// RunInstruction Run single instruction, return number of CPU cycles
	RunInstruction() byte
	// GetState Get current CPU state
	GetState() *Z80
	// SetState Set current CPU state
	SetState(state *Z80)
}

type Z80 struct {

	// cycle count (t-states)
	cyc      uint64
	inst_cyc byte

	// special purpose registers
	pc, sp, ix, iy uint16

	// "wz" register
	mem_ptr uint16

	// main registers
	a, b, c, d, e, h, l byte

	// alternate registers
	a_, b_, c_, d_, e_, h_, l_, f_ byte

	// interrupt vector, memory refresh
	i, r byte

	// flags: sign, zero, yf, half-carry, xf, parity/overflow, negative, carry
	sf, zf, yf, hf, xf, pf, nf, cf bool
	iff_delay                      byte
	interrupt_mode                 byte
	int_data                       byte
	iff1                           bool
	iff2                           bool
	halted                         bool
	int_pending                    bool
	nmi_pending                    bool

	core MemIoRW
}

// New initializes a Z80 instance and return pointer to it
func New(core MemIoRW) *Z80 {
	z := Z80{}
	z.core = core

	z.cyc = 0

	z.pc = 0
	z.sp = 0xFFFF
	z.ix = 0
	z.iy = 0
	z.mem_ptr = 0

	// af and sp are set to 0xFFFF after reset,
	// and the other values are undefined (z80-documented)
	z.a = 0xFF
	z.b = 0
	z.c = 0
	z.d = 0
	z.e = 0
	z.h = 0
	z.l = 0

	z.a_ = 0
	z.b_ = 0
	z.c_ = 0
	z.d_ = 0
	z.e_ = 0
	z.h_ = 0
	z.l_ = 0
	z.f_ = 0

	z.i = 0
	z.r = 0

	z.sf = true
	z.zf = true
	z.yf = true
	z.hf = true
	z.xf = true
	z.pf = true
	z.nf = true
	z.cf = true

	z.iff_delay = 0
	z.interrupt_mode = 0
	z.iff1 = false
	z.iff2 = false
	z.halted = false
	z.int_pending = false
	z.nmi_pending = false
	z.int_data = 0
	return &z
}

// RunInstruction executes the next instruction in memory + handles interrupts
func (z *Z80) RunInstruction() uint64 {
	pre := z.cyc
	if z.halted {
		z.exec_opcode(0x00)
	} else {
		opcode := z.nextb()
		z.exec_opcode(opcode)
	}
	z.process_interrupts()
	return z.cyc - pre
}

func (z *Z80) GetState() *z80.Z80CPU {
	return &z80.Z80CPU{
		A:                 z.a,
		B:                 z.b,
		C:                 z.c,
		D:                 z.d,
		E:                 z.e,
		H:                 z.h,
		L:                 z.l,
		AAlt:              z.a_,
		BAlt:              z.b_,
		CAlt:              z.c_,
		DAlt:              z.d_,
		EAlt:              z.e_,
		HAlt:              z.h_,
		LAlt:              z.l_,
		IX:                z.ix,
		IY:                z.iy,
		I:                 z.i,
		R:                 z.r,
		SP:                z.sp,
		PC:                z.pc,
		Flags:             z.getFlags(),
		FlagsAlt:          z.getAltFlags(),
		IMode:             z.interrupt_mode,
		Iff1:              z.iff1,
		Iff2:              z.iff2,
		Halted:            z.halted,
		DoDelayedDI:       z.int_pending,
		DoDelayedEI:       z.int_pending,
		CycleCounter:      z.inst_cyc,
		InterruptOccurred: false,
	}
}

func (z *Z80) getFlags() z80.FlagsType {
	return z80.FlagsType{
		S: z.sf,
		Z: z.zf,
		Y: z.yf,
		H: z.hf,
		X: z.xf,
		P: z.pf,
		N: z.nf,
		C: z.cf,
	}
}

func (z *Z80) getAltFlags() z80.FlagsType {
	return z80.FlagsType{
		S: z.f_&0x80 != 0,
		Z: z.f_&0x40 != 0,
		Y: z.f_&0x20 != 0,
		H: z.f_&0x10 != 0,
		X: z.f_&0x08 != 0,
		P: z.f_&0x04 != 0,
		N: z.f_&0x02 != 0,
		C: z.f_&0x01 != 0,
	}
}
