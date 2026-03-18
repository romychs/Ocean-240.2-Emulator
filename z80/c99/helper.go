package c99

import log "github.com/sirupsen/logrus"

func (z *Z80) rb(addr uint16) byte {
	z.memAccess[addr] = MemAccessRead
	return z.core.MemRead(addr)
}

func (z *Z80) wb(addr uint16, val byte) {
	z.memAccess[addr] = MemAccessWrite
	z.core.MemWrite(addr, val)
}

func (z *Z80) rw(addr uint16) uint16 {
	z.memAccess[addr] = MemAccessRead
	z.memAccess[addr+1] = MemAccessRead
	return (uint16(z.core.MemRead(addr+1)) << 8) | uint16(z.core.MemRead(addr))
}

func (z *Z80) ww(addr uint16, val uint16) {
	z.memAccess[addr] = MemAccessWrite
	z.memAccess[addr+1] = MemAccessWrite
	z.core.MemWrite(addr, byte(val))
	z.core.MemWrite(addr+1, byte(val>>8))
}

func (z *Z80) pushW(val uint16) {
	z.sp -= 2
	z.ww(z.sp, val)
}

func (z *Z80) popW() uint16 {
	z.sp += 2
	return z.rw(z.sp - 2)
}

func (z *Z80) nextB() byte {
	b := z.core.MemRead(z.pc)
	z.pc++
	return b
}

func (z *Z80) nextW() uint16 {
	w := (uint16(z.core.MemRead(z.pc+1)) << 8) | uint16(z.core.MemRead(z.pc))
	z.pc += 2
	return w
}

func (z *Z80) bc() uint16 {
	return (uint16(z.b) << 8) | uint16(z.c)
}

func (z *Z80) de() uint16 {
	return (uint16(z.d) << 8) | uint16(z.e)
}

func (z *Z80) hl() uint16 {
	return (uint16(z.h) << 8) | uint16(z.l)
}

func (z *Z80) setBC(val uint16) {
	z.b = byte(val >> 8)
	z.c = byte(val)
}

func (z *Z80) setDE(val uint16) {
	z.d = byte(val >> 8)
	z.e = byte(val)
}

func (z *Z80) setHL(val uint16) {
	z.h = byte(val >> 8)
	z.l = byte(val)
}

func (z *Z80) f() byte {
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

func (z *Z80) setF(val byte) {
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
func (z *Z80) incR() {
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
func carry(bitNo int, a uint16, b uint16, cy bool) bool {
	result := int32(a) + int32(b) + boolToInt32(cy)
	carry := result ^ int32(a) ^ int32(b)
	return (carry & (1 << bitNo)) != 0
}

// parity returns the parity of byte: 0 if odd, else 1
func parity(val byte) bool {
	ones := byte(0)
	for i := 0; i < 8; i++ {
		ones += (val >> i) & 1
	}
	return (ones & 1) == 0
}

// updateXY set undocumented 3rd (X) and 5th (Y) flags
func (z *Z80) updateXY(result byte) {
	z.yf = result&0x20 != 0
	z.xf = result&0x08 != 0
}

func (z *Z80) DebugOutput() {
	log.Debugf("PC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X, IX: %04X, IY: %04X, I: %02X, R: %02X",
		z.pc, (uint16(z.a)<<8)|uint16(z.f()), z.bc(), z.de(), z.hl(), z.sp,
		z.ix, z.iy, z.i, z.r)

	log.Debugf("\t(%02X %02X %02X %02X), cycleCount: %d\n", z.rb(z.pc), z.rb(z.pc+1),
		z.rb(z.pc+2), z.rb(z.pc+3), z.cycleCount)
}

func (z *Z80) Reset() {
	z.cycleCount = 0
	z.pc = 0
	z.sp = 0xFFFF
	z.ix = 0
	z.iy = 0
	z.memPtr = 0

	z.a = 0xFF
	z.b = 0
	z.c = 0
	z.d = 0
	z.e = 0
	z.h = 0
	z.l = 0

	z.a_ = 0
	z.b_ = 0
	z.c_ = 0
	z.d_ = 0
	z.e_ = 0
	z.h_ = 0
	z.l_ = 0
	z.f_ = 0

	z.i = 0
	z.r = 0

	z.sf = true
	z.zf = true
	z.yf = true
	z.hf = true
	z.xf = true
	z.pf = true
	z.nf = true
	z.cf = true

	z.iffDelay = 0
	z.interruptMode = 0
	z.iff1 = false
	z.iff2 = false
	z.isHalted = false
	z.intPending = false
	z.nmiPending = false
	z.intData = 0
}
