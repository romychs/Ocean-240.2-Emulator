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
	case PIC_DD75RS:
		// PIO VN59
		v := c.ioPorts[PIC_DD75RS]
		c.ioPorts[PIC_DD75RS] = 0
		return v
	case UART_DD72RR:
		// SIO VV51 CMD
		return c.dd72.Status()
	case UART_DD72RD:
		// SIO VV51 Data
		return c.dd72.Receive()
	case KBD_DD78PA:
		// Keyboard data
		return c.ioPorts[KBD_DD78PA]
	case KBD_DD78PB:
		return c.ioPorts[KBD_DD78PB]
	case FDC_CMD:
		return c.fdc.Status()
	case FDC_DRQ:
		return c.fdc.Drq()
	case FLOPPY:
		return c.fdc.GetFloppy()
	case FDC_DATA:
		return c.fdc.Data()

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
	case SYS_DD17PB:
		if c.dd17EnableOut {
			c.memory.Configure(val)
		}
	case SYS_DD17CTR:
		c.dd17EnableOut = val == 0x80
	case VID_DD67PB:
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
		c.bgColor = val & 0x38 >> 3
	case SYS_DD17PA:
		c.vShift = val
	case SYS_DD17PC:
		c.hShift = val
	case TMR_DD70CTR:
		// Timer VI63 config register
		c.dd70.Configure(val)
	case TMR_DD70C1:
		// Timer VI63 counter0 register
		c.dd70.Load(0, val)
	case TMR_DD70C2:
		// Timer VI63 counter1 register
		c.dd70.Load(1, val)
	case TMR_DD70C3:
		// Timer VI63 counter2 register
		c.dd70.Load(2, val)

	case UART_DD72RR:
		// SIO VV51 CMD
		c.dd72.Command(val)
	case UART_DD72RD:
		// SIO VV51 Data
		c.dd72.Send(val)
	case FDC_CMD:
		c.fdc.SetCmd(val)
	case FDC_DATA:
		c.fdc.SetData(val)
	case FDC_TRACK:
		c.fdc.SetTrack(val)
	case FDC_SECT:
		c.fdc.SetSector(val)
	case FLOPPY:
		c.fdc.SetFloppy(val)
	default:
		//log.Debugf("OUT to Unknown port (%x), %x", bp, val)

	}
}
