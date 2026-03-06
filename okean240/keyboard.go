package okean240

import (
	"fyne.io/fyne/v2"
)

func (c *ComputerType) PutKey(key *fyne.KeyEvent) {

	code := RemapCmdKey[key.Name]
	if code > 0 {
		//log.Debugf("PutKey keyName: %s", key.Name)
		c.ioPorts[KBD_DD78PA] = code
		c.dd75.SetIRQ(RstKbdNo)
	}

}

func (c *ComputerType) PutRune(key rune) {

	//log.Debugf("Put Rune: %c  Lo: %x, Hi: %x", key, key&0xff, key>>8)

	c.ioPorts[KBD_DD78PA] = byte(key & 0xff)
	c.dd75.SetIRQ(RstKbdNo)
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
	c.dd75.SetIRQ(RstKbdNo)
	//c.ioPorts[PIC_DD75RS] |= Rst1Mask
	c.ioPorts[KBD_DD78PB] &= 0x1f | 0x20
}
