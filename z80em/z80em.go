package z80em

import "fmt"

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

// Z80Type - Processor state
type Z80Type struct {
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
	IX         uint16
	IY         uint16
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

type CPUInterface interface {
	// Reset CPU to initial state
	Reset()
	// RunInstruction Run single instruction, return number of CPU cycles
	RunInstruction() byte
	// GetState Get current CPU state
	GetState() *Z80Type
	// SetState Set current CPU state
	SetState(state *Z80Type)
}

func (z *Z80Type) Reset() {
	z.A = 0
	z.R = 0
	z.SP = SpDefault
	z.PC = 0
	z.setFlagsRegister(0)
	// Interrupts disabled
	z.IMode = 0
	z.Iff1 = 0
	z.Iff2 = 0
	z.interruptOccurred = false

	// Start not halted
	z.Halted = false
	z.DoDelayedDI = false
	z.DoDelayedEI = false
	// no cycles
	z.CycleCounter = 0
	fmt.Println("CPUInterface State Reset")
}

func (z *Z80Type) GetState() *Z80Type {
	return &Z80Type{
		A:            z.A,
		B:            z.B,
		C:            z.C,
		D:            z.D,
		E:            z.E,
		H:            z.H,
		L:            z.L,
		APrime:       z.APrime,
		BPrime:       z.BPrime,
		CPrime:       z.CPrime,
		DPrime:       z.DPrime,
		EPrime:       z.EPrime,
		HPrime:       z.HPrime,
		IX:           z.IX,
		IY:           z.IY,
		R:            z.R,
		SP:           z.SP,
		PC:           z.PC,
		Flags:        z.Flags,
		FlagsPrime:   z.FlagsPrime,
		IMode:        z.IMode,
		Iff1:         z.Iff1,
		Iff2:         z.Iff2,
		Halted:       z.Halted,
		DoDelayedDI:  z.DoDelayedDI,
		DoDelayedEI:  z.DoDelayedEI,
		CycleCounter: z.CycleCounter,
	}
}

func (z *Z80Type) SetState(state *Z80Type) {
	z.A = state.A
	z.B = state.B
	z.C = state.C
	z.D = state.D
	z.E = state.E
	z.H = state.H
	z.L = state.L
	z.APrime = state.APrime
	z.BPrime = state.BPrime
	z.CPrime = state.CPrime
	z.DPrime = state.DPrime
	z.EPrime = state.EPrime
	z.HPrime = state.HPrime
	z.IX = state.IX
	z.IY = state.IY
	z.I = state.I
	z.R = state.R
	z.SP = state.SP
	z.PC = state.PC
	z.Flags = state.Flags
	z.FlagsPrime = state.FlagsPrime
	z.IMode = state.IMode
	z.Iff1 = state.Iff1
	z.Iff2 = state.Iff2
	z.Halted = state.Halted
	z.DoDelayedDI = state.DoDelayedDI
	z.DoDelayedEI = state.DoDelayedEI
	z.CycleCounter = state.CycleCounter
}

// New Create new
func New(memIoRW MemIoRW) *Z80Type {
	return &Z80Type{
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

func (z *Z80Type) RunInstruction() byte {

	// R is incremented at the start of every instruction cycle,
	// before the instruction actually runs.
	// The high bit of R is not affected by this increment,
	// it can only be changed using the LD R, A instruction.
	// Note: also a HALT does increment the R register.
	z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)

	if !z.Halted {
		// If the previous instruction was a DI or an EI,
		//  we'll need to disable or enable interrupts
		//  after whatever instruction we're about to run is finished.
		doingDelayedDi := false
		doingDelayedEi := false
		if z.DoDelayedDI {
			z.DoDelayedDI = false
			doingDelayedDi = true
		} else if z.DoDelayedEI {
			z.DoDelayedEI = false
			doingDelayedEi = true
		}

		// Read the byte at the PC and run the instruction it encodes.
		opcode := z.core.M1MemRead(z.PC)
		z.decodeInstruction(opcode)

		// HALT does not increase the PC
		if !z.Halted {
			z.PC++
		}

		// Actually do the delayed interrupt disable/enable if we have one.
		if doingDelayedDi {
			z.Iff1 = 0
			z.Iff2 = 0
		} else if doingDelayedEi {
			z.Iff1 = 1
			z.Iff2 = 1
		}

		// And finally clear out the cycle counter for the next instruction
		//  before returning it to the emulator core.
		cycleCounter := z.CycleCounter
		z.CycleCounter = 0
		return cycleCounter
	} else { // HALTED
		// During HALT, NOPs are executed which is 4T
		z.core.M1MemRead(z.PC) // HALT does a normal M1 fetch to keep the memory refresh active. The result is ignored (NOP).
		return 4
	}
}

// Simulates pulsing the processor's INT (or NMI) pin
//
//	nonMaskable - true if this is a non-maskable interrupt
//	data - the value to be placed on the data bus, if needed
func (z *Z80Type) interrupt(nonMaskable bool, data byte) {
	if nonMaskable {
		// An interrupt, if halted, does increase the PC
		if z.Halted {
			z.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R, A instruction.
		z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)

		// Non-maskable interrupts are always handled the same way;
		// clear IFF1 and then do a CALL 0x0066.
		// Also, all interrupts reset the HALT state.

		z.Halted = false
		z.Iff2 = z.Iff1
		z.Iff1 = 0
		z.pushWord(z.PC)
		z.PC = 0x66

		z.CycleCounter += 11
	} else if z.Iff1 != 0 {
		//  An interrupt, if halted, does increase the PC
		if z.Halted {
			z.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R,A instruction.
		z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)

		z.Halted = false
		z.Iff1 = 0
		z.Iff2 = 0

		if z.IMode == 0 {
			// In the 8080-compatible interrupt mode,
			// decode the content of the data bus as an instruction and run it.
			// it'z probably a RST instruction, which pushes (PC+1) onto the stack
			// so we should decrement PC before we decode the instruction
			z.PC--
			z.decodeInstruction(data)
			z.PC++ // increment PC upon return
			z.CycleCounter += 2
		} else if z.IMode == 1 {
			// Mode 1 is always just RST 0x38.
			z.pushWord(z.PC)
			z.PC = 0x0038
			z.CycleCounter += 13
		} else if z.IMode == 2 {
			// Mode 2 uses the value on the data bus as in index
			// into the vector table pointer to by the I register.
			z.pushWord(z.PC)

			// The Z80 manual says that this address must be 2-byte aligned,
			//  but it doesn't appear that this is actually the case on the hardware,
			//  so we don't attempt to enforce that here.
			vectorAddress := (uint16(z.I) << 8) | uint16(data)
			z.PC = uint16(z.core.MemRead(vectorAddress)) | (uint16(z.core.MemRead(vectorAddress+1)) << 8)
			z.CycleCounter += 19
			// A "notification" is generated so that the calling program can break on it.
			z.interruptOccurred = true
		}
	}
}

func (z *Z80Type) pushWord(operand uint16) {
	z.SP--
	z.core.MemWrite(z.SP, byte(operand>>8))
	z.SP--
	z.core.MemWrite(z.SP, byte(operand&0x00ff))
}

func (z *Z80Type) getOperand(opcode byte) byte {
	switch opcode & 0x07 {
	case 0:
		return z.B
	case 1:
		return z.C
	case 2:
		return z.D
	case 3:
		return z.E
	case 4:
		return z.H
	case 5:
		return z.L
	case 6:
		return z.core.MemRead(uint16(z.H)<<8 | uint16(z.L))
	default:
		return z.A
	}
}

func (z *Z80Type) decodeInstruction(opcode byte) {
	// Handle HALT right up front, because it fouls up our LD decoding
	//  by falling where LD (HL), (HL) ought to be.
	if opcode == OpHalt {
		z.Halted = true
	} else if opcode >= OpLdBB && opcode < OpAddAB {
		// This entire range is all 8-bit register loads.
		// Get the operand and assign it to the correct destination.
		z.load8bit(opcode, z.getOperand(opcode))
	} else if (opcode >= OpAddAB) && (opcode < OpRetNz) {
		// These are the 8-bit register ALU instructions.
		// We'll get the operand and then use this "jump table"
		// to call the correct utility function for the instruction.
		z.alu8bit(opcode, z.getOperand(opcode))
	} else {
		// This is one of the less formulaic instructions;
		//  we'll get the specific function for it from our array.
		fun := instructions[opcode]
		fun(z)
		//z.otherInstructions(opcode)
	}
	z.CycleCounter += CycleCounts[opcode]
}

func (z *Z80Type) load8bit(opcode byte, operand byte) {
	switch (opcode & 0x38) >> 3 {
	case 0:
		z.B = operand
	case 1:
		z.C = operand
	case 2:
		z.D = operand
	case 3:
		z.E = operand
	case 4:
		z.H = operand
	case 5:
		z.L = operand
	case 6:
		z.core.MemWrite(uint16(z.H)<<8|uint16(z.L), operand)
	default:
		z.A = operand
	}
}

func (z *Z80Type) alu8bit(opcode byte, operand byte) {
	switch (opcode & 0x38) >> 3 {
	case 0:
		z.doAdd(operand)
	case 1:
		z.doAdc(operand)
	case 2:
		z.doSub(operand)
	case 3:
		z.doSbc(operand)
	case 4:
		z.doAnd(operand)
	case 5:
		z.doXor(operand)
	case 6:
		z.doOr(operand)
	default:
		z.doCp(operand)
	}
}

func (z *Z80Type) otherInstructions(opcode byte) {

}

// getFlagsRegister return whole F register
func (z *Z80Type) getFlagsRegister() byte {
	return getFlags(&z.Flags)
}

// getFlagsRegister return whole F' register
func (z *Z80Type) getFlagsPrimeRegister() byte {
	return getFlags(&z.FlagsPrime)
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

func (z *Z80Type) setFlagsRegister(flags byte) {
	setFlags(flags, &z.Flags)
}

func (z *Z80Type) setFlagsPrimeRegister(flags byte) {
	setFlags(flags, &z.FlagsPrime)
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

func (z *Z80Type) updateXYFlags(result byte) {
	// Most of the time, the undocumented flags
	//  (sometimes called X and Y, or 3 and 5),
	//  take their values from the corresponding bits
	//  of the result of the instruction,
	//  or from some other related value.
	// This is a utility function to set those flags based on those bits.
	z.Flags.Y = (result&0x20)>>5 != 0
	z.Flags.X = (result&0x08)>>3 != 0
}

func getParity(value byte) bool {
	return ParityBits[value]
}

// PushWord - Decrement stack pointer and put specified word value to stack
func (z *Z80Type) PushWord(operand uint16) {
	z.SP--
	z.core.MemWrite(z.SP, byte((operand&0xff00)>>8))
	z.SP--
	z.core.MemWrite(z.SP, byte(operand&0x00ff))
}

// PopWord - Get word value from top of stack and increment stack pointer
func (z *Z80Type) PopWord() uint16 {
	result := uint16(z.core.MemRead(z.SP))
	z.SP++
	result |= uint16(z.core.MemRead(z.SP)) << 8
	z.SP++
	return result
}

// doConditionalAbsoluteJump - Implements the JP [condition],nn instructions.
func (z *Z80Type) doConditionalAbsoluteJump(condition bool) {
	if condition {
		// We're taking this jump, so write the new PC,
		//  and then decrement the thing we just wrote,
		//  because the instruction decoder increments the PC
		//  unconditionally at the end of every instruction,
		//  and we need to counteract that so we end up at the jump target.
		// TODO: Check for increment CycleCounter
		z.PC = uint16(z.core.MemRead(z.PC+1)) | (uint16(z.core.MemRead(z.PC+2)) << 8)
		z.PC--
	} else {
		// We're not taking this jump, just move the PC past the operand.
		z.PC += 2
	}
}

// doConditionalRelativeJump - Implements the JR [condition],nn instructions.
func (z *Z80Type) doConditionalRelativeJump(condition bool) {
	if condition {
		// We need a few more cycles to actually take the jump.
		z.CycleCounter += 5
		// Calculate the offset specified by our operand.
		offset := z.core.MemRead(z.PC + 1)

		// Add the offset to the PC, also skipping past this instruction.
		if offset < 0 {
			z.PC = z.PC - uint16(-offset)
		} else {
			z.PC = z.PC + uint16(offset)
		}

	}
	z.PC++
}

// doConditionalCall - Implements CALL [condition],nn instructions.
func (z *Z80Type) doConditionalCall(condition bool) {
	if condition {
		z.CycleCounter += 7
		z.PushWord(z.PC + 3)
		z.PC = uint16(z.core.MemRead(z.PC+1)) | (uint16(z.core.MemRead(z.PC+2)) << 8)
		z.PC--
	} else {
		z.PC += 2
	}
}

func (z *Z80Type) doConditionalReturn(condition bool) {
	if condition {
		z.CycleCounter += 6
		z.PC = z.PopWord() - 1
	}
}

// doReset - Implements RST [address] instructions.
func (z *Z80Type) doReset(address uint16) {
	z.PushWord(z.PC + 1)
	z.PC = address - 1
}

func (z *Z80Type) setBaseFlags(operand byte, result uint16) {
	z.Flags.C = result > 0x00ff
	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result&0xff == 0
	z.Flags.H = (((operand & 0x0f) + (z.A & 0x0f)) & 0x10) != 0
	// An overflow has happened if the sign bits of the accumulator and the operand
	//  don't match the sign bit of the result value.
	z.Flags.P = ((z.A & 0x80) == (operand & 0x80)) && (z.A&0x80 != byte(result&0x80))
}

func (z *Z80Type) doAdd(operand byte) {
	// This is the ADD A, [operand] instructions.
	// We'll do the literal addition, which includes any overflow,
	//  so that we can more easily figure out whether we had
	//  an overflow or a carry and set the flags accordingly.
	var result = uint16(z.A) + uint16(operand)
	z.A = byte(result & 0xff)
	z.setBaseFlags(operand, result)
	z.Flags.N = false
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doAdc(operand byte) {
	var result = uint16(z.A) + uint16(operand)
	if z.Flags.C {
		result++
	}
	z.A = byte(result & 0xff)
	z.setBaseFlags(operand, result)
	z.Flags.N = false
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doSub(operand byte) {
	var result = uint16(z.A) - uint16(operand)
	z.A = byte(result & 0xff)
	z.setBaseFlags(operand, result)
	z.Flags.N = true
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doSbc(operand byte) {
	var result = uint16(z.A) - uint16(operand)
	if z.Flags.C {
		result--
	}
	z.A = byte(result & 0xff)
	z.setBaseFlags(operand, result)
	z.Flags.N = true
	z.updateXYFlags(z.A)
}

func (z *Z80Type) setLogicFlags() {
	z.Flags.S = z.A&0x80 != 0
	z.Flags.Z = z.A == 0
	z.Flags.H = true
	z.Flags.P = ParityBits[z.A]
	z.Flags.N = false
	z.Flags.C = false
}

func (z *Z80Type) doAnd(operand byte) {
	z.A &= operand
	z.setLogicFlags()
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doXor(operand byte) {
	z.A ^= operand
	z.setLogicFlags()
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doOr(operand byte) {
	z.A |= operand
	z.setLogicFlags()
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doCp(operand byte) {
	tmp := z.A
	z.doSub(operand)
	z.A = tmp
	z.updateXYFlags(operand)
}

func (z *Z80Type) doInc(operand byte) byte {
	var result = uint16(operand) + 1
	r8 := byte(result & 0xff)
	z.Flags.S = r8&0x80 > 0
	z.Flags.Z = r8 == 0
	z.Flags.H = (operand & 0x0f) == 0x0f
	// It'z a good deal easier to detect overflow for an increment/decrement.
	z.Flags.P = operand == 0x7f
	z.Flags.N = false
	z.updateXYFlags(r8)
	return r8
}

func (z *Z80Type) doDec(operand byte) byte {
	var result = uint16(operand) - 1
	r8 := byte(result & 0xff)
	z.Flags.S = r8&0x80 > 0
	z.Flags.Z = r8 == 0
	z.Flags.H = (operand & 0x0f) == 0x00
	z.Flags.P = operand == 0x80
	z.Flags.N = true
	z.updateXYFlags(r8)
	return r8
}

func (z *Z80Type) doHlAdd(operand uint16) {
	// The HL arithmetic instructions are the same as the A ones,
	//  just with twice as many bits happening.
	hl := uint16(z.L) | (uint16(z.H) << 8)
	result := uint32(hl) + uint32(operand)
	z.Flags.N = false
	z.Flags.C = result > 0xffff
	z.Flags.H = ((hl&0x0fff)+(operand&0x0fff))&0x1000 > 0

	z.L = byte(result & 0xff)
	z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

func (z *Z80Type) doHlAdc(operand uint16) {
	if z.Flags.C {
		operand++
	}

	hl := uint16(z.L) | (uint16(z.H) << 8)
	result := uint32(hl) + uint32(operand)

	z.Flags.S = (result & 0x8000) != 0
	z.Flags.Z = result&0xffff == 0
	z.Flags.H = (((hl & 0x0fff) + (operand & 0x0fff)) & 0x1000) != 0
	z.Flags.P = ((hl & 0x8000) == (operand & 0x8000)) && (uint16(result&0x8000) != (hl & 0x8000))
	z.Flags.N = false
	z.Flags.C = result > 0xffff

	z.L = byte(result & 0xff)
	z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

func (z *Z80Type) doHlSbc(operand uint16) {
	if z.Flags.C {
		operand++
	}

	hl := uint16(z.L) | (uint16(z.H) << 8)
	result := uint32(hl) - uint32(operand)

	z.Flags.S = (result & 0x8000) != 0
	z.Flags.Z = result&0xffff == 0
	z.Flags.H = (((hl & 0x0fff) - (operand & 0x0fff)) & 0x1000) != 0
	z.Flags.P = ((hl & 0x8000) != (operand & 0x8000)) && (uint16(result&0x8000) != (hl & 0x8000))
	z.Flags.N = true
	z.Flags.C = result > 0xffff

	z.L = byte(result & 0xff)
	z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

func (z *Z80Type) doIn(port uint16) byte {
	var result = z.core.IORead(port)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = false
	z.Flags.P = ParityBits[result]
	z.Flags.N = false
	z.updateXYFlags(result)
	return result
}

func (z *Z80Type) doNeg() {
	// This instruction is defined to not alter the register if it === 0x80.
	a := int8(z.A)
	if z.A != 0x80 {
		// This is a signed operation, so convert A to a signed value.
		z.A = byte(-a)
	}
	z.Flags.S = z.A&0x80 != 0
	z.Flags.Z = z.A == 0
	z.Flags.H = ((-a) & 0x0f) > 0
	z.Flags.P = z.A == 0x80
	z.Flags.N = true
	z.Flags.C = z.A != 0
	z.updateXYFlags(z.A)
}

func (z *Z80Type) doLdi() {
	// Copy the value that we're supposed to copy.
	hl := uint16(z.L) | (uint16(z.H) << 8)
	de := uint16(z.E) | (uint16(z.D) << 8)
	readValue := z.core.MemRead(hl)
	z.core.MemWrite(de, readValue)

	z.incDe()
	z.incHl()
	z.decBc()

	z.Flags.H = false
	z.Flags.P = (z.C | z.B) != 0
	z.Flags.N = false
	z.Flags.Y = ((z.A+readValue)&0x02)>>1 != 0
	z.Flags.X = ((z.A+readValue)&0x08)>>3 != 0
}

func (z *Z80Type) doCpi() {
	tempCarry := z.Flags.C
	hl := uint16(z.L) | (uint16(z.H) << 8)
	readValue := z.core.MemRead(hl)
	z.doCp(readValue)

	z.Flags.C = tempCarry
	var fh byte = 0
	if z.Flags.H {
		fh = 1
	}
	z.Flags.Y = ((z.A-readValue-fh)&0x02)>>1 != 0
	z.Flags.X = ((z.A-readValue-fh)&0x08)>>3 != 0

	z.incHl()
	z.decBc()

	z.Flags.P = (z.B | z.C) != 0
}

func (z *Z80Type) doIni() {
	hl := (uint16(z.H) << 8) | uint16(z.L)
	bc := (uint16(z.B) << 8) | uint16(z.C)
	z.core.MemWrite(hl, z.core.IORead(bc))

	z.incHl()

	z.B = z.doDec(z.B)
	z.Flags.N = true
}

func (z *Z80Type) doOuti() {
	// Zilog pseudo code is wrong, see: https://github.com/maziac/z80-instruction-set/pull/10
	z.B = z.doDec(z.B)
	hl := (uint16(z.H) << 8) | uint16(z.L)
	bc := (uint16(z.B) << 8) | uint16(z.C)
	z.core.IOWrite(bc, z.core.MemRead(hl))

	z.incHl()

	z.Flags.N = true
}

func (z *Z80Type) doLdd() {
	z.Flags.N = false
	z.Flags.H = false

	hl := (uint16(z.H) << 8) | uint16(z.L)
	de := (uint16(z.D) << 8) | uint16(z.E)

	readValue := z.core.MemRead(hl)
	z.core.MemWrite(de, readValue)

	z.decDe()
	z.decHl()
	z.decBc()

	z.Flags.P = (z.C | z.B) != 0
	z.Flags.Y = ((z.A+readValue)&0x02)>>1 != 0
	z.Flags.X = ((z.A+readValue)&0x08)>>3 != 0
}

func (z *Z80Type) doCpd() {
	tempCarry := z.Flags.C
	hl := uint16(z.L) | (uint16(z.H) << 8)
	readValue := z.core.MemRead(hl)
	z.doCp(readValue)

	z.Flags.C = tempCarry

	var fh byte = 0
	if z.Flags.H {
		fh = 1
	}

	z.Flags.Y = ((z.A-readValue-fh)&0x02)>>1 != 0
	z.Flags.X = ((z.A-readValue-fh)&0x08)>>3 != 0

	z.decHl()
	z.decBc()

	z.Flags.P = (z.B | z.C) != 0
}

func (z *Z80Type) doInd() {
	hl := (uint16(z.H) << 8) | uint16(z.L)
	bc := (uint16(z.B) << 8) | uint16(z.C)
	z.core.MemWrite(hl, z.core.IORead(bc))
	z.decHl()
	z.B = z.doDec(z.B)
	z.Flags.N = true
}

func (z *Z80Type) doOutd() {
	// Zilog pseudo code is wrong, see: https://github.com/maziac/z80-instruction-set/pull/10
	z.B = z.doDec(z.B)
	hl := (uint16(z.H) << 8) | uint16(z.L)
	bc := (uint16(z.B) << 8) | uint16(z.C)
	z.core.IOWrite(bc, z.core.MemRead(hl))
	z.decHl()
	z.Flags.N = true
}

type OpShift func(operand byte) byte

func (z *Z80Type) doRlc(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	var fc byte = 0
	if z.Flags.C {
		fc = 1
	}
	operand = (operand << 1) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doRrc(operand byte) byte {
	z.Flags.C = operand&1 != 0
	var fc byte = 0
	if z.Flags.C {
		fc = 0x80
	}
	operand = ((operand >> 1) & 0x7f) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doRl(operand byte) byte {
	var fc byte = 0
	if z.Flags.C {
		fc = 1
	}
	z.Flags.C = operand&0x80 != 0
	operand = (operand << 1) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doRr(operand byte) byte {
	var fc byte = 0
	if z.Flags.C {
		fc = 0x80
	}
	z.Flags.C = operand&1 != 0
	operand = ((operand >> 1) & 0x7f) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doSla(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	operand = operand << 1
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doSra(operand byte) byte {
	z.Flags.C = operand&1 != 0
	operand = ((operand >> 1) & 0x7f) | (operand & 0x80)
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doSll(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	operand = (operand << 1) | 1
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80Type) doSrl(operand byte) byte {
	z.Flags.C = operand&1 != 0
	operand = (operand >> 1) & 0x7f
	z.setShiftFlags(operand)
	z.Flags.S = false
	return operand
}

func (z *Z80Type) doIxAdd(operand uint16) {
	z.Flags.N = false
	result := uint32(z.IX) + uint32(operand)
	z.Flags.C = result > 0xffff
	z.Flags.H = (((z.IX & 0xfff) + (operand & 0xfff)) & 0x1000) != 0

	z.updateXYFlags(byte((result & 0xff00) >> 8))
	z.IX = uint16(result & 0xffff)
}

func (z *Z80Type) setShiftFlags(operand byte) {
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.Z = operand == 0
	z.Flags.P = ParityBits[operand]
	z.Flags.S = (operand & 0x80) != 0
	z.updateXYFlags(operand)
}

//type Operation func()

func (z *Z80Type) opcodeFD() {
	z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)
	z.PC++
	opcode := z.core.M1MemRead(z.PC)
	fun := ddInstructions[opcode]
	if fun != nil {
		// Rather than copy and paste all the IX instructions into IY instructions,
		//  what we'll do is sneakily copy IY into IX, run the IX instruction,
		//  and then copy the result into IY and restore the old IX.
		var temp = z.IX
		z.IX = z.IY
		fun(z)
		z.IY = z.IX
		z.IX = temp
		z.CycleCounter += CycleCountsDd[opcode]
	} else {
		z.PC--
		z.CycleCounter += CycleCounts[0]
	}
}

// ============== get register pairs

func (z *Z80Type) bc() uint16 {
	return (uint16(z.B) << 8) | uint16(z.C)
}

func (z *Z80Type) de() uint16 {
	return (uint16(z.D) << 8) | uint16(z.E)
}

func (z *Z80Type) hl() uint16 {
	return (uint16(z.H) << 8) | uint16(z.L)
}

// ============ helper fn

func (z *Z80Type) incBc() {
	z.changeBc(+1)
}

func (z *Z80Type) decBc() {
	z.changeBc(-1)
}

func (z *Z80Type) incHl() {
	z.changeHl(+1)
}

func (z *Z80Type) decHl() {
	z.changeHl(-1)
}

func (z *Z80Type) incDe() {
	z.changeDe(+1)
}

func (z *Z80Type) decDe() {
	z.changeDe(-1)
}

func (z *Z80Type) changeHl(val int8) {
	hl := (uint16(z.H) << 8) | uint16(z.L)
	if val < 0 {
		hl--
	} else {
		hl++
	}
	z.L = byte(hl & 0xff)
	z.H = byte((hl & 0xff00) >> 8)
}

func (z *Z80Type) changeDe(val int8) {
	de := (uint16(z.D) << 8) | uint16(z.E)
	if val < 0 {
		de--
	} else {
		de++
	}
	z.E = byte(de & 0xff)
	z.D = byte((de & 0xff00) >> 8)
}

func (z *Z80Type) changeBc(val int8) {
	bc := (uint16(z.B) << 8) | uint16(z.C)
	if val < 0 {
		bc--
	} else {
		bc++
	}
	z.C = byte(bc & 0x00ff)
	z.B = byte((bc & 0xff00) >> 8)
}
