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
	RunInstruction() uint64
	// GetState Get current CPU state
	GetState() *Z80CPU
	// SetState Set current CPU state
	SetState(state *Z80CPU)
	// DebugOutput out current CPU state
	DebugOutput()
}

// FlagsType - Processor flags
type FlagsType struct {
	S bool
	Z bool
	Y bool
	H bool
	X bool
	P bool
	N bool
	C bool
}

// Z80CPU - Processor state
type Z80CPU struct {
	A                 byte
	B                 byte
	C                 byte
	D                 byte
	E                 byte
	H                 byte
	L                 byte
	AAlt              byte
	BAlt              byte
	CAlt              byte
	DAlt              byte
	EAlt              byte
	HAlt              byte
	LAlt              byte
	IX                uint16
	IY                uint16
	I                 byte
	R                 byte
	SP                uint16
	PC                uint16
	Flags             FlagsType
	FlagsAlt          FlagsType
	IMode             byte
	Iff1              bool
	Iff2              bool
	Halted            bool
	DoDelayedDI       bool
	DoDelayedEI       bool
	CycleCounter      byte
	InterruptOccurred bool
	MemPtr            uint16
	//core              MemIoRW
}

func (f *FlagsType) GetFlags() byte {
	var flags byte = 0
	if f.S {
		flags |= 0x80
	}
	if f.Z {
		flags |= 0x40
	}
	if f.Y {
		flags |= 0x20
	}
	if f.H {
		flags |= 0x10
	}
	if f.X {
		flags |= 0x08
	}
	if f.P {
		flags |= 0x04
	}
	if f.N {
		flags |= 0x02
	}
	if f.C {
		flags |= 0x01
	}
	return flags
}

func GetFlags(f byte) FlagsType {
	return FlagsType{
		S: f&0x80 != 0,
		Z: f&0x40 != 0,
		Y: f&0x20 != 0,
		H: f&0x10 != 0,
		X: f&0x08 != 0,
		P: f&0x04 != 0,
		N: f&0x02 != 0,
		C: f&0x01 != 0,
	}
}

func (f *FlagsType) SetFlags(flags byte) {
	f.S = flags&0x80 != 0
	f.Z = flags&0x40 != 0
	f.Y = flags&0x20 != 0
	f.H = flags&0x10 != 0
	f.X = flags&0x08 != 0
	f.P = flags&0x04 != 0
	f.N = flags&0x02 != 0
	f.C = flags&0x01 != 0
}
