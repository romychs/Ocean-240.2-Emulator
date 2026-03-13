package c99

import log "github.com/sirupsen/logrus"

// jumps to an address
func (z *Z80) jump(addr uint16) {
	z.pc = addr
	z.memPtr = addr
}

// jumps to next word in memory if condition is true
func (z *Z80) condJump(condition bool) {
	addr := z.nextW()
	if condition {
		z.jump(addr)
	}
	z.memPtr = addr
}

// calls to next word in memory
func (z *Z80) call(addr uint16) {
	z.pushW(z.pc)
	z.pc = addr
	z.memPtr = addr
}

// calls to next word in memory if condition is true
func (z *Z80) condCall(condition bool) {
	addr := z.nextW()
	if condition {
		z.call(addr)
		z.cycleCount += 7
	}
	z.memPtr = addr
}

// returns from subroutine
func (z *Z80) ret() {
	z.pc = z.popW()
	z.memPtr = z.pc
}

// returns from subroutine if condition is true
func (z *Z80) condRet(condition bool) {
	if condition {
		z.ret()
		z.cycleCount += 6
	}
}

func (z *Z80) jr(offset byte) {
	if offset&0x80 != 0 {
		z.pc += 0xFF00 | uint16(offset)
	} else {
		z.pc += uint16(offset)
	}
	z.memPtr = z.pc
}

func (z *Z80) condJr(condition bool) {
	b := z.nextB()
	if condition {
		z.jr(b)
		z.cycleCount += 5
	}
}

func bToByte(cond bool) byte {
	if cond {
		return byte(1)
	}
	return byte(0)
}

