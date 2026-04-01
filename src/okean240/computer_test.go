package okean240

func calculateVAddress(x uint16, y uint16) (uint16, uint16) {

	var offset uint16
	if (c.vShift != 0) && (y > 255-uint16(c.vShift)) {
		offset = 0x100
	} else {
		offset = 0
	}
	y += uint16(c.vShift) & 0x00ff
	x += uint16(c.hShift-7) & 0x00ff

	// Color 256x256 mode
	addr = ((x & 0xf8) << 6) | y

	a1 := (addr - offset) & 0x3fff
	a2 := (addr + 0x100 - offset) & 0x3fff

}
