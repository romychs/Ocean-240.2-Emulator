package c99

import log "github.com/sirupsen/logrus"

// executes a ED opcode
func (z *Z80) execOpcodeED(opcode byte) {
	z.cycleCount += uint32(cyclesED[opcode])
	z.incR()
	switch opcode {
	case 0x47:
		z.i = z.a // ld i,a
	case 0x4F:
		z.r = z.a // ld r,a
	case 0x57:
		// ld a,i
		z.a = z.i
		z.sf = z.a&0x80 != 0
		z.zf = z.a == 0
		z.hf = false
		z.nf = false
		z.pf = z.iff2
		z.updateXY(z.a)
	case 0x5F:
		// ld a,r
		z.a = z.r
		z.sf = z.a&0x80 != 0
		z.zf = z.a == 0
		z.hf = false
		z.nf = false
		z.pf = z.iff2
	case 0x45, 0x55, 0x5D, 0x65, 0x6D, 0x75, 0x7D:
		// retn
		z.iff1 = z.iff2
		z.ret()
	case 0x4D:
		z.ret() // reti

	case 0xA0:
		z.ldi() // ldi
	case 0xB0:
		{
			z.ldi()

			if z.bc() != 0 {
				z.pc -= 2
				z.cycleCount += 5
				z.memPtr = z.pc + 1
			}
		} // ldir

	case 0xA8:
		z.ldd() // ldd
	case 0xB8:
		{
			z.ldd()

			if z.bc() != 0 {
				z.pc -= 2
				z.cycleCount += 5
				z.memPtr = z.pc + 1
			}
		} // lddr

	case 0xA1:
		z.cpi() // cpi
	case 0xA9:
		z.cpd() // cpd
	case 0xB1:
		// cpir
		z.cpi()
		if z.bc() != 0 && !z.zf {
			z.pc -= 2
			z.cycleCount += 5
			z.memPtr = z.pc + 1
		} else {
			//z.mem_ptr++
		}
		//z.cpir()
	case 0xB9:
		// cpdr
		z.cpd()
		if z.bc() != 0 && !z.zf {
			z.pc -= 2
			z.cycleCount += 5
			z.memPtr = z.pc + 1
		} else {
			//z.mem_ptr++
		}
	case 0x40:
		z.inRC(&z.b) // in b, (c)
		z.memPtr = z.bc() + 1
	case 0x48:
		z.memPtr = z.bc() + 1
		z.inRC(&z.c) // in c, (c)
		z.updateXY(z.c)
	//case 0x4e:
	// ld c,(iy+dd)

	case 0x50:
		z.inRC(&z.d) // in d, (c)
		z.memPtr = z.bc() + 1
	case 0x58:
		// in e, (c)
		z.inRC(&z.e)
		z.memPtr = z.bc() + 1
		z.updateXY(z.e)
	case 0x60:
		z.inRC(&z.h) // in h, (c)
		z.memPtr = z.bc() + 1
	case 0x68:
		z.inRC(&z.l) // in l, (c)
		z.memPtr = z.bc() + 1
		z.updateXY(z.l)
	case 0x70:
		// in (c)
		var val byte
		z.inRC(&val)
		z.memPtr = z.bc() + 1
	case 0x78:
		// in a, (c)
		z.inRC(&z.a)
		z.memPtr = z.bc() + 1
		z.updateXY(z.a)
	case 0xA2:
		z.ini() // ini
	case 0xB2:
		// inir
		z.ini()
		if z.b > 0 {
			z.pc -= 2
			z.cycleCount += 5
		}
	case 0xAA:
		// ind
		z.ind()
	case 0xBA:
		// indr
		z.ind()
		if z.b > 0 {
			z.pc -= 2
			z.cycleCount += 5
		}
	case 0x41:
		z.core.IOWrite(z.bc(), z.b) // out (c), b
		z.memPtr = z.bc() + 1
	case 0x49:
		z.core.IOWrite(z.bc(), z.c) // out (c), c
		z.memPtr = z.bc() + 1
	case 0x51:
		z.core.IOWrite(z.bc(), z.d) // out (c), d
		z.memPtr = z.bc() + 1
	case 0x59:
		z.core.IOWrite(z.bc(), z.e) // out (c), e
		z.memPtr = z.bc() + 1
	case 0x61:
		z.core.IOWrite(z.bc(), z.h) // out (c), h
		z.memPtr = z.bc() + 1
	case 0x69:
		z.core.IOWrite(z.bc(), z.l) // out (c), l
		z.memPtr = z.bc() + 1
	case 0x71:
		z.core.IOWrite(z.bc(), 0) // out (c), 0
		z.memPtr = z.bc() + 1
	case 0x79:
		// out (c), a
		z.core.IOWrite(z.bc(), z.a)
		z.memPtr = z.bc() + 1
	case 0xA3:
		z.outi() // outi
	case 0xB3:
		// otir
		z.outi()
		if z.b > 0 {
			z.pc -= 2
			z.cycleCount += 5
		}
	case 0xAB:
		z.outd() // outd
	case 0xBB:
		// otdr
		z.outd()
		if z.b > 0 {
			z.cycleCount += 5
			z.pc -= 2
		}

	case 0x42:
		z.sbcHL(z.bc()) // sbc hl,bc
	case 0x52:
		z.sbcHL(z.de()) // sbc hl,de
	case 0x62:
		z.sbcHL(z.hl()) // sbc hl,hl
	case 0x72:
		z.sbcHL(z.sp) // sbc hl,sp
	case 0x4A:
		z.adcHL(z.bc()) // adc hl,bc
	case 0x5A:
		z.adcHL(z.de()) // adc hl,de
	case 0x6A:
		z.adcHL(z.hl()) // adc hl,hl
	case 0x7A:
		z.adcHL(z.sp) // adc hl,sp
	case 0x43:
		// ld (**), bc
		addr := z.nextW()
		z.ww(addr, z.bc())
		z.memPtr = addr + 1
	case 0x53:
		// ld (**), de
		addr := z.nextW()
		z.ww(addr, z.de())
		z.memPtr = addr + 1
	case 0x63:
		// ld (**), hl
		addr := z.nextW()
		z.ww(addr, z.hl())
		z.memPtr = addr + 1
	case 0x73:
		// ld (**), hl
		addr := z.nextW()
		z.ww(addr, z.sp)
		z.memPtr = addr + 1
	case 0x4B:
		// ld bc, (**)
		addr := z.nextW()
		z.setBC(z.rw(addr))
		z.memPtr = addr + 1
	case 0x5B:
		// ld de, (**)
		addr := z.nextW()
		z.setDE(z.rw(addr))
		z.memPtr = addr + 1
	case 0x6B:
		// ld hl, (**)
		addr := z.nextW()
		z.setHL(z.rw(addr))
		z.memPtr = addr + 1
	case 0x7B:
		// ld sp,(**)
		addr := z.nextW()
		z.sp = z.rw(addr)
		z.memPtr = addr + 1
	case 0x44, 0x54, 0x64, 0x74, 0x4C, 0x5C, 0x6C, 0x7C:
		z.a = z.subB(0, z.a, false) // neg
	case 0x46, 0x4e, 0x66, 0x6e:
		z.interruptMode = 0 // im 0
	case 0x56, 0x76:
		z.interruptMode = 1 // im 1
	case 0x5E, 0x7E:
		z.interruptMode = 2 // im 2
	case 0x67:
		// rrd
		a := z.a
		val := z.rb(z.hl())
		z.a = (a & 0xF0) | (val & 0xF)
		z.wb(z.hl(), (val>>4)|(a<<4))

		z.nf = false
		z.hf = false
		z.updateXY(z.a)
		z.zf = z.a == 0
		z.sf = z.a&0x80 != 0
		z.pf = parity(z.a)
		z.memPtr = z.hl() + 1
	case 0x6F:
		// rld
		a := z.a
		val := z.rb(z.hl())
		z.a = (a & 0xF0) | (val >> 4)
		z.wb(z.hl(), (val<<4)|(a&0xF))
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
		z.zf = z.a == 0
		z.sf = z.a&0x80 != 0
		z.pf = parity(z.a)
		z.memPtr = z.hl() + 1
	default:
		log.Errorf("Unknown ED opcode: %02X\n", opcode)
	}
}
