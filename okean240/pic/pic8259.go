package pic

/*
    Programmable Interrupt Controller
	i8058, MSM82C59, КР580ВН59

	By Romych, 2025.03.05
*/

type I8259 struct {
	irr byte
	csw byte
}

type I8259Interface interface {
	SetIRQ(irq byte)
	IRR() byte
	CSW() byte
	SetCSW(val byte)
}

func (c *I8259) IRR() byte {
	irr := c.irr
	// Reset the highest IR bit
	if irr&0x80 != 0 {
		c.irr &= 0x7F
	} else if irr&0x40 != 0 {
		c.irr &= 0x3F
	} else if irr&0x20 != 0 {
		c.irr &= 0x1F
	} else if irr&0x08 != 0 {
		c.irr &= 0x07
	} else if irr&0x04 != 0 {
		c.irr &= 0x03
	} else if irr&0x02 != 0 {
		c.irr &= 0x1
	} else {
		c.irr = 0
	}
	return irr
}

func (c *I8259) ResetIRQ(irq byte) {
	c.irr &= ^(byte(1) << (irq & 0x07))
}
func (c *I8259) SetIRQ(irq byte) {
	c.irr |= 1 << (irq & 0x07)
}

func (c *I8259) CSW() byte {
	return c.csw
}

func (c *I8259) SetCSW(val byte) {
	c.csw = val
}

func New() *I8259 {
	return &I8259{
		irr: 0,
	}
}
