package z80

import log "github.com/sirupsen/logrus"

// jumps to an address
func (z *Z80) jump(addr uint16) {
	z.pc = addr
	z.mem_ptr = addr
}

// jumps to next word in memory if condition is true
func (z *Z80) cond_jump(condition bool) {
	addr := z.nextw()
	if condition {
		z.jump(addr)
	}
	z.mem_ptr = addr
}

// calls to next word in memory
func (z *Z80) call(addr uint16) {
	z.pushw(z.pc)
	z.pc = addr
	z.mem_ptr = addr
}

// calls to next word in memory if condition is true
func (z *Z80) cond_call(condition bool) {
	addr := z.nextw()
	if condition {
		z.call(addr)
		z.cyc += 7
	}
	z.mem_ptr = addr
}

// returns from subroutine
func (z *Z80) ret() {
	z.pc = z.popw()
	z.mem_ptr = z.pc
}

// returns from subroutine if condition is true
func (z *Z80) cond_ret(condition bool) {
	if condition {
		z.ret()
		z.cyc += 6
	}
}

func (z *Z80) jr(displacement byte) {
	z.pc += uint16(displacement)
	z.mem_ptr = z.pc
}

func (z *Z80) cond_jr(condition bool) {
	b := z.nextb()
	if condition {
		z.jr(b)
		z.cyc += 5
	}
}

func bToByte(cond bool) byte {
	if cond {
		return byte(1)
	}
	return byte(0)
}

// ADD Byte: adds two bytes together
func (z *Z80) addb(a byte, b byte, cy bool) byte {
	result := a + b + bToByte(cy)
	z.sf = result&0x80 != 0
	z.zf = result == 0
	z.hf = carry(4, uint16(a), uint16(b), cy)
	z.pf = carry(7, uint16(a), uint16(b), cy) != carry(8, uint16(a), uint16(b), cy)
	z.cf = carry(8, uint16(a), uint16(b), cy)
	z.nf = false
	z.updateXY(result)
	return result
}

// SUBtract Byte: subtracts two bytes (with optional carry)
func (z *Z80) subb(a byte, b byte, cy bool) byte {
	val := z.addb(a, ^b, !cy)
	z.cf = !z.cf
	z.hf = !z.hf
	z.nf = true
	return val
}

// ADD Word: adds two words together
func (z *Z80) addw(a uint16, b uint16, cy bool) uint16 {
	lsb := z.addb(byte(a), byte(b), cy)
	msb := z.addb(byte(a>>8), byte(b>>8), z.cf)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.zf = result == 0
	z.mem_ptr = a + 1
	return result
}

// SUBtract Word: subtracts two words (with optional carry)
func (z *Z80) subw(a uint16, b uint16, cy bool) uint16 {
	lsb := z.subb(byte(a), byte(b), cy)
	msb := z.subb(byte(a>>8), byte(b>>8), z.cf)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.zf = result == 0
	z.mem_ptr = a + 1
	return result
}

// Adds a word to HL
func (z *Z80) addhl(val uint16) {
	sf := z.sf
	zf := z.zf
	pf := z.pf
	result := z.addw(z.get_hl(), val, false)
	z.set_hl(result)
	z.sf = sf
	z.zf = zf
	z.pf = pf
}

// adds a word to IX or IY
func (z *Z80) addiz(reg *uint16, val uint16) {
	sf := z.sf
	zf := z.zf
	pf := z.pf
	result := z.addw(*reg, val, false)
	*reg = result
	z.sf = sf
	z.zf = zf
	z.pf = pf
}

// adchl adds a word (+ carry) to HL
func (z *Z80) adchl(val uint16) {
	result := z.addw(z.get_hl(), val, z.cf)
	z.sf = result&0x8000 != 0
	z.zf = result == 0
	z.set_hl(result)
}

// sbchl subtracts a word (+ carry) to HL
func (z *Z80) sbchl(val uint16) {
	result := z.subw(z.get_hl(), val, z.cf)
	z.sf = result&0x8000 != 0
	z.zf = result == 0
	z.set_hl(result)
}

// increments a byte value
func (z *Z80) inc(a byte) byte {
	cf := z.cf
	result := z.addb(a, 1, false)
	z.cf = cf
	return result
}

