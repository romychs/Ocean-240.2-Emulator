package z80

import log "github.com/sirupsen/logrus"

// executes a CB opcode
func (z *Z80) exec_opcode_cb(opcode byte) {
	z.cyc += 8
	z.inc_r()

	// decoding instructions from http://z80.info/decoding.htm#cb
	x_ := (opcode >> 6) & 3 // 0b11
	y_ := (opcode >> 3) & 7 // 0b111
	z_ := opcode & 7        // 0b111

	var hl byte
	v := byte(0)
	reg := &v
	switch z_ {
	case 0:
		reg = &z.b
	case 1:
		reg = &z.c
	case 2:
		reg = &z.d
	case 3:
		reg = &z.e
	case 4:
		reg = &z.h
	case 5:
		reg = &z.l
	case 6:
		hl = z.rb(z.get_hl())
		reg = &hl
	case 7:
		reg = &z.a
	}

	switch x_ {
	case 0:
		// rot[y] r[z]
		switch y_ {
		case 0:
			*reg = z.cb_rlc(*reg)
		case 1:
			*reg = z.cb_rrc(*reg)
		case 2:
			*reg = z.cb_rl(*reg)
		case 3:
			*reg = z.cb_rr(*reg)
		case 4:
			*reg = z.cb_sla(*reg)
		case 5:
			*reg = z.cb_sra(*reg)
		case 6:
			*reg = z.cb_sll(*reg)
		case 7:
			*reg = z.cb_srl(*reg)
		}

	case 1:
		// BIT y, r[z]
		z.cb_bit(*reg, y_)

		// in bit (hl), x/y flags are handled differently:
		if z_ == 6 {
			z.updateXY(byte(z.mem_ptr >> 8))
			z.cyc += 4
		}

	case 2:
		*reg &= ^(1 << y_) // RES y, r[z]
	case 3:
		*reg |= 1 << y_ // SET y, r[z]
	}

	if (x_ == 0 || x_ == 2 || x_ == 3) && z_ == 6 {
		z.cyc += 7
	}

	if reg == &hl {
		z.wb(z.get_hl(), hl)
	}
}

// exec_opcode_dcb executes a displaced CB opcode (DDCB or FDCB)
func (z *Z80) exec_opcode_dcb(opcode byte, addr uint16) {
	val := z.rb(addr)
	result := byte(0)

	// decoding instructions from http://z80.info/decoding.htm#ddcb
	x_ := (opcode >> 6) & 3 // 0b11
	y_ := (opcode >> 3) & 7 // 0b111
	z_ := opcode & 7        // 0b111

	switch x_ {
	case 0:
		// rot[y] (iz+d)
		switch y_ {
		case 0:
			result = z.cb_rlc(val)
		case 1:
			result = z.cb_rrc(val)
		case 2:
			result = z.cb_rl(val)
		case 3:
			result = z.cb_rr(val)
		case 4:
			result = z.cb_sla(val)
		case 5:
			result = z.cb_sra(val)
		case 6:
			result = z.cb_sll(val)
		case 7:
			result = z.cb_srl(val)
		}

	case 1:
		// bit y,(iz+d)
		result = z.cb_bit(val, y_)
		z.updateXY(byte(addr >> 8))
	case 2:
		result = val & ^(1 << y_) // res y, (iz+d)
	case 3:
		result = val | (1 << y_) // set y, (iz+d)

	default:
		log.Errorf("Unknown XYCB opcode: %02X\n", opcode)
	}

	// ld r[z], rot[y] (iz+d)
	// ld r[z], res y,(iz+d)
	// ld r[z], set y,(iz+d)
	if x_ != 1 && z_ != 6 {
		switch z_ {
		case 0:
			z.b = result
		case 1:
			z.c = result
		case 2:
			z.d = result
		case 3:
			z.e = result
		case 4:
			z.h = result
		case 5:
			z.l = result
		// always false
		//case 6:
		//	z.wb(z.get_hl(), result)
		case 7:
			z.a = result
		}
	}

	if x_ == 1 {
		// bit instructions take 20 cycles, others take 23
		z.cyc += 20
	} else {
		z.wb(addr, result)
		z.cyc += 23
	}
}
