package z80em

var ddInstructions = []func(s *Z80Type){
	// 0x09 : ADD IX, BC
	0x09: func(s *Z80Type) {
		s.doIxAdd(s.bc())
	},
	// 0x19 : ADD IX, DE
	0x19: func(s *Z80Type) {
		s.doIxAdd(s.de())
	},
	// 0x21 : LD IX, nn
	0x21: func(s *Z80Type) {
		s.IX = s.getAddr()
	},
	// 0x22 : LD (nn), IX
	0x22: func(s *Z80Type) {
		addr := s.getAddr()
		s.core.MemWrite(addr, byte(s.IX&0x00ff))
		s.core.MemWrite(addr+1, byte(s.IX>>8))
	},
	// 0x23 : INC IX
	0x23: func(s *Z80Type) {
		s.IX++
	},
	// 0x24 : INC IXH (Undocumented)
	0x24: func(s *Z80Type) {
		s.IX = (uint16(s.doInc(byte(s.IX>>8))) << 8) | (s.IX & 0x00ff)
	},
	// 0x25 : DEC IXH (Undocumented)
	0x25: func(s *Z80Type) {
		s.IX = (uint16(s.doDec(byte(s.IX>>8))) << 8) | (s.IX & 0x00ff)
	},
	// 0x26 : LD IXH, n (Undocumented)
	0x26: func(s *Z80Type) {
		s.PC++
		s.IX = (uint16(s.core.MemRead(s.PC)) << 8) | (s.IX & 0x00ff)
	},
	// 0x29 : ADD IX, IX
	0x29: func(s *Z80Type) {
		s.doIxAdd(s.IX)
	},
	// 0x2a : LD IX, (nn)
	0x2A: func(s *Z80Type) {
		addr := s.getAddr()
		s.IX = (uint16(s.core.MemRead(addr)) << 8) | uint16(s.core.MemRead(addr+1))
	},
	// 0x2b : DEC IX
	0x2B: func(s *Z80Type) {
		s.IX--
	},
	// 0x2c : INC IXL (Undocumented)
	0x2C: func(s *Z80Type) {
		s.IX = (uint16(s.doInc(byte(s.IX & 0x00ff)))) | (s.IX & 0xff00)
	},
	// 0x2d : DEC IXL (Undocumented)
	0x2D: func(s *Z80Type) {
		s.IX = (uint16(s.doDec(byte(s.IX & 0x00ff)))) | (s.IX & 0xff00)
	},
	// 0x2e : LD IXL, n (Undocumented)
	0x2E: func(s *Z80Type) {
		s.PC++
		s.IX = (uint16(s.core.MemRead(s.PC))) | (s.IX & 0xff00)
	},
	// 0x34 : INC (IX+n)
	0x34: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		value := s.core.MemRead(offset)
		s.core.MemWrite(offset, s.doInc(value))
	},
	// 0x35 : DEC (IX+n)
	0x35: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		value := s.core.MemRead(offset)
		s.core.MemWrite(offset, s.doDec(value))
	},
	// 0x36 : LD (IX+n), n
	0x36: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.PC++
		s.core.MemWrite(offset, s.core.MemRead(s.PC))
	},
	// 0x39 : ADD IX, SP
	0x39: func(s *Z80Type) {
		s.doIxAdd(s.SP)
	},
	// 0x44 : LD B, IXH (Undocumented)
	0x44: func(s *Z80Type) {
		s.B = byte(s.IX >> 8)
	},
	// 0x45 : LD B, IXL (Undocumented)
	0x45: func(s *Z80Type) {
		s.B = byte(s.IX & 0x00ff)
	},
	// 0x46 : LD B, (IX+n)
	0x46: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.B = s.core.MemRead(offset)
	},
	// 0x4c : LD C, IXH (Undocumented)
	0x4C: func(s *Z80Type) {
		s.C = byte(s.IX >> 8)
	},
	// 0x4d : LD C, IXL (Undocumented)
	0x4D: func(s *Z80Type) {
		s.C = byte(s.IX & 0x00ff)
	},
	// 0x4e : LD C, (IX+n)
	0x4E: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.C = s.core.MemRead(offset)
	},
	// 0x54 : LD D, IXH (Undocumented)
	0x54: func(s *Z80Type) {
		s.D = byte(s.IX >> 8)
	},
	// 0x55 : LD D, IXL (Undocumented)
	0x55: func(s *Z80Type) {
		s.D = byte(s.IX & 0x00ff)
	},
	// 0x56 : LD D, (IX+n)
	0x56: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.D = s.core.MemRead(offset)
	},
	// 0x5d : LD E, IXL (Undocumented)
	0x5D: func(s *Z80Type) {
		s.E = byte(s.IX & 0x00ff)
	},
	// 0x5e : LD E, (IX+n)
	0x5E: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.E = s.core.MemRead(offset)
	},
	// 0x60 : LD IXH, B (Undocumented)
	0x60: func(s *Z80Type) {
		s.IX = uint16(s.B)<<8 | s.IX&0x00ff
	},
	// 0x61 : LD IXH, C (Undocumented)
	0x61: func(s *Z80Type) {
		s.IX = uint16(s.C)<<8 | s.IX&0x00ff
	},
	// 0x62 : LD IXH, D (Undocumented)
	0x62: func(s *Z80Type) {
		s.IX = uint16(s.D)<<8 | s.IX&0x00ff
	},
	// 0x63 : LD IXH, E (Undocumented)
	0x63: func(s *Z80Type) {
		s.IX = uint16(s.E)<<8 | s.IX&0x00ff
	},
	// 0x64 : LD IXH, IXH (Undocumented)
	0x64: func(s *Z80Type) {
		// NOP
	},
	// 0x65 : LD IXH, IXL (Undocumented)
	0x65: func(s *Z80Type) {
		s.IX = (s.IX << 8) | (s.IX & 0x00ff)
	},
	// 0x66 : LD H, (IX+n)
	0x66: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.H = s.core.MemRead(offset)
	},
	// 0x67 : LD IXH, A (Undocumented)
	0x67: func(s *Z80Type) {
		s.IX = (uint16(s.A) << 8) | (s.IX & 0x00ff)
	},
	// 0x68 : LD IXL, B (Undocumented)
	0x68: func(s *Z80Type) {
		s.IX = (s.IX & 0xff00) | uint16(s.B)
	},
	// 0x69 : LD IXL, C (Undocumented)
	0x69: func(s *Z80Type) {
		s.IX = (s.IX & 0xff00) | uint16(s.C)
	},
	// 0x6a : LD IXL, D (Undocumented)
	0x6a: func(s *Z80Type) {
		s.IX = (s.IX & 0xff00) | uint16(s.D)
	},
	// 0x6b : LD IXL, E (Undocumented)
	0x6b: func(s *Z80Type) {
		s.IX = (s.IX & 0xff00) | uint16(s.E)
	},
	// 0x6c : LD IXL, IXH (Undocumented)
	0x6c: func(s *Z80Type) {
		s.IX = (s.IX >> 8) | (s.IX & 0xff00)
	},
	// 0x6d : LD IXL, IXL (Undocumented)
	0x6d: func(s *Z80Type) {
		// NOP
	},
	// 0x6e : LD L, (IX+n)
	0x6e: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.L = s.core.MemRead(offset)
	},
	// 0x6f : LD IXL, A (Undocumented)
	0x6f: func(s *Z80Type) {
		s.IX = uint16(s.A) | (s.IX & 0xff00)
	},
	// 0x70 : LD (IX+n), B
	0x70: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.B)
	},
	// 0x71 : LD (IX+n), C
	0x71: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.C)
	},
	// 0x72 : LD (IX+n), D
	0x72: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.D)
	},
	// 0x73 : LD (IX+n), E
	0x73: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.E)
	},
	// 0x74 : LD (IX+n), H
	0x74: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.H)
	},
	// 0x75 : LD (IX+n), L
	0x75: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.L)
	},
	// 0x77 : LD (IX+n), A
	0x77: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.core.MemWrite(offset, s.A)
	},
	// 0x7c : LD A, IXH (Undocumented)
	0x7C: func(s *Z80Type) {
		s.A = byte(s.IX >> 8)
	},
	// 0x7d : LD A, IXL (Undocumented)
	0x7D: func(s *Z80Type) {
		s.A = byte(s.IX & 0x00ff)
	},
	// 0x7e : LD A, (IX+n)
	0x7E: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.A = s.core.MemRead(offset)
	},
	// 0x84 : ADD A, IXH (Undocumented)
	0x84: func(s *Z80Type) {
		s.doAdd(byte(s.IX >> 8))
	},
	// 0x85 : ADD A, IXL (Undocumented)
	0x85: func(s *Z80Type) {
		s.doAdd(byte(s.IX & 0x00ff))
	},
	// 0x86 : ADD A, (IX+n)
	0x86: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doAdd(s.core.MemRead(offset))
	},
	// 0x8c : ADC A, IXH (Undocumented)
	0x8C: func(s *Z80Type) {
		s.doAdc(byte(s.IX >> 8))
	},
	// 0x8d : ADC A, IXL (Undocumented)
	0x8D: func(s *Z80Type) {
		s.doAdc(byte(s.IX & 0x00ff))
	},
	// 0x8e : ADC A, (IX+n)
	0x8E: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doAdc(s.core.MemRead(offset))
	},
	// 0x94 : SUB IXH (Undocumented)
	0x94: func(s *Z80Type) {
		s.doSub(byte(s.IX >> 8))
	},
	// 0x95 : SUB IXL (Undocumented)
	0x95: func(s *Z80Type) {
		s.doSub(byte(s.IX & 0x00ff))
	},
	// 0x96 : SUB A, (IX+n)
	0x96: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doSub(s.core.MemRead(offset))
	},
	// 0x9c : SBC IXH (Undocumented)
	0x9C: func(s *Z80Type) {
		s.doSbc(byte(s.IX >> 8))
	},
	// 0x9d : SBC IXL (Undocumented)
	0x9D: func(s *Z80Type) {
		s.doSbc(byte(s.IX & 0x00ff))
	},
	// 0x9e : SBC A, (IX+n)
	0x9E: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doSbc(s.core.MemRead(offset))
	},
	// 0xa4 : AND IXH (Undocumented)
	0xA4: func(s *Z80Type) {
		s.doAnd(byte(s.IX >> 8))
	},
	// 0xa5 : AND IXL (Undocumented)
	0xA5: func(s *Z80Type) {
		s.doAnd(byte(s.IX & 0x00ff))
	},
	// 0xa6 : AND A, (IX+n)
	0xA6: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doAnd(s.core.MemRead(offset))
	},
	// 0xac : XOR IXH (Undocumented)
	0xAC: func(s *Z80Type) {
		s.doXor(byte(s.IX >> 8))
	},
	// 0xad : XOR IXL (Undocumented)
	0xAD: func(s *Z80Type) {
		s.doXor(byte(s.IX & 0x00ff))
	},
	// 0xae : XOR A, (IX+n)
	0xAE: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doXor(s.core.MemRead(offset))
	},
	// 0xb4 : OR IXH (Undocumented)
	0xB4: func(s *Z80Type) {
		s.doOr(byte(s.IX >> 8))
	},
	// 0xb5 : OR IXL (Undocumented)
	0xB5: func(s *Z80Type) {
		s.doOr(byte(s.IX & 0x00ff))
	},
	// 0xb6 : OR A, (IX+n)
	0xB6: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doOr(s.core.MemRead(offset))
	},
	// 0xbc : CP IXH (Undocumented)
	0xBC: func(s *Z80Type) {
		s.doCp(byte(s.IX >> 8))
	},
	// 0xbd : CP IXL (Undocumented)
	0xBD: func(s *Z80Type) {
		s.doCp(byte(s.IX & 0x00ff))
	},
	// 0xbe : CP A, (IX+n)
	0xBE: func(s *Z80Type) {
		offset := s.getOffset(s.IX)
		s.doCp(s.core.MemRead(offset))
	},
	// 0xcb : CB Prefix (IX bit instructions)
	0xCB: func(s *Z80Type) {
		s.opcodeDDCB()
	},
	// 0xe1 : POP IX
	0xE1: func(s *Z80Type) {
		s.IX = s.PopWord()
	},
	// 0xe3 : EX (SP), IX
	0xE3: func(s *Z80Type) {
		temp := s.IX
		s.IX = uint16(s.core.MemRead(s.SP))
		s.IX |= uint16(s.core.MemRead(s.SP+1)) << 8
		s.core.MemWrite(s.SP, byte(temp&0x00ff))
		s.core.MemWrite(s.SP+1, byte(temp>>8))
	},
	// 0xe5 : PUSH IX
	0xE5: func(s *Z80Type) {
		s.pushWord(s.IX)
	},
	// 0xe9 : JP (IX)
	0xE9: func(s *Z80Type) {
		s.PC = s.IX - 1
	},
	// 0xf9 : LD SP, IX
	0xf9: func(s *Z80Type) {
		s.SP = s.IX
	},
}