// decrements a byte value
func (z *Z80) dec(a byte) byte {
	cf := z.cf
	result := z.subb(a, 1, false)
	z.cf = cf
	return result
}

// executes a logic "and" between register A and a byte, then stores the
// result in register A
func (z *Z80) land(val byte) {
	result := z.a & val
	z.sf = result&0x80 != 0
	z.zf = result == 0
	z.hf = true
	z.pf = parity(result)
	z.nf = false
	z.cf = false
	z.updateXY(result)
	z.a = result
}

// executes a logic "xor" between register A and a byte, then stores the
// result in register A
func (z *Z80) lxor(val byte) {
	result := z.a ^ val
	z.sf = result&0x80 != 0
	z.zf = result == 0
	z.hf = false
	z.pf = parity(result)
	z.nf = false
	z.cf = false
	z.updateXY(result)
	z.a = result
}

// executes a logic "or" between register A and a byte, then stores the
// result in register A
func (z *Z80) lor(val byte) {
	result := z.a | val

	z.sf = result&0x80 != 0
	z.zf = result == 0
	z.hf = false
	z.pf = parity(result)
	z.nf = false
	z.cf = false
	z.updateXY(result)
	z.a = result
}

// compares a value with register A
func (z *Z80) cp(val byte) {
	z.subb(z.a, val, false)

	// the only difference between cp and sub is that
	// the xf/yf are taken from the value to be subtracted,
	// not the result

	z.updateXY(val)
}

// 0xCB opcodes
// rotate left with carry
func (z *Z80) cb_rlc(val byte) byte {
	old := val >> 7
	val = (val << 1) | old
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.pf = parity(val)
	z.nf = false
	z.hf = false
	z.cf = old != 0
	z.updateXY(val)
	return val
}

