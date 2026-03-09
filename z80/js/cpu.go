package js

import (
	"fmt"
	"okemu/z80"

	log "github.com/sirupsen/logrus"
)

// const SpDefault uint16 = 0xffff
type Z80 struct {
	z80.Z80CPU
	core z80.MemIoRW
}

func (z *Z80) Reset() {
	z.A = 0
	z.R = 0
	z.SP = 0xff
	z.PC = 0
	z.setFlagsRegister(0xff)
	// Interrupts disabled
	z.IMode = 0
	z.Iff1 = false
	z.Iff2 = false
	z.InterruptOccurred = false

	// Start not halted
	z.Halted = false
	z.DoDelayedDI = false
	z.DoDelayedEI = false
	// no cycles
	z.CycleCounter = 0
	fmt.Println("CPUInterface State Reset")
}

func (z *Z80) GetState() *z80.Z80CPU {
	return &z80.Z80CPU{
		A:            z.A,
		B:            z.B,
		C:            z.C,
		D:            z.D,
		E:            z.E,
		H:            z.H,
		L:            z.L,
		AAlt:         z.AAlt,
		BAlt:         z.BAlt,
		CAlt:         z.CAlt,
		DAlt:         z.DAlt,
		EAlt:         z.EAlt,
		HAlt:         z.HAlt,
		IX:           z.IX,
		IY:           z.IY,
		R:            z.R,
		SP:           z.SP,
		PC:           z.PC,
		Flags:        z.Flags,
		FlagsAlt:     z.FlagsAlt,
		IMode:        z.IMode,
		Iff1:         z.Iff1,
		Iff2:         z.Iff2,
		Halted:       z.Halted,
		DoDelayedDI:  z.DoDelayedDI,
		DoDelayedEI:  z.DoDelayedEI,
		CycleCounter: z.CycleCounter,
	}
}

func (z *Z80) SetState(state *Z80) {
	z.A = state.A
	z.B = state.B
	z.C = state.C
	z.D = state.D
	z.E = state.E
	z.H = state.H
	z.L = state.L
	z.AAlt = state.AAlt
	z.BAlt = state.BAlt
	z.CAlt = state.CAlt
	z.DAlt = state.DAlt
	z.EAlt = state.EAlt
	z.HAlt = state.HAlt
	z.IX = state.IX
	z.IY = state.IY
	z.I = state.I
	z.R = state.R
	z.SP = state.SP
	z.PC = state.PC
	z.Flags = state.Flags
	z.FlagsAlt = state.FlagsAlt
	z.IMode = state.IMode
	z.Iff1 = state.Iff1
	z.Iff2 = state.Iff2
	z.Halted = state.Halted
	z.DoDelayedDI = state.DoDelayedDI
	z.DoDelayedEI = state.DoDelayedEI
	z.CycleCounter = state.CycleCounter
}

// New Create new
func New(memIoRW z80.MemIoRW) *Z80 {
	return &Z80{
		Z80CPU: z80.Z80CPU{
			A:    0,
			B:    0,
			C:    0,
			D:    0,
			E:    0,
			H:    0,
			L:    0,
			AAlt: 0,
			BAlt: 0,
			CAlt: 0,
			DAlt: 0,
			EAlt: 0,
			HAlt: 0,
			IX:   0,
			IY:   0,
			I:    0,

			R:                 0,
			SP:                0xffff,
			PC:                0,
			Flags:             z80.FlagsType{S: true, Z: true, Y: true, H: true, X: true, P: true, N: true, C: true},
			FlagsAlt:          z80.FlagsType{S: true, Z: true, Y: true, H: true, X: true, P: true, N: true, C: true},
			IMode:             0,
			Iff1:              false,
			Iff2:              false,
			Halted:            false,
			DoDelayedDI:       false,
			DoDelayedEI:       false,
			CycleCounter:      0,
			InterruptOccurred: false,
		},
		core: memIoRW,
	}
}

