package pic

import log "github.com/sirupsen/logrus"

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

// IRR Return value of interrupt request register
func (c *I8259) IRR() byte {
	irr := c.irr
	// Reset the highest IR bit
	if irr&0x01 != 0 {
		c.irr &= 0xFE
	} else if irr&0x02 != 0 {
		c.irr &= 0xFD
	} else if irr&0x04 != 0 {
		c.irr &= 0xFB
	} else if irr&0x08 != 0 {
		c.irr &= 0xF7
	} else if irr&0x10 != 0 {
		c.irr &= 0xEF
	} else if irr&0x20 != 0 {
		c.irr &= 0xDF
	} else if irr&0x40 != 0 {
		c.irr &= 0xBF
	} else if irr&0x80 != 0 {
		c.irr &= 0x7F
	}
	return irr
}

// ResetIRQ  Reset interrupt request flag for specified irq
func (c *I8259) ResetIRQ(irq byte) {
	c.irr &= ^(byte(1) << (irq & 0x07))
	log.Tracef("RESET IRQ %d -> IRR: %08b", irq, c.irr)
}

// SetIRQ  Set interrupt request flag for specified irq
func (c *I8259) SetIRQ(irq byte) {
	c.irr |= 1 << (irq & 0x07)
	if irq == 1 {
		log.Tracef("SET IRQ %d -> IRR: %08b", irq, c.irr)
	}
}

// CSW Return value of CSW register
func (c *I8259) CSW() byte {
	return c.csw
}

// SetCSW Set value of CSW register
func (c *I8259) SetCSW(val byte) {
	c.csw = val
}

// NewI8259 Create and initialize new i8259 controller
func NewI8259() *I8259 {
	return &I8259{
		irr: 0,
		csw: 0,
	}
}