// rotate right with carry
func (z *Z80) cb_rrc(val byte) byte {
	old := val & 1
	val = (val >> 1) | (old << 7)
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.cf = old != 0
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// rotate left (simple)
func (z *Z80) cb_rl(val byte) byte {
	cf := z.cf
	z.cf = val>>7 != 0
	val = (val << 1) | bToByte(cf)
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// rotate right (simple)
func (z *Z80) cb_rr(val byte) byte {
	c := z.cf
	z.cf = (val & 1) != 0
	val = (val >> 1) | (bToByte(c) << 7)
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// shift left preserving sign
func (z *Z80) cb_sla(val byte) byte {
	z.cf = (val >> 7) != 0
	val <<= 1
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// SLL (exactly like SLA, but sets the first bit to 1)
func (z *Z80) cb_sll(val byte) byte {
	z.cf = val&0x80 != 0
	val <<= 1
	val |= 1
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// shift right preserving sign
func (z *Z80) cb_sra(val byte) byte {
	z.cf = (val & 1) != 0
	val = (val >> 1) | (val & 0x80) // 0b10000000
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// shift register right
func (z *Z80) cb_srl(val byte) byte {
	z.cf = (val & 1) != 0
	val >>= 1
	z.sf = val&0x80 != 0
	z.zf = val == 0
	z.nf = false
	z.hf = false
	z.pf = parity(val)
	z.updateXY(val)
	return val
}

// tests bit "n" from a byte
func (z *Z80) cb_bit(val byte, n byte) byte {
	result := val & (1 << n)
	z.sf = result&0x80 != 0
	z.zf = result == 0
	z.hf = true
	z.updateXY(val)
	z.pf = z.zf
	z.nf = false
	return result
}

func (z *Z80) ldi() {
	de := z.get_de()
	hl := z.get_hl()
	val := z.rb(hl)
	z.wb(de, val)

	z.set_hl(z.get_hl() + 1)
	z.set_de(z.get_de() + 1)
	z.set_bc(z.get_bc() - 1)

	// see https://wikiti.brandonw.net/index.php?title=Z80_Instruction_Set
	// for the calculation of xf/yf on LDI
	result := val + z.a

	z.updateXY(result)

	z.nf = false
	z.hf = false
	z.pf = z.get_bc() > 0

}

func (z *Z80) ldd() {
	z.ldi()
	// same as ldi but HL and DE are decremented instead of incremented
	z.set_hl(z.get_hl() - 2)
	z.set_de(z.get_de() - 2)
}

func (z *Z80) cpi() {
	cf := z.cf
	result := z.subb(z.a, z.rb(z.get_hl()), false)
	z.set_hl(z.get_hl() + 1)
	z.set_bc(z.get_bc() - 1)

	val := result - bToByte(z.hf)
	z.xf = val&0x08 != 0
	z.yf = val&0x02 != 0
	z.pf = z.get_bc() != 0
	z.cf = cf
	z.mem_ptr += 1
}

func (z *Z80) cpd() {
	z.cpi()
	// same as cpi but HL is decremented instead of incremented
	z.set_hl(z.get_hl() - 2)
	z.mem_ptr -= 2
}

func (z *Z80) in_r_c(r *byte) {
	*r = z.core.IORead(z.get_bc())
	z.zf = *r == 0
	z.sf = *r&0x80 != 0
	z.pf = parity(*r)
	z.nf = false
	z.hf = false
}

func (z *Z80) ini() {
	val := z.core.IORead(z.get_bc())
	z.wb(z.get_hl(), val)
	z.set_hl(z.get_hl() + 1)
	z.b -= 1
	z.zf = z.b == 0
	z.nf = true
	z.mem_ptr = z.get_bc() + 1
}

func (z *Z80) ind() {
	z.ini()
	z.set_hl(z.get_hl() - 2)
	z.mem_ptr = z.get_bc() - 2
}

func (z *Z80) outi() {
	z.core.IOWrite(z.get_bc(), z.rb(z.get_hl()))
	z.set_hl(z.get_hl() + 1)
	z.b -= 1
	z.zf = z.b == 0
	z.nf = true
	z.mem_ptr = z.get_bc() + 1
}

func (z *Z80) outd() {
	z.outi()
	z.set_hl(z.get_hl() - 2)
	z.mem_ptr = z.get_bc() - 2
}

func (z *Z80) daa() {
	// "When this instruction is executed, the A register is BCD corrected
	// using the  contents of the flags. The exact process is the following:
	// if the least significant four bits of A contain a non-BCD digit
	// (i. e. it is greater than 9) or the H flag is set, then $06 is
	// added to the register. Then the four most significant bits are
	// checked. If this more significant digit also happens to be greater
	// than 9 or the C flag is set, then $60 is added."
	// > http://z80-heaven.wikidot.com/instructions-set:daa
	correction := byte(0)

	if (z.a&0x0F) > 0x09 || z.hf {
		correction += 0x06
	}

	if z.a > 0x99 || z.cf {
		correction += 0x60
		z.cf = true
	}

	substraction := z.nf
	if substraction {
		z.hf = z.hf && (z.a&0x0F) < 0x06
		z.a -= correction
	} else {
		z.hf = (z.a & 0x0F) > 0x09
		z.a += correction
	}

	z.sf = z.a&0x80 != 0
	z.zf = z.a == 0
	z.pf = parity(z.a)
	z.updateXY(z.a)
}

func (z *Z80) displace(base_addr uint16, displacement byte) uint16 {
	addr := base_addr + uint16(displacement)
	z.mem_ptr = addr
	return addr
}

func (z *Z80) process_interrupts() {
	// "When an EI instruction is executed, any pending interrupt request
	// is not accepted until after the instruction following EI is executed."
	if z.iff_delay > 0 {
		z.iff_delay -= 1
		if z.iff_delay == 0 {
			z.iff1 = true
			z.iff2 = true
		}
		return
	}

	if z.nmi_pending {
		z.nmi_pending = false
		z.halted = false
		z.iff1 = false
		z.inc_r()

		z.cyc += 11
		z.call(0x0066)
		return
	}

	if z.int_pending && z.iff1 {
		z.int_pending = false
		z.halted = false
		z.iff1 = false
		z.iff2 = false
		z.inc_r()

		switch z.interrupt_mode {
		case 0:
			z.cyc += 11
			z.exec_opcode(z.int_data)
		case 1:
			z.cyc += 13
			z.call(0x38)
		case 2:
			z.cyc += 19
			z.call(z.rw((uint16(z.i) << 8) | uint16(z.int_data)))
		default:
			log.Errorf("Unsupported interrupt mode %d\n", z.interrupt_mode)
		}
		return
	}
}

// z80_gen_nmi function to call when an NMI is to be serviced
func (z *Z80) z80_gen_nmi() {
	z.nmi_pending = true
}

// z80_gen_int function to call when an INT is to be serviced
func (z *Z80) z80_gen_int(data byte) {
	z.int_pending = true
	z.int_data = data
}

// executes a non-prefixed opcode
func (z *Z80) exec_opcode(opcode byte) {
	z.cyc += uint64(cyc_00[opcode])
	z.inc_r()

	switch opcode {
	case 0x7F:
		z.a = z.a // ld a,a
	case 0x78:
		z.a = z.b // ld a,b
	case 0x79:
		z.a = z.c // ld a,c
	case 0x7A:
		z.a = z.d // ld a,d
	case 0x7B:
		z.a = z.e // ld a,e
	case 0x7C:
		z.a = z.h // ld a,h
	case 0x7D:
		z.a = z.l // ld a,l

	case 0x47:
		z.b = z.a // ld b,a
	case 0x40:
		z.b = z.b // ld b,b
	case 0x41:
		z.b = z.c // ld b,c
	case 0x42:
		z.b = z.d // ld b,d
	case 0x43:
		z.b = z.e // ld b,e
	case 0x44:
		z.b = z.h // ld b,h
	case 0x45:
		z.b = z.l // ld b,l

	case 0x4F:
		z.c = z.a // ld c,a
	case 0x48:
		z.c = z.b // ld c,b
	case 0x49:
		z.c = z.c // ld c,c
	case 0x4A:
		z.c = z.d // ld c,d
	case 0x4B:
		z.c = z.e // ld c,e
	case 0x4C:
		z.c = z.h // ld c,h
	case 0x4D:
		z.c = z.l // ld c,l

	case 0x57:
		z.d = z.a // ld d,a
	case 0x50:
		z.d = z.b // ld d,b
	case 0x51:
		z.d = z.c // ld d,c
	case 0x52:
		z.d = z.d // ld d,d
	case 0x53:
		z.d = z.e // ld d,e
	case 0x54:
		z.d = z.h // ld d,h
	case 0x55:
		z.d = z.l // ld d,l

	case 0x5F:
		z.e = z.a // ld e,a
	case 0x58:
		z.e = z.b // ld e,b
	case 0x59:
		z.e = z.c // ld e,c
	case 0x5A:
		z.e = z.d // ld e,d
	case 0x5B:
		z.e = z.e // ld e,e
	case 0x5C:
		z.e = z.h // ld e,h
	case 0x5D:
		z.e = z.l // ld e,l

	case 0x67:
		z.h = z.a // ld h,a
	case 0x60:
		z.h = z.b // ld h,b
	case 0x61:
		z.h = z.c // ld h,c
	case 0x62:
		z.h = z.d // ld h,d
	case 0x63:
		z.h = z.e // ld h,e
	case 0x64:
		z.h = z.h // ld h,h
	case 0x65:
		z.h = z.l // ld h,l

	case 0x6F:
		z.l = z.a // ld l,a
	case 0x68:
		z.l = z.b // ld l,b
	case 0x69:
		z.l = z.c // ld l,c
	case 0x6A:
		z.l = z.d // ld l,d
	case 0x6B:
		z.l = z.e // ld l,e
	case 0x6C:
		z.l = z.h // ld l,h
	case 0x6D:
		z.l = z.l // ld l,l

	case 0x7E:
		z.a = z.rb(z.get_hl()) // ld a,(hl)
	case 0x46:
		z.b = z.rb(z.get_hl()) // ld b,(hl)
	case 0x4E:
		z.c = z.rb(z.get_hl()) // ld c,(hl)
	case 0x56:
		z.d = z.rb(z.get_hl()) // ld d,(hl)
	case 0x5E:
		z.e = z.rb(z.get_hl()) // ld e,(hl)
	case 0x66:
		z.h = z.rb(z.get_hl()) // ld h,(hl)
	case 0x6E:
		z.l = z.rb(z.get_hl()) // ld l,(hl)

	case 0x77:
		z.wb(z.get_hl(), z.a) // ld (hl),a
	case 0x70:
		z.wb(z.get_hl(), z.b) // ld (hl),b
	case 0x71:
		z.wb(z.get_hl(), z.c) // ld (hl),c
	case 0x72:
		z.wb(z.get_hl(), z.d) // ld (hl),d
	case 0x73:
		z.wb(z.get_hl(), z.e) // ld (hl),e
	case 0x74:
		z.wb(z.get_hl(), z.h) // ld (hl),h
	case 0x75:
		z.wb(z.get_hl(), z.l) // ld (hl),l

	case 0x3E:
		z.a = z.nextb() // ld a,*
	case 0x06:
		z.b = z.nextb() // ld b,*
	case 0x0E:
		z.c = z.nextb() // ld c,*
	case 0x16:
		z.d = z.nextb() // ld d,*
	case 0x1E:
		z.e = z.nextb() // ld e,*
	case 0x26:
		z.h = z.nextb() // ld h,*
	case 0x2E:
		z.l = z.nextb() // ld l,*
	case 0x36:
		z.wb(z.get_hl(), z.nextb()) // ld (hl),*
	case 0x0A:
		// ld a,(bc)
		z.a = z.rb(z.get_bc())
		z.mem_ptr = z.get_bc() + 1
	case 0x1A:
		// ld a,(de)
		z.a = z.rb(z.get_de())
		z.mem_ptr = z.get_de() + 1
	case 0x3A:
		// ld a,(**)
		addr := z.nextw()
		z.a = z.rb(addr)
		z.mem_ptr = addr + 1
	case 0x02:
		// ld (bc),a
		z.wb(z.get_bc(), z.a)
		z.mem_ptr = (uint16(z.a) << 8) | ((z.get_bc() + 1) & 0xFF)
	case 0x12:
		// ld (de),a
		z.wb(z.get_de(), z.a)
		z.mem_ptr = (uint16(z.a) << 8) | ((z.get_de() + 1) & 0xFF)
	case 0x32:
		// ld (**),a
		addr := z.nextw()
		z.wb(addr, z.a)
		z.mem_ptr = (uint16(z.a) << 8) | ((addr + 1) & 0xFF)
	case 0x01:
		z.set_bc(z.nextw()) // ld bc,**
	case 0x11:
		z.set_de(z.nextw()) // ld de,**
	case 0x21:
		z.set_hl(z.nextw()) // ld hl,**
	case 0x31:
		z.sp = z.nextw() // ld sp,**

	case 0x2A:
		// ld hl,(**)
		addr := z.nextw()
		z.set_hl(z.rw(addr))
		z.mem_ptr = addr + 1
	case 0x22:
		// ld (**),hl
		addr := z.nextw()
		z.ww(addr, z.get_hl())
		z.mem_ptr = addr + 1
	case 0xF9:
		z.sp = z.get_hl() // ld sp,hl

	case 0xEB:
		// ex de,hl
		de := z.get_de()
		z.set_de(z.get_hl())
		z.set_hl(de)
	case 0xE3:
		// ex (sp),hl
		val := z.rw(z.sp)
		z.ww(z.sp, z.get_hl())
		z.set_hl(val)
		z.mem_ptr = val
	case 0x87:
		z.a = z.addb(z.a, z.a, false) // add a,a
	case 0x80:
		z.a = z.addb(z.a, z.b, false) // add a,b
	case 0x81:
		z.a = z.addb(z.a, z.c, false) // add a,c
	case 0x82:
		z.a = z.addb(z.a, z.d, false) // add a,d
	case 0x83:
		z.a = z.addb(z.a, z.e, false) // add a,e
	case 0x84:
		z.a = z.addb(z.a, z.h, false) // add a,h
	case 0x85:
		z.a = z.addb(z.a, z.l, false) // add a,l
	case 0x86:
		z.a = z.addb(z.a, z.rb(z.get_hl()), false) // add a,(hl)
	case 0xC6:
		z.a = z.addb(z.a, z.nextb(), false) // add a,*

	case 0x8F:
		z.a = z.addb(z.a, z.a, z.cf) // adc a,a
	case 0x88:
		z.a = z.addb(z.a, z.b, z.cf) // adc a,b
	case 0x89:
		z.a = z.addb(z.a, z.c, z.cf) // adc a,c
	case 0x8A:
		z.a = z.addb(z.a, z.d, z.cf) // adc a,d
	case 0x8B:
		z.a = z.addb(z.a, z.e, z.cf) // adc a,e
	case 0x8C:
		z.a = z.addb(z.a, z.h, z.cf) // adc a,h
	case 0x8D:
		z.a = z.addb(z.a, z.l, z.cf) // adc a,l
	case 0x8E:
		z.a = z.addb(z.a, z.rb(z.get_hl()), z.cf) // adc a,(hl)
	case 0xCE:
		z.a = z.addb(z.a, z.nextb(), z.cf) // adc a,*

	case 0x97:
		z.a = z.subb(z.a, z.a, false) // sub a,a
	case 0x90:
		z.a = z.subb(z.a, z.b, false) // sub a,b
	case 0x91:
		z.a = z.subb(z.a, z.c, false) // sub a,c
	case 0x92:
		z.a = z.subb(z.a, z.d, false) // sub a,d
	case 0x93:
		z.a = z.subb(z.a, z.e, false) // sub a,e
	case 0x94:
		z.a = z.subb(z.a, z.h, false) // sub a,h
	case 0x95:
		z.a = z.subb(z.a, z.l, false) // sub a,l
	case 0x96:
		z.a = z.subb(z.a, z.rb(z.get_hl()), false) // sub a,(hl)
	case 0xD6:
		z.a = z.subb(z.a, z.nextb(), false) // sub a,*

	case 0x9F:
		z.a = z.subb(z.a, z.a, z.cf) // sbc a,a
	case 0x98:
		z.a = z.subb(z.a, z.b, z.cf) // sbc a,b
	case 0x99:
		z.a = z.subb(z.a, z.c, z.cf) // sbc a,c
	case 0x9A:
		z.a = z.subb(z.a, z.d, z.cf) // sbc a,d
	case 0x9B:
		z.a = z.subb(z.a, z.e, z.cf) // sbc a,e
	case 0x9C:
		z.a = z.subb(z.a, z.h, z.cf) // sbc a,h
	case 0x9D:
		z.a = z.subb(z.a, z.l, z.cf) // sbc a,l
	case 0x9E:
		z.a = z.subb(z.a, z.rb(z.get_hl()), z.cf) // sbc a,(hl)
	case 0xDE:
		z.a = z.subb(z.a, z.nextb(), z.cf) // sbc a,*

	case 0x09:
		z.addhl(z.get_bc()) // add hl,bc
	case 0x19:
		z.addhl(z.get_de()) // add hl,de
	case 0x29:
		z.addhl(z.get_hl()) // add hl,hl
	case 0x39:
		z.addhl(z.sp) // add hl,sp

	case 0xF3:
		z.iff1 = false
		z.iff2 = false // di
	case 0xFB:
		z.iff_delay = 1 // ei
	case 0x00: // nop
	case 0x76:
		z.halted = true // halt

	case 0x3C:
		z.a = z.inc(z.a) // inc a
	case 0x04:
		z.b = z.inc(z.b) // inc b
	case 0x0C:
		z.c = z.inc(z.c) // inc c
	case 0x14:
		z.d = z.inc(z.d) // inc d
	case 0x1C:
		z.e = z.inc(z.e) // inc e
	case 0x24:
		z.h = z.inc(z.h) // inc h
	case 0x2C:
		z.l = z.inc(z.l) // inc l
	case 0x34:
		// inc (hl)
		result := z.inc(z.rb(z.get_hl()))
		z.wb(z.get_hl(), result)
	case 0x3D:
		z.a = z.dec(z.a) // dec a
	case 0x05:
		z.b = z.dec(z.b) // dec b
	case 0x0D:
		z.c = z.dec(z.c) // dec c
	case 0x15:
		z.d = z.dec(z.d) // dec d
	case 0x1D:
		z.e = z.dec(z.e) // dec e
	case 0x25:
		z.h = z.dec(z.h) // dec h
	case 0x2D:
		z.l = z.dec(z.l) // dec l
	case 0x35:
		// dec (hl)
		result := z.dec(z.rb(z.get_hl()))
		z.wb(z.get_hl(), result)
	case 0x03:
		z.set_bc(z.get_bc() + 1) // inc bc
	case 0x13:
		z.set_de(z.get_de() + 1) // inc de
	case 0x23:
		z.set_hl(z.get_hl() + 1) // inc hl
	case 0x33:
		z.sp = z.sp + 1 // inc sp
	case 0x0B:
		z.set_bc(z.get_bc() - 1) // dec bc
	case 0x1B:
		z.set_de(z.get_de() - 1) // dec de
	case 0x2B:
		z.set_hl(z.get_hl() - 1) // dec hl
	case 0x3B:
		z.sp = z.sp - 1 // dec sp
	case 0x27:
		z.daa() // daa
	case 0x2F:
		// cpl
		z.a = ^z.a
		z.nf = true
		z.hf = true
		z.updateXY(z.a)
	case 0x37:
		// scf
		z.cf = true
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
	case 0x3F:
		// ccf
		z.hf = z.cf
		z.cf = !z.cf
		z.nf = false
		z.updateXY(z.a)
	case 0x07:
		// rlca (rotate left)
		z.cf = z.a&0x80 != 0
		z.a = (z.a << 1) | bToByte(z.cf)
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
	case 0x0F:
		// rrca (rotate right)
		z.cf = z.a&1 != 0
		z.a = (z.a >> 1) | (bToByte(z.cf) << 7)
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
	case 0x17:
		// rla
		cy := bToByte(z.cf)
		z.cf = z.a&0x80 != 0
		z.a = (z.a << 1) | cy
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
	case 0x1F:
		// rra
		cy := bToByte(z.cf)
		z.cf = z.a&1 != 0
		z.a = (z.a >> 1) | (cy << 7)
		z.nf = false
		z.hf = false
		z.updateXY(z.a)
	case 0xA7:
		z.land(z.a) // and a
	case 0xA0:
		z.land(z.b) // and b
	case 0xA1:
		z.land(z.c) // and c
	case 0xA2:
		z.land(z.d) // and d
	case 0xA3:
		z.land(z.e) // and e
	case 0xA4:
		z.land(z.h) // and h
	case 0xA5:
		z.land(z.l) // and l
	case 0xA6:
		z.land(z.rb(z.get_hl())) // and (hl)
	case 0xE6:
		z.land(z.nextb()) // and *

	case 0xAF:
		z.lxor(z.a) // xor a
	case 0xA8:
		z.lxor(z.b) // xor b
	case 0xA9:
		z.lxor(z.c) // xor c
	case 0xAA:
		z.lxor(z.d) // xor d
	case 0xAB:
		z.lxor(z.e) // xor e
	case 0xAC:
		z.lxor(z.h) // xor h
	case 0xAD:
		z.lxor(z.l) // xor l
	case 0xAE:
		z.lxor(z.rb(z.get_hl())) // xor (hl)
	case 0xEE:
		z.lxor(z.nextb()) // xor *

	case 0xB7:
		z.lor(z.a) // or a
	case 0xB0:
		z.lor(z.b) // or b
	case 0xB1:
		z.lor(z.c) // or c
	case 0xB2:
		z.lor(z.d) // or d
	case 0xB3:
		z.lor(z.e) // or e
	case 0xB4:
		z.lor(z.h) // or h
	case 0xB5:
		z.lor(z.l) // or l
	case 0xB6:
		z.lor(z.rb(z.get_hl())) // or (hl)
	case 0xF6:
		z.lor(z.nextb()) // or *

	case 0xBF:
		z.cp(z.a) // cp a
	case 0xB8:
		z.cp(z.b) // cp b
	case 0xB9:
		z.cp(z.c) // cp c
	case 0xBA:
		z.cp(z.d) // cp d
	case 0xBB:
		z.cp(z.e) // cp e
	case 0xBC:
		z.cp(z.h) // cp h
	case 0xBD:
		z.cp(z.l) // cp l
	case 0xBE:
		z.cp(z.rb(z.get_hl())) // cp (hl)
	case 0xFE:
		z.cp(z.nextb()) // cp *

	case 0xC3:
		z.jump(z.nextw()) // jm **
	case 0xC2:
		z.cond_jump(!z.zf) // jp nz, **
	case 0xCA:
		z.cond_jump(z.zf) // jp z, **
	case 0xD2:
		z.cond_jump(!z.cf) // jp nc, **
	case 0xDA:
		z.cond_jump(z.cf) // jp c, **
	case 0xE2:
		z.cond_jump(!z.pf) // jp po, **
	case 0xEA:
		z.cond_jump(z.pf) // jp pe, **
	case 0xF2:
		z.cond_jump(!z.sf) // jp p, **
	case 0xFA:
		z.cond_jump(z.sf) // jp m, **

	case 0x10:
		z.b--
		z.cond_jr(z.b != 0) // djnz *
	case 0x18:
		z.pc += uint16(z.nextb()) // jr *
	case 0x20:
		z.cond_jr(!z.zf) // jr nz, *
	case 0x28:
		z.cond_jr(z.zf) // jr z, *
	case 0x30:
		z.cond_jr(!z.cf) // jr nc, *
	case 0x38:
		z.cond_jr(z.cf) // jr c, *

	case 0xE9:
		z.pc = z.get_hl() // jp (hl)
	case 0xCD:
		z.call(z.nextw()) // call

	case 0xC4:
		z.cond_call(!z.zf) // cnz
	case 0xCC:
		z.cond_call(z.zf) // cz
	case 0xD4:
		z.cond_call(!z.cf) // cnc
	case 0xDC:
		z.cond_call(z.cf) // cc
	case 0xE4:
		z.cond_call(!z.pf) // cpo
	case 0xEC:
		z.cond_call(z.pf) // cpe
	case 0xF4:
		z.cond_call(!z.sf) // cp
	case 0xFC:
		z.cond_call(z.sf) // cm

	case 0xC9:
		z.ret() // ret
	case 0xC0:
		z.cond_ret(!z.zf) // ret nz
	case 0xC8:
		z.cond_ret(z.zf) // ret z
	case 0xD0:
		z.cond_ret(!z.cf) // ret nc
	case 0xD8:
		z.cond_ret(z.cf) // ret c
	case 0xE0:
		z.cond_ret(!z.pf) // ret po
	case 0xE8:
		z.cond_ret(z.pf) // ret pe
	case 0xF0:
		z.cond_ret(!z.sf) // ret p
	case 0xF8:
		z.cond_ret(z.sf) // ret m

	case 0xC7:
		z.call(0x00) // rst 0
	case 0xCF:
		z.call(0x08) // rst 1
	case 0xD7:
		z.call(0x10) // rst 2
	case 0xDF:
		z.call(0x18) // rst 3
	case 0xE7:
		z.call(0x20) // rst 4
	case 0xEF:
		z.call(0x28) // rst 5
	case 0xF7:
		z.call(0x30) // rst 6
	case 0xFF:
		z.call(0x38) // rst 7

	case 0xC5:
		z.pushw(z.get_bc()) // push bc
	case 0xD5:
		z.pushw(z.get_de()) // push de
	case 0xE5:
		z.pushw(z.get_hl()) // push hl
	case 0xF5:
		z.pushw((uint16(z.a) << 8) | uint16(z.get_f())) // push af

	case 0xC1:
		z.set_bc(z.popw()) // pop bc
	case 0xD1:
		z.set_de(z.popw()) // pop de
	case 0xE1:
		z.set_hl(z.popw()) // pop hl
	case 0xF1:
		// pop af
		val := z.popw()
		z.a = byte(val >> 8)
		z.set_f(byte(val))
	case 0xDB:
		// in a,(n)
		port := uint16(z.nextb())
		a := z.a
		z.a = z.core.IORead(port)
		z.mem_ptr = (uint16(a) << 8) | uint16(z.a+1)
	case 0xD3:
		// out (n), a
		port := uint16(z.nextb())
		z.core.IOWrite(port, z.a)
		z.mem_ptr = (port + 1) | (uint16(z.a) << 8)
	case 0x08:
		// ex af,af'
		a := z.a
		f := z.get_f()

		z.a = z.a_
		z.set_f(z.f_)

		z.a_ = a
		z.f_ = f
	case 0xD9:
		// exx
		b := z.b
		c := z.c
		d := z.d
		e := z.e
		h := z.h
		l := z.l

		z.b = z.b_
		z.c = z.c_
		z.d = z.d_
		z.e = z.e_
		z.h = z.h_
		z.l = z.l_

		z.b_ = b
		z.c_ = c
		z.d_ = d
		z.e_ = e
		z.h_ = h
		z.l_ = l
	case 0xCB:
		z.exec_opcode_cb(z.nextb())
	case 0xED:
		z.exec_opcode_ed(z.nextb())
	case 0xDD:
		z.exec_opcode_ddfd(z.nextb(), &z.ix)
	case 0xFD:
		z.exec_opcode_ddfd(z.nextb(), &z.iy)

	default:
		log.Errorf("Unknown opcode %02X\n", opcode)
	}
}
