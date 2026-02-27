package z80em

const SpDefault uint16 = 0xdff0

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

// StateType - Processor state
type StateType struct {
	A          byte
	B          byte
	C          byte
	D          byte
	E          byte
	H          byte
	L          byte
	APrime     byte
	BPrime     byte
	CPrime     byte
	DPrime     byte
	EPrime     byte
	HPrime     byte
	LPrime     byte
	IX         byte
	IY         byte
	I          byte
	R          byte
	SP         uint16
	PC         uint16
	Flags      FlagsType
	FlagsPrime FlagsType

	IMode             byte
	Iff1              byte
	Iff2              byte
	Halted            bool
	DoDelayedDI       bool
	DoDelayedEI       bool
	CycleCounter      byte
	interruptOccurred bool
	core              MemIoRW
}

type MemIoRW interface {
	// M1MemRead Read byte from memory for specified address
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

type Z80em interface {
	// Reset CPU to initial state
	Reset()
	// RunInstruction Run single instruction, return number of CPU cycles
	RunInstruction() byte
	// GetState Get current CPU state
	GetState() *StateType
	// SetState Set current CPU state
	SetState(state *StateType)
	// setFlagsRegister Set value for CPU flags by specified byte [7:0] = SZYHXPNC
	setFlagsRegister(flags byte)

	decodeInstruction(opcode byte) byte
	pushWord(pc uint16)
}

func (s StateType) Reset() {
	s.A = 0
	s.R = 0
	s.SP = SpDefault
	s.PC = 0
	s.setFlagsRegister(0)
	// Interrupts disabled
	s.IMode = 0
	s.Iff1 = 0
	s.Iff2 = 0
	s.interruptOccurred = false

	// Start not halted
	s.Halted = false
	s.DoDelayedDI = false
	s.DoDelayedEI = false
	// no cycles
	s.CycleCounter = 0
}

func (s StateType) GetState() *StateType {
	return &StateType{
		A:            s.A,
		B:            s.B,
		C:            s.C,
		D:            s.D,
		E:            s.E,
		H:            s.H,
		L:            s.L,
		APrime:       s.APrime,
		BPrime:       s.BPrime,
		CPrime:       s.CPrime,
		DPrime:       s.DPrime,
		EPrime:       s.EPrime,
		HPrime:       s.HPrime,
		IX:           s.IX,
		IY:           s.IY,
		R:            s.R,
		SP:           s.SP,
		PC:           s.PC,
		Flags:        s.Flags,
		FlagsPrime:   s.FlagsPrime,
		IMode:        s.IMode,
		Iff1:         s.Iff1,
		Iff2:         s.Iff2,
		Halted:       s.Halted,
		DoDelayedDI:  s.DoDelayedDI,
		DoDelayedEI:  s.DoDelayedEI,
		CycleCounter: s.CycleCounter,
	}
}

func (s StateType) SetState(state *StateType) {
	s.A = state.A
	s.B = state.B
	s.C = state.C
	s.D = state.D
	s.E = state.E
	s.H = state.H
	s.L = state.L
	s.APrime = state.APrime
	s.BPrime = state.BPrime
	s.CPrime = state.CPrime
	s.DPrime = state.DPrime
	s.EPrime = state.EPrime
	s.HPrime = state.HPrime
	s.IX = state.IX
	s.IY = state.IY
	s.I = state.I
	s.R = state.R
	s.SP = state.SP
	s.PC = state.PC
	s.Flags = state.Flags
	s.FlagsPrime = state.FlagsPrime
	s.IMode = state.IMode
	s.Iff1 = state.Iff1
	s.Iff2 = state.Iff2
	s.Halted = state.Halted
	s.DoDelayedDI = state.DoDelayedDI
	s.DoDelayedEI = state.DoDelayedEI
	s.CycleCounter = state.CycleCounter
}

// New Create new
func New(memIoRW MemIoRW) *StateType {
	return &StateType{
		A:      0,
		B:      0,
		C:      0,
		D:      0,
		E:      0,
		H:      0,
		L:      0,
		APrime: 0,
		BPrime: 0,
		CPrime: 0,
		DPrime: 0,
		EPrime: 0,
		HPrime: 0,
		IX:     0,
		IY:     0,
		I:      0,

		R:                 0,
		SP:                SpDefault,
		PC:                0,
		Flags:             FlagsType{false, false, false, false, false, false, false, false},
		FlagsPrime:        FlagsType{false, false, false, false, false, false, false, false},
		IMode:             0,
		Iff1:              0,
		Iff2:              0,
		Halted:            false,
		DoDelayedDI:       false,
		DoDelayedEI:       false,
		CycleCounter:      0,
		interruptOccurred: false,
		core:              memIoRW,
	}
}

func (s StateType) RunInstruction() byte {

	// R is incremented at the start of every instruction cycle,
	// before the instruction actually runs.
	// The high bit of R is not affected by this increment,
	// it can only be changed using the LD R, A instruction.
	// Note: also a HALT does increment the R register.
	s.R = (s.R & 0x80) | (((s.R & 0x7f) + 1) & 0x7f)

	if !s.Halted {
		// If the previous instruction was a DI or an EI,
		//  we'll need to disable or enable interrupts
		//  after whatever instruction we're about to run is finished.
		doingDelayedDi := false
		doingDelayedEi := false
		if s.DoDelayedDI {
			s.DoDelayedDI = false
			doingDelayedDi = true
		} else if s.DoDelayedEI {
			s.DoDelayedEI = false
			doingDelayedEi = true
		}

		// Read the byte at the PC and run the instruction it encodes.
		opcode := s.core.M1MemRead(s.PC)
		s.decodeInstruction(opcode)

		// HALT does not increase the PC
		if !s.Halted {
			s.PC++
		}

		// Actually do the delayed interrupt disable/enable if we have one.
		if doingDelayedDi {
			s.Iff1 = 0
			s.Iff2 = 0
		} else if doingDelayedEi {
			s.Iff1 = 1
			s.Iff2 = 1
		}

		// And finally clear out the cycle counter for the next instruction
		//  before returning it to the emulator core.
		cycleCounter := s.CycleCounter
		s.CycleCounter = 0
		return cycleCounter
	} else { // HALTED
		// During HALT, NOPs are executed which is 4T
		s.core.M1MemRead(s.PC) // HALT does a normal M1 fetch to keep the memory refresh active. The result is ignored (NOP).
		return 4
	}
}

// Simulates pulsing the processor's INT (or NMI) pin
//
//	nonMaskable - true if this is a non-maskable interrupt
//	data - the value to be placed on the data bus, if needed
func (s StateType) interrupt(nonMaskable bool, data byte) {
	if nonMaskable {
		// An interrupt, if halted, does increase the PC
		if s.Halted {
			s.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R, A instruction.
		s.R = (s.R & 0x80) | (((s.R & 0x7f) + 1) & 0x7f)

		// Non-maskable interrupts are always handled the same way;
		// clear IFF1 and then do a CALL 0x0066.
		// Also, all interrupts reset the HALT state.

		s.Halted = false
		s.Iff2 = s.Iff1
		s.Iff1 = 0
		s.pushWord(s.PC)
		s.PC = 0x66

		s.CycleCounter += 11
	} else if s.Iff1 != 0 {
		//  An interrupt, if halted, does increase the PC
		if s.Halted {
			s.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R,A instruction.
		s.R = (s.R & 0x80) | (((s.R & 0x7f) + 1) & 0x7f)

		s.Halted = false
		s.Iff1 = 0
		s.Iff2 = 0

		if s.IMode == 0 {
			// In the 8080-compatible interrupt mode,
			// decode the content of the data bus as an instruction and run it.
			// it's probably a RST instruction, which pushes (PC+1) onto the stack
			// so we should decrement PC before we decode the instruction
			s.PC--
			s.decodeInstruction(data)
			s.PC++ // increment PC upon return
			s.CycleCounter += 2
		} else if s.IMode == 1 {
			// Mode 1 is always just RST 0x38.
			s.pushWord(s.PC)
			s.PC = 0x0038
			s.CycleCounter += 13
		} else if s.IMode == 2 {
			// Mode 2 uses the value on the data bus as in index
			// into the vector table pointer to by the I register.
			s.pushWord(s.PC)

			// The Z80 manual says that this address must be 2-byte aligned,
			//  but it doesn't appear that this is actually the case on the hardware,
			//  so we don't attempt to enforce that here.
			vectorAddress := (uint16(s.I) << 8) | uint16(data)
			s.PC = uint16(s.core.MemRead(vectorAddress)) | (uint16(s.core.MemRead(vectorAddress+1)) << 8)
			s.CycleCounter += 19
			// A "notification" is generated so that the calling program can break on it.
			s.interruptOccurred = true
		}
	}
}

func (s StateType) pushWord(pc uint16) {
	// TODO: Implement
	panic("not yet implemented")
}

func (s StateType) getOperand(opcode byte) byte {
	switch opcode & 0x07 {
	case 0:
		return s.B
	case 1:
		return s.C
	case 2:
		return s.D
	case 3:
		return s.E
	case 4:
		return s.H
	case 5:
		return s.L
	case 6:
		return s.core.MemRead(uint16(s.H)<<8 | uint16(s.L))
	default:
		return s.A
	}
}

func (s StateType) decodeInstruction(opcode byte) {
	// Handle HALT right up front, because it fouls up our LD decoding
	//  by falling where LD (HL), (HL) ought to be.
	if opcode == OP_HALT {
		s.Halted = true
	} else if opcode >= OP_LD_B_B && opcode < OP_ADD_A_B {
		// This entire range is all 8-bit register loads.
		// Get the operand and assign it to the correct destination.
		s.load8bit(opcode, s.getOperand(opcode))
	} else if (opcode >= OP_ADD_A_B) && (opcode < OP_RET_NZ) {
		// These are the 8-bit register ALU instructions.
		// We'll get the operand and then use this "jump table"
		// to call the correct utility function for the instruction.
		s.alu8bit(opcode, s.getOperand(opcode))
	} else {
		// This is one of the less formulaic instructions;
		//  we'll get the specific function for it from our array.
		s.otherInstructions(opcode)
	}
	s.CycleCounter += CYCLE_COUNTS[opcode]
}

func (s StateType) load8bit(opcode byte, operand byte) {
	switch (opcode & 0x38) >> 3 {
	case 0:
		s.B = operand
	case 1:
		s.C = operand
	case 2:
		s.D = operand
	case 3:
		s.E = operand
	case 4:
		s.H = operand
	case 5:
		s.L = operand
	case 6:
		s.core.MemWrite(uint16(s.H)<<8|uint16(s.L), operand)
	default:
		s.A = operand
	}
}

func (s StateType) alu8bit(opcode byte, operand byte) {
	switch (opcode & 0x38) >> 3 {
	case 0:
		s.doAdd(operand)
	case 1:
		s.doAdc(operand)
	case 2:
		s.doSub(operand)
	case 3:
		s.doSbc(operand)
	case 4:
		s.doAnd(operand)
	case 5:
		s.doXor(operand)
	case 6:
		s.doOr(operand)
	default:
		s.doCp(operand)
	}
}

func (s StateType) doAdd(operand byte) {

}

func (s StateType) doAdc(operand byte) {

}

func (s StateType) doSub(operand byte) {

}

func (s StateType) doSbc(operand byte) {

}

func (s StateType) doAnd(operand byte) {

}

func (s StateType) doXor(operand byte) {

}

func (s StateType) doOr(operand byte) {

}

func (s StateType) doCp(operand byte) {

}

func (s StateType) otherInstructions(opcode byte) {

}

// getFlagsRegister return whole F register
func (s StateType) getFlagsRegister() byte {
	return getFlags(&s.Flags)
}

// getFlagsRegister return whole F' register
func (s StateType) getFlagsPrimeRegister() byte {
	return getFlags(&s.FlagsPrime)
}

func getFlags(f *FlagsType) byte {
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

func (s StateType) setFlagsRegister(flags byte) {
	setFlags(flags, &s.Flags)
}

func (s StateType) setFlagsPrimeRegister(flags byte) {
	setFlags(flags, &s.FlagsPrime)
}

func setFlags(flags byte, f *FlagsType) {
	f.S = flags&0x80 != 0
	f.Z = flags&0x40 != 0
	f.Y = flags&0x20 != 0
	f.H = flags&0x10 != 0
	f.X = flags&0x08 != 0
	f.P = flags&0x04 != 0
	f.N = flags&0x02 != 0
	f.C = flags&0x01 != 0
}

func (s StateType) updateXYFlags(result byte) {
	// Most of the time, the undocumented flags
	//  (sometimes called X and Y, or 3 and 5),
	//  take their values from the corresponding bits
	//  of the result of the instruction,
	//  or from some other related value.
	// This is a utility function to set those flags based on those bits.
	s.Flags.Y = (result&0x20)>>5 != 0
	s.Flags.X = (result&0x08)>>3 != 0
}

func getParity(value byte) bool {
	return PARITY_BITS[value]
}

func (s StateType) PushWord(operand uint16) {
	// Pretty obvious what this function does; given a 16-bit value,
	//  decrement the stack pointer, write the high byte to the new
	//  stack pointer location, then repeat for the low byte.
	s.SP--
	s.core.MemWrite(s.SP, byte((operand&0xff00)>>8))
	s.SP--
	s.core.MemWrite(s.SP, byte(operand&0x00ff))
}

func (s StateType) PopWord() uint16 {
	// Again, not complicated; read a byte off the top of the stack,
	//  increment the stack pointer, rinse and repeat.
	result := uint16(s.core.MemRead(s.SP))
	s.SP++
	result |= uint16(s.core.MemRead(s.SP)) << 8
	s.SP++
	return result
}
