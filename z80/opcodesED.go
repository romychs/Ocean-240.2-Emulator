package z80

import log "github.com/sirupsen/logrus"

// executes a ED opcode
func (z *Z80) exec_opcode_ed(opcode byte) {
	z.cyc += uint64(cyc_ed[opcode])
	z.inc_r()
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

			if z.get_bc() != 0 {
				z.pc -= 2
				z.cyc += 5
				z.mem_ptr = z.pc + 1
			}
		} // ldir

	case 0xA8:
		z.ldd() // ldd
	case 0xB8:
		{
			z.ldd()

			if z.get_bc() != 0 {
				z.pc -= 2
				z.cyc += 5
				z.mem_ptr = z.pc + 1
			}
		} // lddr

	case 0xA1:
		z.cpi() // cpi
	case 0xA9:
		z.cpd() // cpd
	case 0xB1:
		{
			z.cpi()
			if z.get_bc() != 0 && !z.zf {
				z.pc -= 2
				z.cyc += 5
				z.mem_ptr = z.pc + 1
			} else {
				z.mem_ptr += 1
			}
		} // cpir
	case 0xB9:
		{
			z.cpd()
			if z.get_bc() != 0 && !z.zf {
				z.pc -= 2
				z.cyc += 5
			} else {
				z.mem_ptr += 1
			}
		} // cpdr

	case 0x40:
		z.in_r_c(&z.b) // in b, (c)
	case 0x48:
		z.in_r_c(&z.c) // in c, (c)
	case 0x50:
		z.in_r_c(&z.d) // in d, (c)
	case 0x58:
		z.in_r_c(&z.e) // in e, (c)
	case 0x60:
		z.in_r_c(&z.h) // in h, (c)
	case 0x68:
		z.in_r_c(&z.l) // in l, (c)
	case 0x70:
		// in (c)
		var val byte
		z.in_r_c(&val)
	case 0x78:
		// in a, (c)
		z.in_r_c(&z.a)
		z.mem_ptr = z.get_bc() + 1
	case 0xA2:
		z.ini() // ini
	case 0xB2:
		// inir
		z.ini()
		if z.b > 0 {
			z.pc -= 2
			z.cyc += 5
		}
	case 0xAA:
		z.ind() // ind
	case 0xBA:
		// indr
		z.ind()
		if z.b > 0 {
			z.pc -= 2
			z.cyc += 5
		}
	case 0x41:
		z.core.IOWrite(z.get_bc(), z.b) // out (c), b
	case 0x49:
		z.core.IOWrite(z.get_bc(), z.c) // out (c), c
	case 0x51:
		z.core.IOWrite(z.get_bc(), z.d) // out (c), d
	case 0x59:
		z.core.IOWrite(z.get_bc(), z.e) // out (c), e
	case 0x61:
		z.core.IOWrite(z.get_bc(), z.h) // out (c), h
	case 0x69:
		z.core.IOWrite(z.get_bc(), z.l) // out (c), l
	case 0x71:
		z.core.IOWrite(z.get_bc(), 0) // out (c), 0
	case 0x79:
		// out (c), a
		z.core.IOWrite(z.get_bc(), z.a)
		z.mem_ptr = z.get_bc() + 1
	case 0xA3:
		z.outi() // outi
	case 0xB3:
		// otir
		z.outi()
		if z.b > 0 {
			z.pc -= 2
			z.cyc += 5
		}
	case 0xAB:
		z.outd() // outd
	case 0xBB:
		// otdr
		z.outd()
		if z.b > 0 {
			z.pc -= 2
		}

	case 0x42:
		z.sbchl(z.get_bc()) // sbc hl,bc
	case 0x52:
		z.sbchl(z.get_de()) // sbc hl,de
	case 0x62:
		z.sbchl(z.get_hl()) // sbc hl,hl
	case 0x72:
		z.sbchl(z.sp) // sbc hl,sp
	case 0x4A:
		z.adchl(z.get_bc()) // adc hl,bc
	case 0x5A:
		z.adchl(z.get_de()) // adc hl,de
	case 0x6A:
		z.adchl(z.get_hl()) // adc hl,hl
	case 0x7A:
		z.adchl(z.sp) // adc hl,sp
	case 0x43:
		// ld (**), bc
		addr := z.nextw()
		z.ww(addr, z.get_bc())
		z.mem_ptr = addr + 1
	case 0x53:
		// ld (**), de
		addr := z.nextw()
		z.ww(addr, z.get_de())
		z.mem_ptr = addr + 1
	case 0x63:
		// ld (**), hl
		addr := z.nextw()
		z.ww(addr, z.get_hl())
		z.mem_ptr = addr + 1
	case 0x73:
		// ld (**), hl
		addr := z.nextw()
		z.ww(addr, z.sp)
		z.mem_ptr = addr + 1
	case 0x4B:
		// ld bc, (**)
		addr := z.nextw()
		z.set_bc(z.rw(addr))
		z.mem_ptr = addr + 1
	case 0x5B:
		// ld de, (**)
		addr := z.nextw()
		z.set_de(z.rw(addr))
		z.mem_ptr = addr + 1
	case 0x6B:
		// ld hl, (**)
		addr := z.nextw()
		z.set_hl(z.rw(addr))
		z.mem_ptr = addr + 1
	case 0x7B:
		// ld sp,(**)
		addr := z.nextw()
		z.sp = z.rw(addr)
		z.mem_ptr = addr + 1
	case 0x44, 0x54, 0x64, 0x74, 0x4C, 0x5C, 0x6C, 0x7C:
		z.a = z.subb(0, z.a, false) // neg
	case 0x46, 0x66:
		z.interrupt_mode = 0 // im 0
	case 0x56, 0x76:
		z.interrupt_mode = 1 // im 1
	case 0x5E, 0x7E:
		z.interrupt_mode = 2 // im 2
	case 0x67:
		// rrd
		a := z.a
		val := z.rb(z.get_hl())
		z.a = (a & 0xF0) | (val & 0xF)
		z.wb(z.get_hl(), (val>>4)|(a<<4))

		z.nf = false
		z.hf = false
		z.updateXY(z.a)
		z.zf = z.a == 0
		z.sf = z.a&0x80 != 0
		z.pf = parity(z.a)
		z.mem_ptr = z.get_hl() + 1
	case 0x6F:
		// rld
		a := z.a
		val := z.rb(z.get_hl())
		z.a = (a & 0xF0) | (val >> 4)
		z.wb(z.get_hl(), (val<<4)|(a&0xF))
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
		z.zf = z.a == 0
		z.sf = z.a&0x80 != 0
		z.pf = parity(z.a)
		z.mem_ptr = z.get_hl() + 1
	default:
		log.Errorf("Unknown ED opcode: %02X\n", opcode)
	}
}