// =====================================================
func (z *Z80Type) getOffset(reg uint16) uint16 {
	z.PC++
	offset := z.core.MemRead(z.PC)
	if offset < 0 {
		reg -= uint16(-offset)
	} else {
		reg += uint16(offset)
	}
	return reg
}

func (z *Z80Type) opcodeDD() {
	z.R = (z.R & 0x80) | (((z.R & 0x7f) + 1) & 0x7f)
	z.PC++
	opcode := z.core.M1MemRead(z.PC)

	fun := ddInstructions[opcode]
	if fun != nil {
		//func = func.bind(this);
		fun(z)
		z.CycleCounter += CycleCountsDd[opcode]
	} else {
		// Apparently if a DD opcode doesn't exist,
		// it gets treated as an unprefixed opcode.
		// What we'll do to handle that is just back up the
		// program counter, so that this byte gets decoded
		// as a normal instruction.
		z.PC--
		// And we'll add in the cycle count for a NOP.
		z.CycleCounter += CycleCounts[0]
	}
}

func (z *Z80Type) opcodeDDCB() {

	offset := z.getOffset(z.IX)
	z.PC++

	opcode := z.core.MemRead(z.PC)

	value := byte(0)
	bitTestOp := false

	// As with the "normal" CB prefix, we implement the DDCB prefix
	//  by decoding the opcode directly, rather than using a table.
	if opcode < 0x40 {
		// Shift and rotate instructions.
		ddcbFunctions := []OpShift{z.doRlc, z.doRrc, z.doRl, z.doRr, z.doSla, z.doSra, z.doSll, z.doSrl}
		// Most of the opcodes in this range are not valid,
		//  so we map this opcode onto one of the ones that is.
		fun := ddcbFunctions[(opcode&0x38)>>3]
		value = fun(z.core.MemRead(offset))
		z.core.MemWrite(offset, value)
	} else {
		bitNumber := (opcode & 0x38) >> 3
		if opcode < 0x80 {
			// bit test
			bitTestOp = true
			z.Flags.N = false
			z.Flags.H = true
			z.Flags.Z = z.core.MemRead(offset)&(1<<bitNumber) == 0
			z.Flags.P = z.Flags.Z
			z.Flags.S = (bitNumber == 7) && !z.Flags.Z
		} else if opcode < 0xc0 {
			// RES
			value = z.core.MemRead(offset) & (^(1 << bitNumber))
			z.core.MemWrite(offset, value)
		} else {
			// SET
			value = z.core.MemRead(offset) | (1 << bitNumber)
			z.core.MemWrite(offset, value)
		}
	}

	// This implements the undocumented shift, RES, and SET opcodes,
	//  which write their result to memory and also to an 8080 register.

	if !bitTestOp {
		//value := byte(1)
		switch opcode & 0x07 {
		case 0:
			z.B = value
		case 1:
			z.C = value
		case 2:
			z.D = value
		case 3:
			z.E = value
		case 4:
			z.H = value
		case 5:
			z.L = value
		// 6 is the documented opcode, which doesn't set a register.
		case 7:
			z.A = value
		}
	}

	z.CycleCounter += CycleCountsCb[opcode] + 8
}
