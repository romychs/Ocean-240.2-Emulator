package c99

import log "github.com/sirupsen/logrus"

// GetBit return bit "n" from byte "val"
func getBit(n byte, val byte) byte {
	return ((val) >> (n)) & 1
}

func getBit3(val byte) bool {
	return (val & 0x08) != 0
}

func getBit5(val byte) bool {
	return (val & 0x20) != 0
}

func (z *Z80) rb(addr uint16) byte {
	return z.core.MemRead(addr)
}

func (z *Z80) wb(addr uint16, val byte) {
	z.core.MemWrite(addr, val)
}

func (z *Z80) rw(addr uint16) uint16 {
	return (uint16(z.core.MemRead(addr+1)) << 8) | uint16(z.core.MemRead(addr))
}

func (z *Z80) ww(addr uint16, val uint16) {
	z.core.MemWrite(addr, byte(val))
	z.core.MemWrite(addr+1, byte(val>>8))
}

func (z *Z80) pushw(val uint16) {
	z.sp -= 2
	z.ww(z.sp, val)
}

func (z *Z80) popw() uint16 {
	z.sp += 2
	return z.rw(z.sp - 2)
}

func (z *Z80) nextb() byte {
	b := z.rb(z.pc)
	z.pc++
	return b
}

func (z *Z80) nextw() uint16 {
	w := z.rw(z.pc)
	z.pc += 2
	return w
}

func (z *Z80) get_bc() uint16 {
	return (uint16(z.b) << 8) | uint16(z.c)
}

func (z *Z80) get_de() uint16 {
	return (uint16(z.d) << 8) | uint16(z.e)
}

func (z *Z80) get_hl() uint16 {
	return (uint16(z.h) << 8) | uint16(z.l)
}

func (z *Z80) set_bc(val uint16) {
	z.b = byte(val >> 8)
	z.c = byte(val)
}

func (z *Z80) set_de(val uint16) {
	z.d = byte(val >> 8)
	z.e = byte(val)
}

func (z *Z80) set_hl(val uint16) {
	z.h = byte(val >> 8)
	z.l = byte(val)
}

func (z *Z80) get_f() byte {
	val := byte(0)
	if z.cf {
		val |= 0x01
	}
	if z.nf {
		val |= 0x02
	}
	if z.pf {
		val |= 0x04
	}
	if z.xf {
		val |= 0x08
	}
	if z.hf {
		val |= 0x10
	}
	if z.yf {
		val |= 0x20
	}
	if z.zf {
		val |= 0x40
	}
	if z.sf {
		val |= 0x80
	}
	return val
}

func (z *Z80) set_f(val byte) {
	z.cf = val&1 != 0
	z.nf = (val>>1)&1 != 0
	z.pf = (val>>2)&1 != 0
	z.xf = (val>>3)&1 != 0
	z.hf = (val>>4)&1 != 0
	z.yf = (val>>5)&1 != 0
	z.zf = (val>>6)&1 != 0
	z.sf = (val>>7)&1 != 0
}

// increments R, keeping the highest byte intact
func (z *Z80) inc_r() {
	z.r = (z.r & 0x80) | ((z.r + 1) & 0x7f)
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// returns if there was a carry between bit "bit_no" and "bit_no - 1" when
// executing "a + b + cy"
func carry(bit_no int, a uint16, b uint16, cy bool) bool {
	result := int32(a) + int32(b) + boolToInt32(cy)
	carry := result ^ int32(a) ^ int32(b)
	return (carry & (1 << bit_no)) != 0
}

// returns the parity of byte: 0 if number of 1 bits in `val` is odd, else 1
func parity(val byte) bool {
	ones := byte(0)
	for i := 0; i < 8; i++ {
		ones += (val >> i) & 1
	}
	return (ones & 1) == 0
}

func (z *Z80) updateXY(result byte) {
	z.yf = result&0x20 != 0
	z.xf = result&0x08 != 0
}

func (z *Z80) debugOutput() {
	log.Debugf("PC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X, IX: %04X, IY: %04X, I: %02X, R: %02X",
		z.pc, (uint16(z.a)<<8)|uint16(z.get_f()), z.get_bc(), z.get_de(), z.get_hl(), z.sp,
		z.ix, z.iy, z.i, z.r)

	log.Debugf("\t(%02X %02X %02X %02X), cyc: %d\n", z.rb(z.pc), z.rb(z.pc+1),
		z.rb(z.pc+2), z.rb(z.pc+3), z.cyc)
}
