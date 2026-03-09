package js

var instructions = []func(s *Z80){
	// NOP
	0x00: func(s *Z80) {
		// NOP
	},
	// LD BC, nn
	0x01: func(s *Z80) {
		s.PC++
		s.C = s.core.MemRead(s.PC)
		s.PC++
		s.B = s.core.MemRead(s.PC)
	},
	// LD (BC), A
	0x02: func(s *Z80) {
		s.core.MemWrite(s.bc(), s.A)
	},
	// 0x03 : INC BC
	0x03: func(s *Z80) {
		s.incBc()
	},
	// 0x04 : INC B
	0x04: func(s *Z80) {
		s.B = s.doInc(s.B)
	},
	// 0x05 : DEC B
	0x05: func(s *Z80) {
		s.B = s.doDec(s.B)
	},
	// 0x06 : LD B, n
	0x06: func(s *Z80) {
		s.PC++
		s.B = s.core.MemRead(s.PC)
	},
	// 0x07 : RLCA
	0x07: func(s *Z80) {
		// This instruction is implemented as a special case of the
		//  more general Z80-specific RLC instruction.
		// Specifially, RLCA is a version of RLC A that affects fewer flags.
		// The same applies to RRCA, RLA, and RRA.
		tempS := s.Flags.S
		tempZ := s.Flags.Z
		tempP := s.Flags.P
		s.A = s.doRlc(s.A)
		s.Flags.S = tempS
		s.Flags.Z = tempZ
		s.Flags.P = tempP
	},
	// 0x08 : EX AF, AF'
	0x08: func(s *Z80) {
		s.A, s.AAlt = s.AAlt, s.A
		temp := s.getFlagsRegister()
		s.setFlagsRegister(s.getFlagsPrimeRegister())
		s.setFlagsPrimeRegister(temp)
	},
	// 0x09 : ADD HL, BC
	0x09: func(s *Z80) {
		s.doHlAdd(s.bc())
	},
	// 0x0a : LD A, (BC)
	0x0A: func(s *Z80) {
		s.A = s.core.MemRead(s.bc())
	},
	// 0x0b : DEC BC
	0x0B: func(s *Z80) {
		s.decBc()
	},
	// 0x0c : INC C
	0x0C: func(s *Z80) {
		s.C = s.doInc(s.C)
	},
	// 0x0d : DEC C
	0x0D: func(s *Z80) {
		s.C = s.doDec(s.C)
	},
	// 0x0e : LD C, n
	0x0E: func(s *Z80) {
		s.PC++
		s.C = s.core.MemRead(s.PC)
	},
	// 0x0f : RRCA
	0x0F: func(s *Z80) {
		tempS := s.Flags.S
		tempZ := s.Flags.Z
		tempP := s.Flags.P
		s.A = s.doRrc(s.A)
		s.Flags.S = tempS
		s.Flags.Z = tempZ
		s.Flags.P = tempP
	},
	// 0x10 : DJNZ nn
	0x10: func(s *Z80) {
		s.B--
		s.doConditionalRelativeJump(s.B != 0)
	},
	// 0x11 : LD DE, nn
	0x11: func(s *Z80) {
		s.PC++
		s.E = s.core.MemRead(s.PC)
		s.PC++
		s.D = s.core.MemRead(s.PC)
	},
	// 0x12 : LD (DE), A
	0x12: func(s *Z80) {
		s.core.MemWrite(s.de(), s.A)
	},
	// 0x13 : INC DE
	0x13: func(s *Z80) {
		s.incDe()
	},
	// 0x14 : INC D
	0x14: func(s *Z80) {
		s.D = s.doInc(s.D)
	},
	// 0x15 : DEC D
	0x15: func(s *Z80) {
		s.D = s.doDec(s.D)
	},
	// 0x16 : LD D, n
	0x16: func(s *Z80) {
		s.PC++
		s.D = s.core.MemRead(s.PC)
	},
	// 0x17 : RLA
	0x17: func(s *Z80) {
		tempS := s.Flags.S
		tempZ := s.Flags.Z
		tempP := s.Flags.P
		s.A = s.doRl(s.A)
		s.Flags.S = tempS
		s.Flags.Z = tempZ
		s.Flags.P = tempP
	},
	// 0x18 : JR n
	0x18: func(s *Z80) {
		var o = int8(s.core.MemRead(s.PC + 1))
		if o > 0 {
			s.PC += uint16(o)
		} else {
			s.PC -= uint16(-o)
		}
		s.PC++
	},
	// 0x19 : ADD HL, DE
	0x19: func(s *Z80) {
		s.doHlAdd(s.de())
	},
	// 0x1a : LD A, (DE)
	0x1A: func(s *Z80) {
		s.A = s.core.MemRead(s.de())
	},
	// 0x1b : DEC DE
	0x1B: func(s *Z80) {
		s.decDe()
	},
	// 0x1c : INC E
	0x1C: func(s *Z80) {
		s.E = s.doInc(s.E)
	},
	// 0x1d : DEC E
	0x1D: func(s *Z80) {
		s.E = s.doDec(s.E)
	},
	// 0x1e : LD E, n
	0x1E: func(s *Z80) {
		s.PC++
		s.E = s.core.MemRead(s.PC)
	},
	// 0x1f : RRA
	0x1F: func(s *Z80) {
		tempS := s.Flags.S
		tempZ := s.Flags.Z
		tempP := s.Flags.P
		s.A = s.doRr(s.A)
		s.Flags.S = tempS
		s.Flags.Z = tempZ
		s.Flags.P = tempP
	},
	// 0x20 : JR NZ, n
	0x20: func(s *Z80) {
		s.doConditionalRelativeJump(!s.Flags.Z)
	},
	// 0x21 : LD HL, nn
	0x21: func(s *Z80) {
		s.PC++
		s.L = s.core.MemRead(s.PC)
		s.PC++
		s.H = s.core.MemRead(s.PC)
	},
	// 0x22 : LD (nn), HL
	0x22: func(s *Z80) {
		addr := s.nextWord()
		s.core.MemWrite(addr, s.L)
		s.core.MemWrite(addr+1, s.H)
	},
	// 0x23 : INC HL
	0x23: func(s *Z80) {
		s.incHl()
	},
	// 0x24 : INC H
	0x24: func(s *Z80) {
		s.H = s.doInc(s.H)
	},
	// 0x25 : DEC H
	0x25: func(s *Z80) {
		s.H = s.doDec(s.H)
	},
	// 0x26 : LD H, n
	0x26: func(s *Z80) {
		s.PC++
		s.H = s.core.MemRead(s.PC)
	},
	// 0x27 : DAA
	0x27: func(s *Z80) {
		temp := s.A
		if !s.Flags.N {
			if s.Flags.H || ((s.A & 0x0f) > 9) {
				temp += 0x06
			}
			if s.Flags.C || (s.A > 0x99) {
				temp += 0x60
			}
		} else {
			if s.Flags.H || ((s.A & 0x0f) > 9) {
				temp -= 0x06
			}
			if s.Flags.C || (s.A > 0x99) {
				temp -= 0x60
			}
		}

		s.Flags.S = (temp & 0x80) != 0
		s.Flags.Z = temp == 0
		s.Flags.H = ((s.A & 0x10) ^ (temp & 0x10)) != 0
		s.Flags.P = ParityBits[temp]
		// DAA never clears the carry flag if it was already set,
		//  but it is able to set the carry flag if it was clear.
		// Don't ask me, I don't know.
		// Note also that we check for a BCD carry, instead of the usual.
		s.Flags.C = s.Flags.C || (s.A > 0x99)
		s.A = temp
		s.updateXYFlags(s.A)
	},
	// 0x28 : JR Z, n
	0x28: func(s *Z80) {
		s.doConditionalRelativeJump(s.Flags.Z)
	},
	// 0x29 : ADD HL, HL
	0x29: func(s *Z80) {
		s.doHlAdd(s.hl())
	},
	// 0x2a : LD HL, (nn)
	0x2A: func(s *Z80) {
		addr := s.nextWord()
		s.L = s.core.MemRead(addr)
		s.H = s.core.MemRead(addr + 1)
	},
	// 0x2b : DEC HL
	0x2B: func(s *Z80) {
		s.decHl()
	},
	// 0x2c : INC L
	0x2C: func(s *Z80) {
		s.L = s.doInc(s.L)
	},
	// 0x2d : DEC L
	0x2D: func(s *Z80) {
		s.L = s.doDec(s.L)
	},
	// 0x2e : LD L, n
	0x2E: func(s *Z80) {
		s.PC++
		s.L = s.core.MemRead(s.PC)
	},
	// 0x2f : CPL
	0x2F: func(s *Z80) {
		s.A = ^s.A
		s.Flags.N = true
		s.Flags.H = true
		s.updateXYFlags(s.A)
	},
	// 0x30 : JR NC, n
	0x30: func(s *Z80) {
		s.doConditionalRelativeJump(!s.Flags.C)
	},
	// 0x31 : LD SP, nn
	0x31: func(s *Z80) {
		s.PC++
		lo := s.core.MemRead(s.PC)
		s.PC++
		s.SP = (uint16(s.core.MemRead(s.PC)) << 8) | uint16(lo)
	},
	// 0x32 : LD (nn), A
	0x32: func(s *Z80) {
		s.core.MemWrite(s.nextWord(), s.A)
	},
	// 0x33 : INC SP
	0x33: func(s *Z80) {
		s.SP++
	},
	// 0x34 : INC (HL)
	0x34: func(s *Z80) {
		s.core.MemWrite(s.hl(), s.doInc(s.core.MemRead(s.hl())))
	},
	// 0x35 : DEC (HL)
	0x35: func(s *Z80) {
		s.core.MemWrite(s.hl(), s.doDec(s.core.MemRead(s.hl())))
	},
	// 0x36 : LD (HL), n
	0x36: func(s *Z80) {
		s.PC++
		s.core.MemWrite(s.hl(), s.core.MemRead(s.PC))
	},
	// 0x37 : SCF
	0x37: func(s *Z80) {
		s.Flags.N = false
		s.Flags.H = false
		s.Flags.C = true
		s.updateXYFlags(s.A)
	},
	// 0x38 : JR C, n
	0x38: func(s *Z80) {
		s.doConditionalRelativeJump(s.Flags.C)
	},
	// 0x39 : ADD HL, SP
	0x39: func(s *Z80) {
		s.doHlAdd(s.SP)
	},
	// 0x3a : LD A, (nn)
	0x3A: func(s *Z80) {
		s.A = s.core.MemRead(s.nextWord())
	},
	// 0x3b : DEC SP
	0x3B: func(s *Z80) {
		s.SP--
	},
	// 0x3c : INC A
	0x3C: func(s *Z80) {
		s.A = s.doInc(s.A)
	},
	// 0x3d : DEC A
	0x3D: func(s *Z80) {
		s.A = s.doDec(s.A)
	},
	// 0x3e : LD A, n
	0x3E: func(s *Z80) {
		s.PC++
		s.A = s.core.MemRead(s.PC)
	},
	// 0x3f : CCF
	0x3F: func(s *Z80) {
		s.Flags.N = false
		s.Flags.H = s.Flags.C
		s.Flags.C = !s.Flags.C
		s.updateXYFlags(s.A)
	},
	// 0xc0 : RET NZ
	0xC0: func(s *Z80) {
		s.doConditionalReturn(!s.Flags.Z)
	},
	// 0xc1 : POP BC
	0xC1: func(s *Z80) {
		result := s.PopWord()
		s.C = byte(result & 0xff)
		s.B = byte((result & 0xff00) >> 8)
	},
	// 0xc2 : JP NZ, nn
	0xC2: func(s *Z80) {
		s.doConditionalAbsoluteJump(!s.Flags.Z)
	},
	// 0xc3 : JP nn
	0xC3: func(s *Z80) {
		s.PC = uint16(s.core.MemRead(s.PC+1)) | (uint16(s.core.MemRead(s.PC+2)) << 8)
		s.PC--
	},
	// 0xc4 : CALL NZ, nn
	0xC4: func(s *Z80) {
		s.doConditionalCall(!s.Flags.Z)
	},
	// 0xc5 : PUSH BC
	0xC5: func(s *Z80) {
		s.pushWord((uint16(s.B) << 8) | uint16(s.C))
	},
	// 0xc6 : ADD A, n
	0xC6: func(s *Z80) {
		s.PC++
		s.doAdd(s.core.MemRead(s.PC))
	},
	// 0xc7 : RST 00h
	0xC7: func(s *Z80) {
		s.doReset(0x0000)
	},
	// 0xc8 : RET Z
	0xC8: func(s *Z80) {
		s.doConditionalReturn(s.Flags.Z)
	},
	// 0xc9 : RET
	0xC9: func(s *Z80) {
		s.PC = s.PopWord() - 1
	},
	// 0xca : JP Z, nn
	0xCA: func(s *Z80) {
		s.doConditionalAbsoluteJump(s.Flags.Z)
	},
	// 0xcb : CB Prefix
	0xCB: func(s *Z80) {
		s.opcodeCB()
	},
	// 0xcc : CALL Z, nn
	0xCC: func(s *Z80) {
		s.doConditionalCall(s.Flags.Z)
	},
	// 0xcd : CALL nn
	0xCD: func(s *Z80) {
		s.pushWord(s.PC + 3)
		s.PC = uint16(s.core.MemRead(s.PC+1)) | (uint16(s.core.MemRead(s.PC+2)) << 8)
		s.PC--
	},
	// 0xce : ADC A, n
	0xCE: func(s *Z80) {
		s.PC++
		s.doAdc(s.core.MemRead(s.PC))
	},
	// 0xcf : RST 08h
	0xCF: func(s *Z80) {
		s.doReset(0x0008)
	},
	// 0xd0 : RET NC
	0xD0: func(s *Z80) {
		s.doConditionalReturn(!s.Flags.C)
	},
	// 0xd1 : POP DE
	0xD1: func(s *Z80) {
		result := s.PopWord()
		s.E = byte(result & 0xff)
		s.D = byte((result & 0xff00) >> 8)
	},
	// 0xd2 : JP NC, nn
	0xD2: func(s *Z80) {
		s.doConditionalAbsoluteJump(!s.Flags.C)
	},
	// 0xd3 : OUT (n), A
	0xD3: func(s *Z80) {
		s.PC++
		s.core.IOWrite((uint16(s.A)<<8)|uint16(s.core.MemRead(s.PC)), s.A)
	},
	// 0xd4 : CALL NC, nn
	0xD4: func(s *Z80) {
		s.doConditionalCall(!s.Flags.C)
	},
	// 0xd5 : PUSH DE
	0xD5: func(s *Z80) {
		s.pushWord((uint16(s.D) << 8) | uint16(s.E))
	},
	// 0xd6 : SUB n
	0xD6: func(s *Z80) {
		s.PC++
		s.doSub(s.core.MemRead(s.PC))
	},
	// 0xd7 : RST 10h
	0xD7: func(s *Z80) {
		s.doReset(0x0010)
	},
	// 0xd8 : RET C
	0xD8: func(s *Z80) {
		s.doConditionalReturn(s.Flags.C)
	},
	// 0xd9 : EXX
	0xD9: func(s *Z80) {
		s.B, s.BAlt = s.BAlt, s.B
		s.C, s.CAlt = s.CAlt, s.C
		s.D, s.DAlt = s.DAlt, s.D
		s.E, s.EAlt = s.EAlt, s.E
		s.H, s.HAlt = s.HAlt, s.H
		s.L, s.LAlt = s.LAlt, s.L
	},
	// 0xda : JP C, nn
	0xDA: func(s *Z80) {
		s.doConditionalAbsoluteJump(s.Flags.C)
	},
	// 0xdb : IN A, (n)
	0xDB: func(s *Z80) {
		s.PC++
		s.A = s.core.IORead((uint16(s.A) << 8) | uint16(s.core.MemRead(s.PC)))
	},
	// 0xdc : CALL C, nn
	0xDC: func(s *Z80) {
		s.doConditionalCall(s.Flags.C)
	},
	// 0xdd : DD Prefix (IX instructions)
	0xDD: func(s *Z80) {
		s.opcodeDD()
	},
	// 0xde : SBC n
	0xDE: func(s *Z80) {
		s.PC++
		s.doSbc(s.core.MemRead(s.PC))
	},
	// 0xdf : RST 18h
	0xDF: func(s *Z80) {
		s.doReset(0x0018)
	},
	// 0xe0 : RET PO
	0xE0: func(s *Z80) {
		s.doConditionalReturn(!s.Flags.P)
	},
	// 0xe1 : POP HL
	0xE1: func(s *Z80) {
		result := s.PopWord()
		s.L = byte(result & 0xff)
		s.H = byte((result & 0xff00) >> 8)
	},
	// 0xe2 : JP PO, (nn)
	0xE2: func(s *Z80) {
		s.doConditionalAbsoluteJump(!s.Flags.P)
	},
	// 0xe3 : EX (SP), HL
	0xE3: func(s *Z80) {
		temp := s.core.MemRead(s.SP)
		s.core.MemWrite(s.SP, s.L)
		s.L = temp
		temp = s.core.MemRead(s.SP + 1)
		s.core.MemWrite(s.SP+1, s.H)
		s.H = temp
	},
	// 0xe4 : CALL PO, nn
	0xE4: func(s *Z80) {
		s.doConditionalCall(!s.Flags.P)
	},
	// 0xe5 : PUSH HL
	0xE5: func(s *Z80) {
		s.pushWord((uint16(s.H) << 8) | uint16(s.L))
	},
	// 0xe6 : AND n
	0xE6: func(s *Z80) {
		s.PC++
		s.doAnd(s.core.MemRead(s.PC))
	},
	// 0xe7 : RST 20h
	0xE7: func(s *Z80) {
		s.doReset(0x0020)
	},
	// 0xe8 : RET PE
	0xE8: func(s *Z80) {
		s.doConditionalReturn(s.Flags.P)
	},
	// 0xe9 : JP (HL)
	0xE9: func(s *Z80) {
		s.PC = uint16(s.H)<<8 | uint16(s.L)
		s.PC--
	},
	// 0xea : JP PE, nn
	0xEA: func(s *Z80) {
		s.doConditionalAbsoluteJump(s.Flags.P)
	},
	// 0xeb : EX DE, HL
	0xEB: func(s *Z80) {
		s.D, s.H = s.H, s.D
		s.E, s.L = s.L, s.E
	},
	// 0xec : CALL PE, nn
	0xEC: func(s *Z80) {
		s.doConditionalCall(s.Flags.P)
	},
	// 0xed : ED Prefix
	0xED: func(s *Z80) {
		s.opcodeED()
	},
	// 0xee : XOR n
	0xEE: func(s *Z80) {
		s.PC++
		s.doXor(s.core.MemRead(s.PC))
	},
	// 0xef : RST 28h
	0xEF: func(s *Z80) {
		s.doReset(0x0028)
	},
	// 0xf0 : RET P
	0xF0: func(s *Z80) {
		s.doConditionalReturn(!s.Flags.S)
	},
	// 0xf1 : POP AF
	0xF1: func(s *Z80) {
		var result = s.PopWord()
		s.setFlagsRegister(byte(result & 0xff))
		s.A = byte((result & 0xff00) >> 8)
	},
	// 0xf2 : JP P, nn
	0xF2: func(s *Z80) {
		s.doConditionalAbsoluteJump(!s.Flags.S)
	},
	// 0xf3 : DI
	0xF3: func(s *Z80) {
		// DI doesn't actually take effect until after the next instruction.
		s.DoDelayedDI = true
	},
	// 0xf4 : CALL P, nn
	0xF4: func(s *Z80) {
		s.doConditionalCall(!s.Flags.S)
	},
	// 0xf5 : PUSH AF
	0xF5: func(s *Z80) {
		s.pushWord(uint16(s.getFlagsRegister()) | (uint16(s.A) << 8))
	},
	// 0xf6 : OR n
	0xF6: func(s *Z80) {
		s.PC++
		s.doOr(s.core.MemRead(s.PC))
	},
	// 0xf7 : RST 30h
	0xF7: func(s *Z80) {
		s.doReset(0x0030)
	},
	// 0xf8 : RET M
	0xF8: func(s *Z80) {
		s.doConditionalReturn(s.Flags.S)
	},
	// 0xf9 : LD SP, HL
	0xF9: func(s *Z80) {
		s.SP = uint16(s.H)<<8 | uint16(s.L)
	},
	// 0xfa : JP M, nn
	0xFA: func(s *Z80) {
		s.doConditionalAbsoluteJump(s.Flags.S)
	},
	// 0xfb : EI
	0xFB: func(s *Z80) {
		// EI doesn't actually take effect until after the next instruction.
		s.DoDelayedEI = true
	},
	// 0xfc : CALL M, nn
	0xFC: func(s *Z80) {
		s.doConditionalCall(s.Flags.S)
	},
	// 0xfd : FD Prefix (IY instructions)
	0xFD: func(s *Z80) {
		s.opcodeFD()
	},
	// 0xfe : CP n
	0xFE: func(s *Z80) {
		s.PC++
		s.doCp(s.core.MemRead(s.PC))
	},
	// 0xff : RST 38h
	0xFF: func(s *Z80) {
		s.doReset(0x0038)
	},
}

func (z *Z80) nextWord() uint16 {
	z.PC++
	word := uint16(z.core.MemRead(z.PC))
	z.PC++
	word |= uint16(z.core.MemRead(z.PC)) << 8
	return word
}
