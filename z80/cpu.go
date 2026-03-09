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
	//core              MemIoRW
}
