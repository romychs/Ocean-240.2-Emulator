package okean240

/*
 * Ocean-240.2
 * Computer with FDC variant.
 * IO Ports definitions
 *
 * By Romych 2026-03-01
 */

import log "github.com/sirupsen/logrus"

func (c *ComputerType) IORead(port uint16) byte {
	switch port & 0x00ff {
	case PicDd75a:
		// PIO xx59, get IRR register
		irr := c.pic.IRR()
		// if irq from keyboard and no ACK applied, re-fire
		if irr&RstKbdMask != 0 && !c.kbAck.Load() {
			log.Tracef("KBD IRQ REFIRE PC=%04X", c.cpu.PC())
			c.pic.SetIRQ(RstKbdNo)
		}
		return irr
	case PicDd75b:
		return c.pic.CSW()
	case UartDd72rr:
		// USART VV51 CMD
		return c.usart.Status()
	case UartDd72rd:
		// USART VV51 Data
		return c.usart.Receive()
	case KbdDd78pa:
		// Keyboard data
		log.Tracef("KBD RD: %d, PC=%04X", c.ioPorts[KbdDd78pa], c.cpu.PC())
		return c.ioPorts[KbdDd78pa]
	case KbdDd78pb:
		return c.ioPorts[KbdDd78pb]
	case FdcCmd:
		return c.fdc.Status()
	case FdcDrq:
		return c.fdc.Drq()
	case Floppy:
		return c.fdc.GetFloppy()
	case FdcData:
		return c.fdc.Data()
	case FdcTrack:
		return c.fdc.Track()
	case FdcSect:
		return c.fdc.Sector()

	default:
		log.Debugf("IORead from port: %x", port)
	}
	return c.ioPorts[byte(port&0x00ff)]
}

func (c *ComputerType) IOWrite(port uint16, val byte) {
	bp := byte(port & 0x00ff)
	c.ioPorts[bp] = val
	//log.Debugf("OUT (%x), %x", bp, val)
	switch bp {
	case SysDd17pb:
		if c.dd17EnableOut {
			c.memory.Configure(val)
		}
	case SysDd17ctr:
		c.dd17EnableOut = val == 0x80
	case VidDd67pb:
		if val&VidVsuBit == 0 {
			// video page 0
			c.vRAM = c.memory.allMemory[VRAMBlock0]
		} else {
			// video page 1
			c.vRAM = c.memory.allMemory[VRAMBlock1]
		}
		if val&VidColorBit != 0 {
			c.colorMode = true
			c.screenWidth = 256
		} else {
			c.colorMode = false
			c.screenWidth = 512
		}
		c.palette = val & 0x07
		c.bgColor = (val >> 3) & 0x07
	case SysDd17pa:
		c.vShift = val
	case SysDd17pc:
		c.hShift = val
	case TmrDd70ctr:
		// Timer VI63 config register
		c.pit.Configure(val)
	case TmrDd70c1:
		// Timer VI63 counter0 register
		c.pit.Load(0, val)
	case TmrDd70c2:
		// Timer VI63 counter1 register
		c.pit.Load(1, val)
	case TmrDd70c3:
		// Timer VI63 counter2 register
		c.pit.Load(2, val)

	case UartDd72rr:
		// USART VV51 CMD
		c.usart.Command(val)
	case UartDd72rd:
		// USART VV51 Data
		c.usart.Send(val)
	case PicDd75b:
		c.pic.SetCSW(val)
	case FdcCmd:
		c.fdc.SetCmd(val)
	case FdcData:
		c.fdc.SetData(val)
	case FdcTrack:
		c.fdc.SetTrackNo(val)
	case FdcSect:
		c.fdc.SetSectorNo(val)
	case Floppy:
		c.fdc.SetFloppy(val)

	case KbdDd78pc:
		if val&KbAckBit != 0 {
			c.kbAck.Store(true)
			log.Trace("KBD ACK")
		} else {
			if c.kbAck.Load() {
				c.pic.ResetIRQ(RstKbdNo)
			}
		}
	default:
		//log.Debugf("OUT to Unknown port (%x), %x", bp, val)
	}
}
