package z80em

func (z *Z80Type) opcodeCB() {
	z.incR()
	z.PC++
	opcode := z.core.M1MemRead(z.PC)
	bitNumber := (opcode & 0x38) >> 3
	regCode := opcode & 0x07
	if opcode < 0x40 {
		// Shift/rotate instructions
		opArray := []OpShift{z.doRlc, z.doRrc, z.doRl, z.doRr, z.doSla, z.doSra, z.doSll, z.doSrl}
		switch regCode {
		case 0:
			z.B = opArray[bitNumber](z.B)
		case 1:
			z.C = opArray[bitNumber](z.C)
		case 2:
			z.D = opArray[bitNumber](z.D)
		case 3:
			z.E = opArray[bitNumber](z.E)
		case 4:
			z.H = opArray[bitNumber](z.H)
		case 5:
			z.L = opArray[bitNumber](z.L)
		case 6:
			z.core.MemWrite(z.hl(), opArray[bitNumber](z.core.MemRead(z.hl())))
		default:
			z.A = opArray[bitNumber](z.A)
		}
	} else if opcode < 0x80 {
		// BIT instructions
		mask := byte(1 << bitNumber)
		switch regCode {
		case 0:
			z.Flags.Z = z.B&mask == 0
		case 1:
			z.Flags.Z = z.C&mask == 0
		case 2:
			z.Flags.Z = z.D&mask == 0
		case 3:
			z.Flags.Z = z.E&mask == 0
		case 4:
			z.Flags.Z = z.H&mask == 0
		case 5:
			z.Flags.Z = z.L&mask == 0
		case 6:
			z.Flags.Z = z.core.MemRead(z.hl())&mask == 0
		default:
			z.Flags.Z = z.A&mask == 0
		}
		z.Flags.N = false
		z.Flags.H = true
		z.Flags.P = z.Flags.Z
		z.Flags.S = (bitNumber == 7) && !z.Flags.Z
		// TODO: ZXALL fail this
		// For the BIT n, (HL) instruction, the X and Y flags are obtained
		//  from what is apparently an internal temporary register used for
		//  some of the 16-bit arithmetic instructions.
		// I haven't implemented that register here,
		//  so for now we'll set X and Y the same way for every BIT opcode,
		//  which means that they will usually be wrong for BIT n, (HL).
		z.Flags.Y = (bitNumber == 5) && !z.Flags.Z
		z.Flags.X = (bitNumber == 3) && !z.Flags.Z
	} else if opcode < 0xC0 {
		// RES instructions
		negMask := byte(^(1 << bitNumber))
		switch regCode {
		case 0:
			z.B &= negMask
		case 1:
			z.C &= negMask
		case 2:
			z.D &= negMask
		case 3:
			z.E &= negMask
		case 4:
			z.H &= negMask
		case 5:
			z.L &= negMask
		case 6:
			z.core.MemWrite(z.hl(), z.core.MemRead(z.hl())&negMask)
		default:
			z.A &= negMask
		}
	} else {
		// SET instructions
		mask := byte(1 << bitNumber)
		switch regCode {
		case 0:
			z.B |= mask
		case 1:
			z.C |= mask
		case 2:
			z.D |= mask
		case 3:
			z.E |= mask
		case 4:
			z.H |= mask
		case 5:
			z.L |= mask
		case 6:
			z.core.MemWrite(z.hl(), z.core.MemRead(z.hl())|mask)
		default:
			z.A |= mask
		}
	}
	z.CycleCounter += CycleCountsCb[opcode]
}