// ADD Byte: adds two bytes together
func (z *Z80) addB(a byte, b byte, cy bool) byte {
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
func (z *Z80) subB(a byte, b byte, cy bool) byte {
	val := z.addB(a, ^b, !cy)
	z.cf = !z.cf
	z.hf = !z.hf
	z.nf = true
	return val
}

// ADD Word: adds two words together
func (z *Z80) addW(a uint16, b uint16, cy bool) uint16 {
	lsb := z.addB(byte(a), byte(b), cy)
	msb := z.addB(byte(a>>8), byte(b>>8), z.cf)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.zf = result == 0
	z.memPtr = a + 1
	return result
}

// SUBtract Word: subtracts two words (with optional carry)
func (z *Z80) subW(a uint16, b uint16, cy bool) uint16 {
	lsb := z.subB(byte(a), byte(b), cy)
	msb := z.subB(byte(a>>8), byte(b>>8), z.cf)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.zf = result == 0
	z.memPtr = a + 1
	return result
}

// Adds a word to HL
func (z *Z80) addHL(val uint16) {
	sf := z.sf
	zf := z.zf
	pf := z.pf
	result := z.addW(z.hl(), val, false)
	z.setHL(result)
	z.sf = sf
	z.zf = zf
	z.pf = pf
}

// adds a word to IX or IY
func (z *Z80) addIZ(reg *uint16, val uint16) {
	sf := z.sf
	zf := z.zf
	pf := z.pf
	result := z.addW(*reg, val, false)
	*reg = result
	z.sf = sf
	z.zf = zf
	z.pf = pf
}

// adcHL adds a word (+ carry) to HL
func (z *Z80) adcHL(val uint16) {
	result := z.addW(z.hl(), val, z.cf)
	z.sf = result&0x8000 != 0
	z.zf = result == 0
	z.setHL(result)
}

// sbcHL subtracts a word (+ carry) to HL
func (z *Z80) sbcHL(val uint16) {
	result := z.subW(z.hl(), val, z.cf)
	z.sf = result&0x8000 != 0
	z.zf = result == 0
	z.setHL(result)
}

// increments a byte value
func (z *Z80) inc(a byte) byte {
	cf := z.cf
	result := z.addB(a, 1, false)
	z.cf = cf
	return result
}

// decrements a byte value
func (z *Z80) dec(a byte) byte {
	cf := z.cf
	result := z.subB(a, 1, false)
	z.cf = cf
	return result
}

// executes a logic "and" between register A and a byte, then stores the
// result in register A
func (z *Z80) lAnd(val byte) {
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
func (z *Z80) lXor(val byte) {
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
func (z *Z80) lOr(val byte) {
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
	z.subB(z.a, val, false)

	// the only difference between cp and sub is that
	// the xf/yf are taken from the value to be subtracted,
	// not the result

	z.updateXY(val)
}

// 0xCB opcodes
// rotate left with carry
func (z *Z80) cbRlc(val byte) byte {
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
func (z *Z80) cbRrc(val byte) byte {
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
func (z *Z80) cbRl(val byte) byte {
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
func (z *Z80) cbRr(val byte) byte {
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
func (z *Z80) cbSla(val byte) byte {
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
func (z *Z80) cbSll(val byte) byte {
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
func (z *Z80) cbSra(val byte) byte {
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
func (z *Z80) cbSrl(val byte) byte {
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
func (z *Z80) cbBit(val byte, n byte) byte {
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
	de := z.de()
	hl := z.hl()
	val := z.rb(hl)
	z.wb(de, val)

	z.setHL(z.hl() + 1)
	z.setDE(z.de() + 1)
	z.setBC(z.bc() - 1)

	// see https://wikiti.brandonw.net/index.php?title=Z80_Instruction_Set
	// for the calculation of xf/yf on LDI
	result := val + z.a

	z.xf = result&0x08 != 0 // bit 3
	z.yf = result&0x02 != 0 // bit 1

	z.nf = false
	z.hf = false
	z.pf = z.bc() > 0

}

func (z *Z80) ldd() {
	z.ldi()
	// same as ldi but HL and DE are decremented instead of incremented
	z.setHL(z.hl() - 2)
	z.setDE(z.de() - 2)
}

func (z *Z80) cpi() {
	cf := z.cf
	result := z.subB(z.a, z.rb(z.hl()), false)
	z.setHL(z.hl() + 1)
	z.setBC(z.bc() - 1)

	val := result - bToByte(z.hf)
	z.xf = val&0x08 != 0
	z.yf = val&0x02 != 0
	z.pf = z.bc() != 0
	z.cf = cf
	z.memPtr += 1
}

func (z *Z80) cpd() {
	z.cpi()
	// same as cpi but HL is decremented instead of incremented
	z.setHL(z.hl() - 2)
	z.memPtr -= 2
}

func (z *Z80) inRC(r *byte) {
	*r = z.core.IORead(z.bc())
	z.zf = *r == 0
	z.sf = *r&0x80 != 0
	z.pf = parity(*r)
	z.nf = false
	z.hf = false
}

func (z *Z80) ini() {
	val := z.core.IORead(z.bc())
	z.wb(z.hl(), val)
	z.memPtr = z.bc() + 1
	z.b--

	other := val + z.c + 1
	if other < val {
		z.hf = true
		z.cf = true
	} else {
		z.hf = false
		z.cf = false
	}
	z.nf = val&0x80 != 0
	z.pf = parity((other & 0x07) ^ z.b)
	z.sf = z.b&0x80 != 0
	z.zf = z.b == 0
	z.updateXY(z.b)
	z.setHL(z.hl() + 1)
}

func (z *Z80) ind() {
	val := z.core.IORead(z.bc())
	z.wb(z.hl(), val)
	z.memPtr = z.bc() - 1
	z.b--

	other := val + z.c - 1
	z.nf = val&0x80 != 0
	if other < val {
		z.hf = true
		z.cf = true
	} else {
		z.hf = false
		z.cf = false
	}
	z.pf = parity((other & 0x07) ^ z.b)

	z.sf = z.b&0x80 != 0
	z.zf = z.b == 0
	z.updateXY(z.b)
	z.setHL(z.hl() - 1)
}

func (z *Z80) outi() {
	val := z.rb(z.hl())
	z.b--
	z.memPtr = z.bc() + 1
	z.core.IOWrite(z.bc(), val)
	z.setHL(z.hl() + 1)
	other := val + z.l
	z.nf = val&0x80 != 0
	if other < val {
		z.hf = true
		z.cf = true
	} else {
		z.hf = false
		z.cf = false
	}
	z.pf = parity((other & 0x07) ^ z.b)
	z.zf = z.b == 0
	z.sf = z.b&0x80 != 0
	z.updateXY(z.b)
}

func (z *Z80) outd() {
	val := z.rb(z.hl())
	z.b--
	z.memPtr = z.bc() - 1
	z.core.IOWrite(z.bc(), val)
	z.setHL(z.hl() - 1)
	other := val + z.l
	z.nf = val&0x80 != 0
	if other < val {
		z.hf = true
		z.cf = true
	} else {
		z.hf = false
		z.cf = false
	}
	z.pf = parity((other & 0x07) ^ z.b)
	z.zf = z.b == 0
	z.sf = z.b&0x80 != 0
	z.updateXY(z.b)
}

func (z *Z80) daa() {
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

func (z *Z80) displace(baseAddr uint16, offset byte) uint16 {
	addr := baseAddr
	if offset&0x80 == 0x80 {
		addr += 0xff00 | uint16(offset)
	} else {
		addr += uint16(offset)
	}
	//addr := baseAddr + uint16(displacement)
	z.memPtr = addr
	return addr
}

func (z *Z80) processInterrupts() {
	// "When an EI instruction is executed, any pending interrupt request
	// is not accepted until after the instruction following EI is executed."
	if z.iffDelay > 0 {
		z.iffDelay -= 1
		if z.iffDelay == 0 {
			z.iff1 = true
			z.iff2 = true
		}
		return
	}

	if z.nmiPending {
		z.nmiPending = false
		z.isHalted = false
		z.iff1 = false
		z.incR()

		z.cycleCount += 11
		z.call(0x0066)
		return
	}

	if z.intPending && z.iff1 {
		z.intPending = false
		z.isHalted = false
		z.iff1 = false
		z.iff2 = false
		z.incR()

		switch z.interruptMode {
		case 0:
			z.cycleCount += 11
			z.execOpcode(z.intData)
		case 1:
			z.cycleCount += 13
			z.call(0x38)
		case 2:
			z.cycleCount += 19
			z.call(z.rw((uint16(z.i) << 8) | uint16(z.intData)))
		default:
			log.Errorf("Unsupported interrupt mode %d\n", z.interruptMode)
		}
		return
	}
}

// GenNMI function to call when an NMI is to be serviced
func (z *Z80) GenNMI() {
	z.nmiPending = true
}

// GenINT function to call when an INT is to be serviced
func (z *Z80) GenINT(data byte) {
	z.intPending = true
	z.intData = data
}

// executes a non-prefixed opcode
func (z *Z80) execOpcode(opcode byte) {
	z.cycleCount += uint32(cycles00[opcode])
	z.incR()

	switch opcode {
	case 0x7F:
		//z.a = z.a // ld a,a
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
		//z.b = z.b // ld b,b
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
		//z.c = z.c // ld c,c
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
		//z.d = z.d // ld d,d
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
		//z.e = z.e // ld e,e
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
		//z.h = z.h // ld h,h
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
		//z.l = z.l // ld l,l

	case 0x7E:
		z.a = z.rb(z.hl()) // ld a,(hl)
	case 0x46:
		z.b = z.rb(z.hl()) // ld b,(hl)
	case 0x4E:
		z.c = z.rb(z.hl()) // ld c,(hl)
	case 0x56:
		z.d = z.rb(z.hl()) // ld d,(hl)
	case 0x5E:
		z.e = z.rb(z.hl()) // ld e,(hl)
	case 0x66:
		z.h = z.rb(z.hl()) // ld h,(hl)
	case 0x6E:
		z.l = z.rb(z.hl()) // ld l,(hl)

	case 0x77:
		z.wb(z.hl(), z.a) // ld (hl),a
	case 0x70:
		z.wb(z.hl(), z.b) // ld (hl),b
	case 0x71:
		z.wb(z.hl(), z.c) // ld (hl),c
	case 0x72:
		z.wb(z.hl(), z.d) // ld (hl),d
	case 0x73:
		z.wb(z.hl(), z.e) // ld (hl),e
	case 0x74:
		z.wb(z.hl(), z.h) // ld (hl),h
	case 0x75:
		z.wb(z.hl(), z.l) // ld (hl),l

	case 0x3E:
		z.a = z.nextB() // ld a,*
	case 0x06:
		z.b = z.nextB() // ld b,*
	case 0x0E:
		z.c = z.nextB() // ld c,*
	case 0x16:
		z.d = z.nextB() // ld d,*
	case 0x1E:
		z.e = z.nextB() // ld e,*
	case 0x26:
		z.h = z.nextB() // ld h,*
	case 0x2E:
		z.l = z.nextB() // ld l,*
	case 0x36:
		z.wb(z.hl(), z.nextB()) // ld (hl),*
	case 0x0A:
		// ld a,(bc)
		z.a = z.rb(z.bc())
		z.memPtr = z.bc() + 1
	case 0x1A:
		// ld a,(de)
		z.a = z.rb(z.de())
		z.memPtr = z.de() + 1
	case 0x3A:
		// ld a,(**)
		addr := z.nextW()
		z.a = z.rb(addr)
		z.memPtr = addr + 1
	case 0x02:
		// ld (bc),a
		z.wb(z.bc(), z.a)
		z.memPtr = (uint16(z.a) << 8) | ((z.bc() + 1) & 0xFF)
	case 0x12:
		// ld (de),a
		z.wb(z.de(), z.a)
		z.memPtr = (uint16(z.a) << 8) | ((z.de() + 1) & 0xFF)
	case 0x32:
		// ld (**),a
		addr := z.nextW()
		z.wb(addr, z.a)
		z.memPtr = (uint16(z.a) << 8) | ((addr + 1) & 0xFF)
	case 0x01:
		z.setBC(z.nextW()) // ld bc,**
	case 0x11:
		z.setDE(z.nextW()) // ld de,**
	case 0x21:
		z.setHL(z.nextW()) // ld hl,**
	case 0x31:
		z.sp = z.nextW() // ld sp,**

	case 0x2A:
		// ld hl,(**)
		addr := z.nextW()
		z.setHL(z.rw(addr))
		z.memPtr = addr + 1
	case 0x22:
		// ld (**),hl
		addr := z.nextW()
		z.ww(addr, z.hl())
		z.memPtr = addr + 1
	case 0xF9:
		z.sp = z.hl() // ld sp,hl

	case 0xEB:
		// ex de,hl
		de := z.de()
		z.setDE(z.hl())
		z.setHL(de)
	case 0xE3:
		// ex (sp),hl
		val := z.rw(z.sp)
		z.ww(z.sp, z.hl())
		z.setHL(val)
		z.memPtr = val
	case 0x87:
		z.a = z.addB(z.a, z.a, false) // add a,a
	case 0x80:
		z.a = z.addB(z.a, z.b, false) // add a,b
	case 0x81:
		z.a = z.addB(z.a, z.c, false) // add a,c
	case 0x82:
		z.a = z.addB(z.a, z.d, false) // add a,d
	case 0x83:
		z.a = z.addB(z.a, z.e, false) // add a,e
	case 0x84:
		z.a = z.addB(z.a, z.h, false) // add a,h
	case 0x85:
		z.a = z.addB(z.a, z.l, false) // add a,l
	case 0x86:
		z.a = z.addB(z.a, z.rb(z.hl()), false) // add a,(hl)
	case 0xC6:
		z.a = z.addB(z.a, z.nextB(), false) // add a,*

	case 0x8F:
		z.a = z.addB(z.a, z.a, z.cf) // adc a,a
	case 0x88:
		z.a = z.addB(z.a, z.b, z.cf) // adc a,b
	case 0x89:
		z.a = z.addB(z.a, z.c, z.cf) // adc a,c
	case 0x8A:
		z.a = z.addB(z.a, z.d, z.cf) // adc a,d
	case 0x8B:
		z.a = z.addB(z.a, z.e, z.cf) // adc a,e
	case 0x8C:
		z.a = z.addB(z.a, z.h, z.cf) // adc a,h
	case 0x8D:
		z.a = z.addB(z.a, z.l, z.cf) // adc a,l
	case 0x8E:
		z.a = z.addB(z.a, z.rb(z.hl()), z.cf) // adc a,(hl)
	case 0xCE:
		z.a = z.addB(z.a, z.nextB(), z.cf) // adc a,*

	case 0x97:
		z.a = z.subB(z.a, z.a, false) // sub a,a
	case 0x90:
		z.a = z.subB(z.a, z.b, false) // sub a,b
	case 0x91:
		z.a = z.subB(z.a, z.c, false) // sub a,c
	case 0x92:
		z.a = z.subB(z.a, z.d, false) // sub a,d
	case 0x93:
		z.a = z.subB(z.a, z.e, false) // sub a,e
	case 0x94:
		z.a = z.subB(z.a, z.h, false) // sub a,h
	case 0x95:
		z.a = z.subB(z.a, z.l, false) // sub a,l
	case 0x96:
		z.a = z.subB(z.a, z.rb(z.hl()), false) // sub a,(hl)
	case 0xD6:
		z.a = z.subB(z.a, z.nextB(), false) // sub a,*

	case 0x9F:
		z.a = z.subB(z.a, z.a, z.cf) // sbc a,a
	case 0x98:
		z.a = z.subB(z.a, z.b, z.cf) // sbc a,b
	case 0x99:
		z.a = z.subB(z.a, z.c, z.cf) // sbc a,c
	case 0x9A:
		z.a = z.subB(z.a, z.d, z.cf) // sbc a,d
	case 0x9B:
		z.a = z.subB(z.a, z.e, z.cf) // sbc a,e
	case 0x9C:
		z.a = z.subB(z.a, z.h, z.cf) // sbc a,h
	case 0x9D:
		z.a = z.subB(z.a, z.l, z.cf) // sbc a,l
	case 0x9E:
		z.a = z.subB(z.a, z.rb(z.hl()), z.cf) // sbc a,(hl)
	case 0xDE:
		z.a = z.subB(z.a, z.nextB(), z.cf) // sbc a,*

	case 0x09:
		z.addHL(z.bc()) // add hl,bc
	case 0x19:
		z.addHL(z.de()) // add hl,de
	case 0x29:
		z.addHL(z.hl()) // add hl,hl
	case 0x39:
		z.addHL(z.sp) // add hl,sp

	case 0xF3:
		z.iff1 = false
		z.iff2 = false // di
	case 0xFB:
		z.iffDelay = 1 // ei
	case 0x00: // nop
	case 0x76:
		z.isHalted = true // halt
		z.pc--
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
		result := z.inc(z.rb(z.hl()))
		z.wb(z.hl(), result)
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
		result := z.dec(z.rb(z.hl()))
		z.wb(z.hl(), result)
	case 0x03:
		z.setBC(z.bc() + 1) // inc bc
	case 0x13:
		z.setDE(z.de() + 1) // inc de
	case 0x23:
		z.setHL(z.hl() + 1) // inc hl
	case 0x33:
		z.sp = z.sp + 1 // inc sp
	case 0x0B:
		z.setBC(z.bc() - 1) // dec bc
	case 0x1B:
		z.setDE(z.de() - 1) // dec de
	case 0x2B:
		z.setHL(z.hl() - 1) // dec hl
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
		z.updateXY(z.a | z.f())
	case 0x3F:
		// ccf
		z.hf = z.cf
		z.cf = !z.cf
		z.nf = false
		z.updateXY(z.a | z.f())
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
		z.lAnd(z.a) // and a
	case 0xA0:
		z.lAnd(z.b) // and b
	case 0xA1:
		z.lAnd(z.c) // and c
	case 0xA2:
		z.lAnd(z.d) // and d
	case 0xA3:
		z.lAnd(z.e) // and e
	case 0xA4:
		z.lAnd(z.h) // and h
	case 0xA5:
		z.lAnd(z.l) // and l
	case 0xA6:
		z.lAnd(z.rb(z.hl())) // and (hl)
	case 0xE6:
		z.lAnd(z.nextB()) // and *

	case 0xAF:
		z.lXor(z.a) // xor a
	case 0xA8:
		z.lXor(z.b) // xor b
	case 0xA9:
		z.lXor(z.c) // xor c
	case 0xAA:
		z.lXor(z.d) // xor d
	case 0xAB:
		z.lXor(z.e) // xor e
	case 0xAC:
		z.lXor(z.h) // xor h
	case 0xAD:
		z.lXor(z.l) // xor l
	case 0xAE:
		z.lXor(z.rb(z.hl())) // xor (hl)
	case 0xEE:
		z.lXor(z.nextB()) // xor *

	case 0xB7:
		z.lOr(z.a) // or a
	case 0xB0:
		z.lOr(z.b) // or b
	case 0xB1:
		z.lOr(z.c) // or c
	case 0xB2:
		z.lOr(z.d) // or d
	case 0xB3:
		z.lOr(z.e) // or e
	case 0xB4:
		z.lOr(z.h) // or h
	case 0xB5:
		z.lOr(z.l) // or l
	case 0xB6:
		z.lOr(z.rb(z.hl())) // or (hl)
	case 0xF6:
		z.lOr(z.nextB()) // or *

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
		z.cp(z.rb(z.hl())) // cp (hl)
	case 0xFE:
		z.cp(z.nextB()) // cp *

	case 0xC3:
		z.jump(z.nextW()) // jm **
	case 0xC2:
		z.condJump(!z.zf) // jp nz, **
	case 0xCA:
		z.condJump(z.zf) // jp z, **
	case 0xD2:
		z.condJump(!z.cf) // jp nc, **
	case 0xDA:
		z.condJump(z.cf) // jp c, **
	case 0xE2:
		z.condJump(!z.pf) // jp po, **
	case 0xEA:
		z.condJump(z.pf) // jp pe, **
	case 0xF2:
		z.condJump(!z.sf) // jp p, **
	case 0xFA:
		z.condJump(z.sf) // jp m, **

	case 0x10:
		z.b--
		z.condJr(z.b != 0) // djnz *
	case 0x18:
		z.pc += uint16(z.nextB()) // jr *
		z.memPtr = z.pc
	case 0x20:
		z.condJr(!z.zf) // jr nz, *
	case 0x28:
		z.condJr(z.zf) // jr z, *
	case 0x30:
		z.condJr(!z.cf) // jr nc, *
	case 0x38:
		z.condJr(z.cf) // jr c, *

	case 0xE9:
		z.pc = z.hl() // jp (hl)
	case 0xCD:
		z.call(z.nextW()) // call

	case 0xC4:
		z.condCall(!z.zf) // cnz
	case 0xCC:
		z.condCall(z.zf) // cz
	case 0xD4:
		z.condCall(!z.cf) // cnc
	case 0xDC:
		z.condCall(z.cf) // cc
	case 0xE4:
		z.condCall(!z.pf) // cpo
	case 0xEC:
		z.condCall(z.pf) // cpe
	case 0xF4:
		z.condCall(!z.sf) // cp
	case 0xFC:
		z.condCall(z.sf) // cm

	case 0xC9:
		z.ret() // ret
	case 0xC0:
		z.condRet(!z.zf) // ret nz
	case 0xC8:
		z.condRet(z.zf) // ret z
	case 0xD0:
		z.condRet(!z.cf) // ret nc
	case 0xD8:
		z.condRet(z.cf) // ret c
	case 0xE0:
		z.condRet(!z.pf) // ret po
	case 0xE8:
		z.condRet(z.pf) // ret pe
	case 0xF0:
		z.condRet(!z.sf) // ret p
	case 0xF8:
		z.condRet(z.sf) // ret m

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
		z.pushW(z.bc()) // push bc
	case 0xD5:
		z.pushW(z.de()) // push de
	case 0xE5:
		z.pushW(z.hl()) // push hl
	case 0xF5:
		z.pushW((uint16(z.a) << 8) | uint16(z.f())) // push af

	case 0xC1:
		z.setBC(z.popW()) // pop bc
	case 0xD1:
		z.setDE(z.popW()) // pop de
	case 0xE1:
		z.setHL(z.popW()) // pop hl
	case 0xF1:
		// pop af
		val := z.popW()
		z.a = byte(val >> 8)
		z.setF(byte(val))
	case 0xDB:
		// in a,(n)
		port := (uint16(z.a) << 8) | uint16(z.nextB())
		z.a = z.core.IORead(port)
		z.memPtr = port + 1 // (uint16(a) << 8) | (uint16(z.a+1) & 0x00ff)
	case 0xD3:
		// out (n), a
		port := uint16(z.nextB())
		z.core.IOWrite(port, z.a)
		z.memPtr = ((port + 1) & 0x00ff) | (uint16(z.a) << 8)
	case 0x08:
		// ex af,af'
		a := z.a
		f := z.f()

		z.a = z.a_
		z.setF(z.f_)

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
		z.execOpcodeCB(z.nextB())
	case 0xED:
		z.execOpcodeED(z.nextB())
	case 0xDD:
		z.execOpcodeDDFD(z.nextB(), &z.ix)
	case 0xFD:
		z.execOpcodeDDFD(z.nextB(), &z.iy)

	default:
		log.Errorf("Unknown opcode %02X\n", opcode)
	}
}