func (z *Z80) RunInstruction() uint64 {

	z.incR()

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
			z.Iff1 = false
			z.Iff2 = false
		} else if doingDelayedEi {
			z.Iff1 = true
			z.Iff2 = true
		}

		// And finally clear out the cycle counter for the next instruction
		//  before returning it to the emulator core.
		cycleCounter := z.CycleCounter
		z.CycleCounter = 0
		return uint64(cycleCounter)
	}

	// HALTED
	// During HALT, NOPs are executed which is 4T
	z.core.M1MemRead(z.PC) // HALT does a normal M1 fetch to keep the memory refresh active. The result is ignored (NOP).
	return 4

}

// Simulates pulsing the processor's INT (or NMI) pin
//
//	nonMaskable - true if this is a non-maskable interrupt
//	data - the value to be placed on the data bus, if needed
func (z *Z80) interrupt(nonMaskable bool, data byte) {
	if nonMaskable {
		// An interrupt, if halted, does increase the PC
		if z.Halted {
			z.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R, A instruction.
		z.incR()

		// Non-maskable interrupts are always handled the same way;
		// clear IFF1 and then do a CALL 0x0066.
		// Also, all interrupts reset the HALT state.

		z.Halted = false
		z.Iff2 = z.Iff1
		z.Iff1 = false
		z.pushWord(z.PC)
		z.PC = 0x66

		z.CycleCounter += 11
	} else if z.Iff1 {
		//  An interrupt, if halted, does increase the PC
		if z.Halted {
			z.PC++
		}

		// The high bit of R is not affected by this increment,
		//  it can only be changed using the LD R,A instruction.
		z.incR()

		z.Halted = false
		z.Iff1 = false
		z.Iff2 = false

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
			z.PC = z.getWord(vectorAddress) //uint16( z.core.MemRead(vectorAddress)) | (uint16(z.core.MemRead(vectorAddress+1)) << 8)
			z.CycleCounter += 19
			// A "notification" is generated so that the calling program can break on it.
			z.InterruptOccurred = true
		}
	}
}

func (z *Z80) pushWord(operand uint16) {
	z.SP--
	z.core.MemWrite(z.SP, byte(operand>>8))
	z.SP--
	z.core.MemWrite(z.SP, byte(operand&0x00ff))
}

func (z *Z80) getOperand(opcode byte) byte {
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
		return z.core.MemRead(z.hl())
	default:
		return z.A
	}
}

func (z *Z80) decodeInstruction(opcode byte) {
	// Handle HALT right up front, because it fouls up our LD decoding
	//  by falling where LD (HL), (HL) ought to be.
	if opcode == OpHalt {
		z.Halted = true
	} else if opcode >= OpLdBB && opcode < OpAddAB {
		// 8-bit register loads.
		z.load8bit(opcode, z.getOperand(opcode))
	} else if (opcode >= OpAddAB) && (opcode < OpRetNz) {
		// 8-bit register ALU instructions.
		z.alu8bit(opcode, z.getOperand(opcode))
	} else {
		fun := instructions[opcode]
		fun(z)
	}
	z.CycleCounter += CycleCounts[opcode]
}

func (z *Z80) load8bit(opcode byte, operand byte) {
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
		z.core.MemWrite(z.hl(), operand)
	default:
		z.A = operand
	}
}

