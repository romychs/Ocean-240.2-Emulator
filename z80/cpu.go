package z80

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
	cyc uint64

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
func (z *Z80) RunInstruction() {
	if z.halted {
		z.exec_opcode(0x00)
	} else {
		opcode := z.nextb()
		z.exec_opcode(opcode)
	}
	z.process_interrupts()
}

func (z *Z80) GetState() *Z80 {
	return &Z80{
		cyc:     z.cyc,
		pc:      z.pc,
		sp:      z.sp,
		ix:      z.ix,
		iy:      z.iy,
		mem_ptr: z.mem_ptr,

		a:  z.a,
		b:  z.b,
		c:  z.c,
		d:  z.d,
		e:  z.e,
		h:  z.h,
		l:  z.l,
		a_: z.a_,
		b_: z.b_,
		c_: z.c_,
		d_: z.d_,
		e_: z.e_,
		h_: z.h_,
		l_: z.l_,
		f_: z.f_,

		i: z.i,
		r: z.r,

		sf: z.sf,
		zf: z.zf,
		yf: z.yf,
		hf: z.hf,
		xf: z.xf,
		pf: z.pf,
		nf: z.nf,
		cf: z.cf,

		iff1: z.iff1,
		iff2: z.iff2,

		iff_delay:      z.iff_delay,
		interrupt_mode: z.interrupt_mode,
		halted:         z.halted,
		int_pending:    z.int_pending,
		nmi_pending:    z.nmi_pending,
		core:           z.core,
	}
}
