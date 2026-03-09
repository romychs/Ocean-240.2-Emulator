package js

var edInstructions = []func(s *Z80){
	// 0x40 : IN B, (C)
	0x40: func(s *Z80) {
		s.B = s.doIn(s.bc())
	},
	// 0x41 : OUT (C), B
	0x41: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.B)
	},
	// 0x42 : SBC HL, BC
	0x42: func(s *Z80) {
		s.doHlSbc(s.bc())
	},
	// 0x43 : LD (nn), BC
	0x43: func(s *Z80) {
		s.setWord(s.nextWord(), s.bc())
	},
	// 0x44 : NEG
	0x44: func(s *Z80) {
		s.doNeg()
	},
	// 0x45 : RETN
	0x45: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x46 : IM 0
	0x46: func(s *Z80) {
		s.IMode = 0
	},
	// 0x47 : LD I, A
	0x47: func(s *Z80) {
		s.I = s.A
	},
	// 0x48 : IN C, (C)
	0x48: func(s *Z80) {
		s.C = s.doIn(s.bc())
	},
	// 0x49 : OUT (C), C
	0x49: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.C)
	},
	// 0x4a : ADC HL, BC
	0x4A: func(s *Z80) {
		s.doHlAdc(s.bc())
	},
	// 0x4b : LD BC, (nn)
	0x4B: func(s *Z80) {
		s.setBc(s.getWord(s.nextWord()))
	},
	// 0x4c : NEG (Undocumented)
	0x4C: func(s *Z80) {
		s.doNeg()
	},
	// 0x4d : RETI
	0x4D: func(s *Z80) {
		s.PC = s.PopWord() - 1
	},
	// 0x4e : IM 0 (Undocumented)
	0x4E: func(s *Z80) {
		s.IMode = 0
	},
	// 0x4f : LD R, A
	0x4F: func(s *Z80) {
		s.R = s.A
	},
	// 0x50 : IN D, (C)
	0x50: func(s *Z80) {
		s.D = s.doIn(s.bc())
	},
	// 0x51 : OUT (C), D
	0x51: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.D)
	},
	// 0x52 : SBC HL, DE
	0x52: func(s *Z80) {
		s.doHlSbc(s.de())
	},
	// 0x53 : LD (nn), DE
	0x53: func(s *Z80) {
		s.setWord(s.nextWord(), s.de())
	},
	// 0x54 : NEG (Undocumented)
	0x54: func(s *Z80) {
		s.doNeg()
	},
	// 0x55 : RETN
	0x55: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x56 : IM 1
	0x56: func(s *Z80) {
		s.IMode = 1
	},
	// 0x57 : LD A, I
	0x57: func(s *Z80) {
		s.A = s.I
		s.Flags.S = s.A&0x80 != 0
		s.Flags.Z = s.A == 0
		s.Flags.H = false
		s.Flags.P = s.Iff2
		s.Flags.N = false
		s.updateXYFlags(s.A)

	},
	// 0x58 : IN E, (C)
	0x58: func(s *Z80) {
		s.E = s.doIn(s.bc())
	},
	// 0x59 : OUT (C), E
	0x59: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.E)
	},
	// 0x5a : ADC HL, DE
	0x5A: func(s *Z80) {
		s.doHlAdc(s.de())
	},
	// 0x5b : LD DE, (nn)
	0x5B: func(s *Z80) {
		s.setDe(s.getWord(s.nextWord()))
	},
	// 0x5c : NEG (Undocumented)
	0x5C: func(s *Z80) {
		s.doNeg()
	},
	// 0x5d : RETN
	0x5D: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x5e : IM 2
	0x5E: func(s *Z80) {
		s.IMode = 2
	},
	// 0x5f : LD A, R
	0x5F: func(s *Z80) {
		s.A = s.R
		s.Flags.S = s.A&0x80 != 0
		s.Flags.Z = s.A == 0
		s.Flags.H = false
		s.Flags.P = s.Iff2
		s.Flags.N = false
		s.updateXYFlags(s.A)

	},
	// 0x60 : IN H, (C)
	0x60: func(s *Z80) {
		s.H = s.doIn(s.bc())
	},
	// 0x61 : OUT (C), H
	0x61: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.H)
	},
	// 0x62 : SBC HL, HL
	0x62: func(s *Z80) {
		s.doHlSbc(s.hl())
	},
	// 0x63 : LD (nn), HL (Undocumented)
	0x63: func(s *Z80) {
		s.setWord(s.nextWord(), s.hl())
	},
	// 0x64 : NEG (Undocumented)
	0x64: func(s *Z80) {
		s.doNeg()
	},
	// 0x65 : RETN
	0x65: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x66 : IM 0
	0x66: func(s *Z80) {
		s.IMode = 0
	},
	// 0x67 : RRD
	0x67: func(s *Z80) {
		hlValue := s.core.M1MemRead(s.hl())
		temp1 := hlValue & 0x0f
		temp2 := s.A & 0x0f
		hlValue = ((hlValue & 0xf0) >> 4) | (temp2 << 4)
		s.A = (s.A & 0xf0) | temp1
		s.core.MemWrite(s.hl(), hlValue)
		s.Flags.S = s.A&0x80 != 0
		s.Flags.Z = s.A == 0
		s.Flags.H = false
		s.Flags.P = ParityBits[s.A]
		s.Flags.N = false
		s.updateXYFlags(s.A)

	},
	// 0x68 : IN L, (C)
	0x68: func(s *Z80) {
		s.L = s.doIn(s.bc())
	},
	// 0x69 : OUT (C), L
	0x69: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.L)
	},
	// 0x6a : ADC HL, HL
	0x6A: func(s *Z80) {
		s.doHlAdc(s.hl())
	},
	// 0x6b : LD HL, (nn) (Undocumented)
	0x6B: func(s *Z80) {
		s.setHl(s.getWord(s.nextWord()))
	},
	// 0x6C : NEG (Undocumented)
	0x6C: func(s *Z80) {
		s.doNeg()
	},
	// 0x6D : RETN
	0x6D: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x6E : IM 0
	0x6E: func(s *Z80) {
		s.IMode = 0
	},
	// 0x6f : RLD
	0x6F: func(s *Z80) {
		hlValue := s.core.MemRead(s.hl())
		temp1 := hlValue & 0xf0
		temp2 := s.A & 0x0f
		hlValue = ((hlValue & 0x0f) << 4) | temp2
		s.A = (s.A & 0xf0) | (temp1 >> 4)
		s.core.MemWrite(s.hl(), hlValue)

		s.Flags.S = s.A&0x80 != 0
		s.Flags.Z = s.A == 0
		s.Flags.H = false
		s.Flags.P = ParityBits[s.A]
		s.Flags.N = false
		s.updateXYFlags(s.A)
	},
	// 0x70 : INF
	0x70: func(s *Z80) {
		s.doIn(s.bc())
	},
	// 0x71 : OUT (C), 0 (Undocumented)
	0x71: func(s *Z80) {
		s.core.IOWrite(s.bc(), 0)
	},
	// 0x72 : SBC HL, SP
	0x72: func(s *Z80) {
		s.doHlSbc(s.SP)
	},
	// 0x73 : LD (nn), SP
	0x73: func(s *Z80) {
		s.setWord(s.nextWord(), s.SP)
	},
	// 0x74 : NEG (Undocumented)
	0x74: func(s *Z80) {
		s.doNeg()
	},
	// 0x75 : RETN
	0x75: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x76 : IM 1
	0x76: func(s *Z80) {
		s.IMode = 1
	},
	// 0x78 : IN A, (C)
	0x78: func(s *Z80) {
		s.A = s.core.IORead(s.bc())
	},
	// 0x79 : OUT (C), A
	0x79: func(s *Z80) {
		s.core.IOWrite(s.bc(), s.A)
	},
	// 0x7a : ADC HL, SP
	0x7A: func(s *Z80) {
		s.doHlAdc(s.SP)
	},
	// 0x7b : LD SP, (nn)
	0x7B: func(s *Z80) {
		s.SP = s.getWord(s.nextWord())
	},
	// 0x7c : NEG (Undocumented)
	0x7C: func(s *Z80) {
		s.doNeg()
	},
	// 0x7d : RETN
	0x7D: func(s *Z80) {
		s.PC = s.PopWord() - 1
		s.Iff1 = s.Iff2
	},
	// 0x7e : IM 2
	0x7E: func(s *Z80) {
		s.IMode = 2
	},
	// 0xa0 : LDI
	0xA0: func(s *Z80) {
		s.doLdi()
	},
	// 0xa1 : CPI
	0xA1: func(s *Z80) {
		s.doCpi()
	},
	// 0xa2 : INI
	0xA2: func(s *Z80) {
		s.doIni()
	},
	// 0xa3 : OUTI
	0xA3: func(s *Z80) {
		s.doOuti()
	},
	// 0xa8 : LDD
	0xA8: func(s *Z80) {
		s.doLdd()
	},
	// 0xa9 : CPD
	0xA9: func(s *Z80) {
		s.doCpd()
	},
	// 0xaa : IND
	0xAA: func(s *Z80) {
		s.doInd()
	},
	// 0xab : OUTD
	0xAB: func(s *Z80) {
		s.doOutd()
	},
	// 0xb0 : LDIR
	0xB0: func(s *Z80) {
		s.doLdi()
		if (s.B | s.C) != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xb1 : CPIR
	0xB1: func(s *Z80) {
		s.doCpi()
		if !s.Flags.Z && (s.B|s.C) != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xb2 : INIR
	0xB2: func(s *Z80) {
		s.doIni()
		if s.B != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xb3 : OTIR
	0xB3: func(s *Z80) {
		s.doOuti()
		if s.B != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xb8 : LDDR
	0xB8: func(s *Z80) {
		s.doLdd()
		if (s.B | s.C) != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xb9 : CPDR
	0xB9: func(s *Z80) {
		s.doCpd()
		if !s.Flags.Z && (s.B|s.C) != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xba : INDR
	0xBA: func(s *Z80) {
		s.doInd()
		if s.B != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
	// 0xbb : OTDR
	0xBB: func(s *Z80) {
		s.doOutd()
		if s.B != 0 {
			s.CycleCounter += 5
			s.PC -= 2
		}
	},
}

func (z *Z80) opcodeED() {
	z.incR()
	z.PC++

	opcode := z.core.M1MemRead(z.PC)

	fun := edInstructions[opcode]
	if fun != nil {
		fun(z)
		z.CycleCounter += CycleCountsEd[opcode]
	} else {
		z.PC--
		z.CycleCounter += CycleCounts[0]
	}

}