// alu8bit Handle ALU Operations, ADD, ADC SUB, SBC, AND, XOR, OR
func (z *Z80) alu8bit(opcode byte, operand byte) {
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

// getFlagsRegister return whole F register
func (z *Z80) getFlagsRegister() byte {
	return getFlags(&z.Flags)
}

// getFlagsRegister return whole F' register
func (z *Z80) getFlagsPrimeRegister() byte {
	return getFlags(&z.FlagsAlt)
}

func getFlags(f *z80.FlagsType) byte {
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

func (z *Z80) setFlagsRegister(flags byte) {
	setFlags(flags, &z.Flags)
}

func (z *Z80) setFlagsPrimeRegister(flags byte) {
	setFlags(flags, &z.FlagsAlt)
}

func setFlags(flags byte, f *z80.FlagsType) {
	f.S = flags&0x80 != 0
	f.Z = flags&0x40 != 0
	f.Y = flags&0x20 != 0
	f.H = flags&0x10 != 0
	f.X = flags&0x08 != 0
	f.P = flags&0x04 != 0
	f.N = flags&0x02 != 0
	f.C = flags&0x01 != 0
}

// updateXYFlags Set flags X and Y based on result bits
func (z *Z80) updateXYFlags(result byte) {
	z.Flags.Y = result&0x20 != 0
	z.Flags.X = result&0x08 != 0
}

// PushWord - Decrement stack pointer and put specified word value to stack
func (z *Z80) PushWord(operand uint16) {
	z.SP--
	z.core.MemWrite(z.SP, byte((operand&0xff00)>>8))
	z.SP--
	z.core.MemWrite(z.SP, byte(operand&0x00ff))
}

// PopWord - Get word value from top of stack and increment stack pointer
func (z *Z80) PopWord() uint16 {
	result := uint16(z.core.MemRead(z.SP))
	z.SP++
	result |= uint16(z.core.MemRead(z.SP)) << 8
	z.SP++
	return result
}

// doConditionalAbsoluteJump - Implements the JP [condition],nn instructions.
func (z *Z80) doConditionalAbsoluteJump(condition bool) {
	if condition {
		// We're taking this jump, so write the new PC,
		//  and then decrement the thing we just wrote,
		//  because the instruction decoder increments the PC
		//  unconditionally at the end of every instruction,
		//  and we need to counteract that so we end up at the jump target.
		z.PC = z.getWord(z.PC + 1) //uint16( z.core.MemRead(z.PC+1)) | (uint16(z.core.MemRead(z.PC+2)) << 8)
		z.PC--
	} else {
		// We're not taking this jump, just move the PC past the operand.
		z.PC += 2
	}
}

// doConditionalRelativeJump - Implements the JR [condition],nn instructions.
func (z *Z80) doConditionalRelativeJump(condition bool) {
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
func (z *Z80) doConditionalCall(condition bool) {
	if condition {
		z.CycleCounter += 7
		z.PushWord(z.PC + 3)
		z.PC = z.getWord(z.PC + 1) // uint16( z.core.MemRead(z.PC+1)) | (uint16(z.core.MemRead(z.PC+2)) << 8)
		z.PC--
	} else {
		z.PC += 2
	}
}

func (z *Z80) doConditionalReturn(condition bool) {
	if condition {
		z.CycleCounter += 6
		z.PC = z.PopWord() - 1
	}
}

// doReset - Implements RST [address] instructions.
func (z *Z80) doReset(address uint16) {
	z.PushWord(z.PC + 1)
	z.PC = address - 1
}

// doAdd Handle ADD A, [operand] instructions.
func (z *Z80) doAdd(operand byte) {
	var result = uint16(z.A) + uint16(operand)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result&0x00ff == 0
	z.Flags.H = (((operand & 0x0f) + (z.A & 0x0f)) & 0x10) != 0
	z.Flags.P = ((z.A & 0x80) == (operand & 0x80)) && (z.A&0x80 != byte(result&0x80))
	z.Flags.N = false
	z.Flags.C = result&0x0100 != 0

	z.A = byte(result & 0xff)
	z.updateXYFlags(z.A)
}

// doAdc Handle ADC A, [operand] instructions.
func (z *Z80) doAdc(operand byte) {
	add := byte(0)
	if z.Flags.C {
		add = 1
	}
	var result = uint16(z.A) + uint16(operand) + uint16(add)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result&0x00ff == 0
	z.Flags.H = (((operand & 0x0f) + (z.A & 0x0f) + add) & 0x10) != 0
	z.Flags.P = ((z.A & 0x80) == (operand & 0x80)) && (z.A&0x80 != byte(result&0x80))
	z.Flags.N = false
	z.Flags.C = result&0x0100 != 0

	z.A = byte(result & 0xff)
	z.updateXYFlags(z.A)
}

// doSub Handle SUB A, [operand] instructions.
func (z *Z80) doSub(operand byte) {
	var result = uint16(z.A) - uint16(operand)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result&0x00ff == 0
	z.Flags.H = (((z.A & 0x0f) - (operand & 0x0f)) & 0x10) != 0
	z.Flags.P = ((z.A & 0x80) != (operand & 0x80)) && ((z.A & 0x80) != byte(result&0x80))
	z.Flags.N = true
	z.Flags.C = result&0x0100 != 0

	z.A = byte(result & 0xff)
	z.updateXYFlags(z.A)
}

// doSbc Handle SBC A, [operand] instructions.
func (z *Z80) doSbc(operand byte) {
	sub := byte(0)
	if z.Flags.C {
		sub = 1
	}
	var result = uint16(z.A) - uint16(operand) - uint16(sub)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result&0x00ff == 0
	z.Flags.H = (((z.A & 0x0f) - (operand & 0x0f) - sub) & 0x10) != 0
	z.Flags.P = ((z.A & 0x80) != (operand & 0x80)) && (z.A&0x80 != byte(result&0x80))
	z.Flags.N = true
	z.Flags.C = result&0x0100 != 0

	z.A = byte(result & 0xff)
	z.updateXYFlags(z.A)
}

// setLogicFlags Set common flags for logic ALU Ops
func (z *Z80) setLogicFlags() {
	z.Flags.S = z.A&0x80 != 0
	z.Flags.Z = z.A == 0
	z.Flags.P = ParityBits[z.A]
	z.Flags.N = false
	z.Flags.C = false
}

// doAnd handle AND [operand] instructions.
func (z *Z80) doAnd(operand byte) {
	z.A &= operand
	z.setLogicFlags()
	z.Flags.H = true
	z.updateXYFlags(z.A)
}

// doXor handle XOR [operand] instructions.
func (z *Z80) doXor(operand byte) {
	z.A ^= operand
	z.setLogicFlags()
	z.Flags.H = false
	z.updateXYFlags(z.A)
}

// doOr handle OR [operand] instructions.
func (z *Z80) doOr(operand byte) {
	z.A |= operand
	z.setLogicFlags()
	z.Flags.H = false
	z.updateXYFlags(z.A)
}

// doCp handle CP [operand] instructions.
func (z *Z80) doCp(operand byte) {
	tmp := z.A
	z.doSub(operand)
	z.A = tmp
	z.updateXYFlags(operand)
}

// doInc handle INC [operand] instructions.
func (z *Z80) doInc(operand byte) byte {
	var result = uint16(operand) + 1
	r8 := byte(result & 0xff)

	z.Flags.S = r8&0x80 != 0
	z.Flags.Z = r8 == 0
	z.Flags.H = (operand & 0x0f) == 0x0f
	z.Flags.P = operand == 0x7f
	z.Flags.N = false

	z.updateXYFlags(r8)
	return r8
}

// doDec handle DEC [operand] instructions.
func (z *Z80) doDec(operand byte) byte {
	var result = uint16(operand) - 1
	r8 := byte(result & 0xff)

	z.Flags.S = r8&0x80 != 0
	z.Flags.Z = r8 == 0
	z.Flags.H = (operand & 0x0f) == 0x00
	z.Flags.P = operand == 0x80
	z.Flags.N = true

	z.updateXYFlags(r8)
	return r8
}

// doHlAdd handle ADD HL,[operand] instructions.
func (z *Z80) doHlAdd(operand uint16) {
	// The HL arithmetic instructions are the same as the A ones,
	//  just with twice as many bits happening.
	hl := z.hl() //uint16(z.L) | (uint16(z.H) << 8)
	result := uint32(hl) + uint32(operand)
	z.Flags.N = false
	z.Flags.C = result > 0xffff
	z.Flags.H = ((hl&0x0fff)+(operand&0x0fff))&0x1000 > 0

	z.setHl(uint16(result))
	//z.L = byte(result & 0xff)
	//z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

// doHlAdc handle ADC HL,[operand] instructions.
func (z *Z80) doHlAdc(operand uint16) {
	if z.Flags.C {
		operand++
	}

	hl := z.hl()
	result := uint32(hl) + uint32(operand)

	z.Flags.S = (result & 0x8000) != 0
	z.Flags.Z = result&0xffff == 0
	z.Flags.H = (((hl & 0x0fff) + (operand & 0x0fff)) & 0x1000) != 0
	z.Flags.P = ((hl & 0x8000) == (operand & 0x8000)) && (uint16(result&0x8000) != (hl & 0x8000))
	z.Flags.N = false
	z.Flags.C = result > 0xffff

	z.setHl(uint16(result))
	//z.L = byte(result & 0xff)
	//z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

// doHlSbc handle SBC HL,[operand] instructions.
func (z *Z80) doHlSbc(operand uint16) {
	if z.Flags.C {
		operand++
	}

	hl := z.hl() //uint16(z.L) | (uint16(z.H) << 8)
	result := uint32(hl) - uint32(operand)

	z.Flags.S = (result & 0x8000) != 0
	z.Flags.Z = result&0xffff == 0
	z.Flags.H = (((hl & 0x0fff) - (operand & 0x0fff)) & 0x1000) != 0
	z.Flags.P = ((hl & 0x8000) != (operand & 0x8000)) && (uint16(result&0x8000) != (hl & 0x8000))
	z.Flags.N = true
	z.Flags.C = result > 0xffff

	z.setHl(uint16(result))
	//z.L = byte(result & 0xff)
	//z.H = byte((result & 0xff00) >> 8)

	z.updateXYFlags(z.H)
}

func (z *Z80) doIn(port uint16) byte {
	var result = z.core.IORead(port)

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = false
	z.Flags.P = ParityBits[result]
	z.Flags.N = false
	z.updateXYFlags(result)
	return result
}

func (z *Z80) doNeg() {
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

func (z *Z80) doLdi() {
	// Copy the value that we're supposed to copy.
	readValue := z.core.MemRead(z.hl())
	z.core.MemWrite(z.de(), readValue)

	z.incDe()
	z.incHl()
	z.decBc()

	z.Flags.H = false
	z.Flags.P = (z.C | z.B) != 0
	z.Flags.N = false
	z.Flags.Y = ((z.A+readValue)&0x02)>>1 != 0
	z.Flags.X = ((z.A+readValue)&0x08)>>3 != 0
}

func (z *Z80) fhv() byte {
	if z.Flags.H {
		return 1
	}
	return 0
}

func (z *Z80) doCpi() {
	tempCarry := z.Flags.C
	readValue := z.core.MemRead(z.hl())
	z.doCp(readValue)

	z.Flags.C = tempCarry
	fh := z.fhv()
	z.Flags.Y = ((z.A - readValue - fh) & 0x02) != 0
	z.Flags.X = ((z.A - readValue - fh) & 0x08) != 0
	z.incHl()
	z.decBc()
	z.Flags.P = (z.B | z.C) != 0
}

func (z *Z80) doIni() {
	z.core.MemWrite(z.hl(), z.core.IORead(z.bc()))
	z.incHl()
	z.B = z.doDec(z.B)
	z.Flags.N = true
}

func (z *Z80) doOuti() {
	z.B = z.doDec(z.B)
	z.core.IOWrite(z.bc(), z.core.MemRead(z.hl()))
	z.incHl()
	z.Flags.N = true
}

func (z *Z80) doLdd() {
	z.Flags.N = false
	z.Flags.H = false
	readValue := z.core.MemRead(z.hl())
	z.core.MemWrite(z.de(), readValue)
	z.decDe()
	z.decHl()
	z.decBc()
	z.Flags.P = (z.C | z.B) != 0
	z.Flags.Y = ((z.A + readValue) & 0x02) != 0
	z.Flags.X = ((z.A + readValue) & 0x08) != 0
}

func (z *Z80) doCpd() {
	tempCarry := z.Flags.C
	readValue := z.core.MemRead(z.hl())
	z.doCp(readValue)

	z.Flags.C = tempCarry

	fh := z.fhv()

	z.Flags.Y = ((z.A-readValue-fh)&0x02)>>1 != 0
	z.Flags.X = ((z.A-readValue-fh)&0x08)>>3 != 0

	z.decHl()
	z.decBc()

	z.Flags.P = (z.B | z.C) != 0
}

func (z *Z80) doInd() {
	z.core.MemWrite(z.hl(), z.core.IORead(z.bc()))
	z.decHl()
	z.B = z.doDec(z.B)
	z.Flags.N = true
}

func (z *Z80) doOutd() {
	z.B = z.doDec(z.B)
	z.core.IOWrite(z.bc(), z.core.MemRead(z.hl()))
	z.decHl()
	z.Flags.N = true
}

type OpShift func(operand byte) byte

func (z *Z80) doRlc(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	var fc byte = 0
	if z.Flags.C {
		fc = 1
	}
	operand = (operand << 1) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doRrc(operand byte) byte {
	z.Flags.C = operand&1 != 0
	var fc byte = 0
	if z.Flags.C {
		fc = 0x80
	}
	operand = ((operand >> 1) & 0x7f) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doRl(operand byte) byte {
	var fc byte = 0
	if z.Flags.C {
		fc = 1
	}
	z.Flags.C = operand&0x80 != 0
	operand = (operand << 1) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doRr(operand byte) byte {
	var fc byte = 0
	if z.Flags.C {
		fc = 0x80
	}
	z.Flags.C = operand&1 != 0
	operand = ((operand >> 1) & 0x7f) | fc
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doSla(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	operand = operand << 1
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doSra(operand byte) byte {
	z.Flags.C = operand&1 != 0
	operand = ((operand >> 1) & 0x7f) | (operand & 0x80)
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doSll(operand byte) byte {
	z.Flags.C = operand&0x80 != 0
	operand = (operand << 1) | 1
	z.setShiftFlags(operand)
	return operand
}

func (z *Z80) doSrl(operand byte) byte {
	z.Flags.C = operand&1 != 0
	operand = (operand >> 1) & 0x7f
	z.setShiftFlags(operand)
	z.Flags.S = false
	return operand
}

func (z *Z80) doIxAdd(operand uint16) {
	z.Flags.N = false
	result := uint32(z.IX) + uint32(operand)
	z.Flags.C = result > 0xffff
	z.Flags.H = (((z.IX & 0xfff) + (operand & 0xfff)) & 0x1000) != 0

	z.updateXYFlags(byte((result & 0xff00) >> 8))
	z.IX = uint16(result & 0xffff)
}

func (z *Z80) setShiftFlags(operand byte) {
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.Z = operand == 0
	z.Flags.P = ParityBits[operand]
	z.Flags.S = (operand & 0x80) != 0
	z.updateXYFlags(operand)
}

// ============== get register pairs

func (z *Z80) bc() uint16 {
	return (uint16(z.B) << 8) | uint16(z.C)
}

func (z *Z80) de() uint16 {
	return (uint16(z.D) << 8) | uint16(z.E)
}

func (z *Z80) hl() uint16 {
	return (uint16(z.H) << 8) | uint16(z.L)
}

// ============ helper fn

func (z *Z80) incBc() {
	z.setBc(z.bc() + 1)
}

func (z *Z80) decBc() {
	z.setBc(z.bc() - 1)
}

func (z *Z80) incDe() {
	z.setDe(z.de() + 1)
}

func (z *Z80) decDe() {
	z.setDe(z.de() - 1)
}

func (z *Z80) incHl() {
	z.setHl(z.hl() + 1)
}

func (z *Z80) decHl() {
	z.setHl(z.hl() - 1)
}

func (z *Z80) setHl(val uint16) {
	z.L = byte(val & 0xff)
	z.H = byte(val >> 8)
}

func (z *Z80) setDe(val uint16) {
	z.E = byte(val & 0xff)
	z.D = byte(val >> 8)
}

func (z *Z80) setBc(val uint16) {
	z.C = byte(val & 0xff)
	z.B = byte(val >> 8)
}

// incR Increment R at the start of every instruction cycle.
// The high bit of R is not affected by this increment,
// it can only be changed using the LD R, A instruction.
// Note: also a HALT does increment the R register.
func (z *Z80) incR() {
	z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)
}

// getWord Return 16bit value from memory by specified address
func (z *Z80) getWord(address uint16) uint16 {
	return (uint16(z.core.MemRead(address+1)) << 8) | uint16(z.core.MemRead(address))
}

func (z *Z80) setWord(address uint16, value uint16) {
	z.core.MemWrite(address, byte(value))
	z.core.MemWrite(address+1, byte(value>>8))
}

func (z *Z80) debugOutput() {
	log.Debugf("PC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X, IX: %04X, IY: %04X, I: %02X, R: %02X",
		z.PC, (uint16(z.A)<<8)|uint16(z.getFlagsRegister()), z.bc(), z.de(), z.hl(), z.SP,
		z.IX, z.IY, z.I, z.R)

	log.Debugf("\t(%02X %02X %02X %02X), cyc: %d\n", z.core.MemRead(z.PC), z.core.MemRead(z.PC+1),
		z.core.MemRead(z.PC+2), z.core.MemRead(z.PC+3), z.CycleCounter)
}
