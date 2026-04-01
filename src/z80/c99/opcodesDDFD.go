package c99

// executes a DD/FD opcode (IZ = IX or IY)
func (z *Z80) execOpcodeDDFD(opcode byte, iz *uint16) {
	z.cycleCount += uint32(cyclesDDFD[opcode])
	z.incR()

	switch opcode {
	case 0xE1:
		*iz = z.popW() // pop iz
	case 0xE5:
		z.pushW(*iz) // push iz
	case 0xE9:
		// jp iz
		z.pc = *iz
		//z.jump(*iz)
	case 0x09:
		z.addIZ(iz, z.bc()) // add iz,bc
	case 0x19:
		z.addIZ(iz, z.de()) // add iz,de
	case 0x29:
		z.addIZ(iz, *iz) // add iz,iz
	case 0x39:
		z.addIZ(iz, z.sp) // add iz,sp

	case 0x84:
		z.a = z.addB(z.a, byte(*iz>>8), false) // add a,izh
	case 0x85:
		z.a = z.addB(z.a, byte(*iz), false) // add a,izl
	case 0x8C:
		z.a = z.addB(z.a, byte(*iz>>8), z.cf) // adc a,izh
	case 0x8D:
		z.a = z.addB(z.a, byte(*iz), z.cf) // adc a,izl
	case 0x86:
		z.a = z.addB(z.a, z.rb(z.displace(*iz, z.nextB())), false) // add a,(iz+*)
	case 0x8E:
		z.a = z.addB(z.a, z.rb(z.displace(*iz, z.nextB())), z.cf) // adc a,(iz+*)
	case 0x96:
		z.a = z.subB(z.a, z.rb(z.displace(*iz, z.nextB())), false) // sub (iz+*)
	case 0x9E:
		z.a = z.subB(z.a, z.rb(z.displace(*iz, z.nextB())), z.cf) // sbc (iz+*)
	case 0x94:
		z.a = z.subB(z.a, byte(*iz>>8), false) // sub izh
	case 0x95:
		z.a = z.subB(z.a, byte(*iz), false) // sub izl
	case 0x9C:
		z.a = z.subB(z.a, byte(*iz>>8), z.cf) // sbc izh
	case 0x9D:
		z.a = z.subB(z.a, byte(*iz), z.cf) // sbc izl

	case 0xA6:
		z.lAnd(z.rb(z.displace(*iz, z.nextB()))) // and (iz+*)
	case 0xA4:
		z.lAnd(byte(*iz >> 8)) // and izh
	case 0xA5:
		z.lAnd(byte(*iz)) // and izl

	case 0xAE:
		z.lXor(z.rb(z.displace(*iz, z.nextB()))) // xor (iz+*)
	case 0xAC:
		z.lXor(byte(*iz >> 8)) // xor izh
	case 0xAD:
		z.lXor(byte(*iz)) // xor izl
	case 0xB6:
		z.lOr(z.rb(z.displace(*iz, z.nextB()))) // or (iz+*)
	case 0xB4:
		z.lOr(byte(*iz >> 8)) // or izh
	case 0xB5:
		z.lOr(byte(*iz)) // or izl
	case 0xBE:
		z.cp(z.rb(z.displace(*iz, z.nextB()))) // cp (iz+*)
	case 0xBC:
		z.cp(byte(*iz >> 8)) // cp izh
	case 0xBD:
		z.cp(byte(*iz)) // cp izl
	case 0x23:
		*iz += 1 // inc iz
	case 0x2B:
		*iz -= 1 // dec iz
	case 0x34:
		// inc (iz+*)
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.inc(z.rb(addr)))
	case 0x35:
		// dec (iz+*)
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.dec(z.rb(addr)))
	case 0x24:
		*iz = (*iz & 0x00ff) | (uint16(z.inc(byte(*iz>>8))) << 8) // inc izh
	case 0x25:
		*iz = (*iz & 0x00ff) | (uint16(z.dec(byte(*iz>>8))) << 8) // dec izh
	case 0x2C:
		*iz = (*iz & 0xff00) | uint16(z.inc(byte(*iz))) // inc izl
	case 0x2D:
		*iz = (*iz & 0xff00) | uint16(z.dec(byte(*iz))) // dec izl
	case 0x2A:
		addr := z.nextW()
		*iz = z.rw(addr) // ld iz,(**)
		z.memPtr = addr + 1
	case 0x22:
		addr := z.nextW()
		z.ww(addr, *iz) // ld (**),iz
		z.memPtr = addr + 1
	case 0x21:
		*iz = z.nextW() // ld iz,**
	case 0x36:
		// ld (iz+*),*
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.nextB())
	case 0x70:
		z.wb(z.displace(*iz, z.nextB()), z.b) // ld (iz+*),b
	case 0x71:
		z.wb(z.displace(*iz, z.nextB()), z.c) // ld (iz+*),c
	case 0x72:
		z.wb(z.displace(*iz, z.nextB()), z.d) // ld (iz+*),d
	case 0x73:
		z.wb(z.displace(*iz, z.nextB()), z.e) // ld (iz+*),e
	case 0x74:
		z.wb(z.displace(*iz, z.nextB()), z.h) // ld (iz+*),h
	case 0x75:
		z.wb(z.displace(*iz, z.nextB()), z.l) // ld (iz+*),l
	case 0x77:
		z.wb(z.displace(*iz, z.nextB()), z.a) // ld (iz+*),a
	case 0x46:
		z.b = z.rb(z.displace(*iz, z.nextB())) // ld b,(iz+*)
	case 0x4E:
		z.c = z.rb(z.displace(*iz, z.nextB())) // ld c,(iz+*)
	case 0x56:
		z.d = z.rb(z.displace(*iz, z.nextB())) // ld d,(iz+*)
	case 0x5E:
		z.e = z.rb(z.displace(*iz, z.nextB())) // ld e,(iz+*)
	case 0x66:
		z.h = z.rb(z.displace(*iz, z.nextB())) // ld h,(iz+*)
	case 0x6E:
		z.l = z.rb(z.displace(*iz, z.nextB())) // ld l,(iz+*)
	case 0x7E:
		z.a = z.rb(z.displace(*iz, z.nextB())) // ld a,(iz+*)
	case 0x44:
		z.b = byte(*iz >> 8) // ld b,izh
	case 0x4C:
		z.c = byte(*iz >> 8) // ld c,izh
	case 0x54:
		z.d = byte(*iz >> 8) // ld d,izh
	case 0x5C:
		z.e = byte(*iz >> 8) // ld e,izh
	case 0x7C:
		z.a = byte(*iz >> 8) // ld a,izh
	case 0x45:
		z.b = byte(*iz) // ld b,izl
	case 0x4D:
		z.c = byte(*iz) // ld c,izl
	case 0x55:
		z.d = byte(*iz) // ld d,izl
	case 0x5D:
		z.e = byte(*iz) // ld e,izl
	case 0x7D:
		z.a = byte(*iz) // ld a,izl
	case 0x60:
		*iz = (*iz & 0x00ff) | (uint16(z.b) << 8) // ld izh,b
	case 0x61:
		*iz = (*iz & 0x00ff) | (uint16(z.c) << 8) // ld izh,c
	case 0x62:
		*iz = (*iz & 0x00ff) | (uint16(z.d) << 8) // ld izh,d
	case 0x63:
		*iz = (*iz & 0x00ff) | (uint16(z.e) << 8) // ld izh,e
	case 0x64: // ld izh,izh
	case 0x65:
		*iz = ((*iz & 0x00ff) << 8) | (*iz & 0x00ff) // ld izh,izl
	case 0x67:
		*iz = (uint16(z.a) << 8) | (*iz & 0x00ff) // ld izh,a
	case 0x26:
		*iz = (uint16(z.nextB()) << 8) | (*iz & 0x00ff) // ld izh,*
	case 0x68:
		*iz = (*iz & 0xff00) | uint16(z.b) // ld izl,b
	case 0x69:
		*iz = (*iz & 0xff00) | uint16(z.c) // ld izl,c
	case 0x6A:
		*iz = (*iz & 0xff00) | uint16(z.d) // ld izl,d
	case 0x6B:
		*iz = (*iz & 0xff00) | uint16(z.e) // ld izl,e
	case 0x6C:
		*iz = (*iz & 0xff00) | (*iz >> 8) // ld izl,izh
	case 0x6D: // ld izl,izl
	case 0x6F:
		*iz = (*iz & 0xff00) | uint16(z.a) // ld izl,a
	case 0x2E:
		*iz = (*iz & 0xff00) | uint16(z.nextB()) // ld izl,*
	case 0xF9:
		z.sp = *iz // ld sp,iz
	case 0xE3:
		// ex (sp),iz
		val := z.rw(z.sp)
		z.ww(z.sp, *iz)
		*iz = val
		z.memPtr = val
	case 0xCB:
		addr := z.displace(*iz, z.nextB())
		op := z.nextB()
		z.execOpcodeDcb(op, addr)
	default:
		// any other FD/DD opcode behaves as a non-prefixed opcode:
		z.execOpcode(opcode)
		// R should not be incremented twice:
		z.incR()
	}
}
