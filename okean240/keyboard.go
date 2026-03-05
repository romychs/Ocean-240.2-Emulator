package okean240

import (
	"fyne.io/fyne/v2"
	log "github.com/sirupsen/logrus"
)

func (c *ComputerType) PutKey(key *fyne.KeyEvent) {

	if key.Name == fyne.KeyUnknown {
		log.Debugf("Unknown key scancode: %X", key.Physical.ScanCode)
		return
	}

	log.Debugf("PutKey keyName: %s", key.Name)

	if len(c.kbdBuffer) < KbdBufferSize {

		var code byte

		if (c.ioPorts[KBD_DD78PB] & 0x40) == 0 {
			// No shift
			code = RemapKey[key.Name]
		} else {
			// Shift
			code = RemapKeyShift[key.Name]
		}
		c.ioPorts[KBD_DD78PB] &= 0x1f

		if code != 0 {
			c.ioPorts[KBD_DD78PA] = code
			c.ioPorts[PIC_DD75RS] |= Rst1KbdFlag
		} else {
			switch key.Name {
			case "LeftAlt", "RightAlt":
				c.ioPorts[KBD_DD78PB] |= 0x80
			case "LeftControl", "RightControl":
				c.ioPorts[KBD_DD78PB] |= 0x20
			case "LeftShift", "RightShift":
				log.Debug("Shift")
				c.ioPorts[KBD_DD78PB] |= 0x40
			default:
				log.Debugf("Unhandled KeyName: %s  code: %X", key.Name, key.Physical.ScanCode)
			}
		}
	}

}

/*
	CTRL_C				EQU	0x03                        ; Warm boot
	CTRL_H		        EQU	0x08                        ; Backspace
	CTRL_E				EQU	0x05                        ; Move to beginning of new line (Physical EOL)
	CTRL_J              EQU 0x0A                        ; LF - Line Feed
	CTRL_M              EQU 0x0D                        ; CR - Carriage Return
	CTRL_P				EQU	0x10                        ; turn on/off printer
	CTRL_R              EQU 0x12                        ; Repeat current cmd line
	CTRL_S				EQU	0x13                        ; Temporary stop display data to console (aka DC3)
	CTRL_U              EQU 0x15                        ; Cancel (erase) current cmd line
	CTRL_X              EQU 0x18                        ; Cancel (erase) current cmd line
*/

func (c *ComputerType) PutCtrlKey(key byte) {
	c.ioPorts[KBD_DD78PA] = key
	c.ioPorts[PIC_DD75RS] |= Rst1KbdFlag
	c.ioPorts[KBD_DD78PB] &= 0x1f | 0x20
}
