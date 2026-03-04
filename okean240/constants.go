package okean240

import "fyne.io/fyne/v2"

/*
 * КР580ВВ55 DD79  USER PORT
 */

// USR_DD79PA User port A
const USR_DD79PA = 0x00

// USR_DD79PB User port B
const USR_DD79PB = 0x01

// USR_DD79PC User port C
const USR_DD79PC = 0x02

// USR_DD79CTR Config
const USR_DD79CTR = 0x03 // Config: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]

/*
 * КР1818ВГ93 FDC Controller
 */

// FDC_CMD FDC Command
const FDC_CMD = 0x20

// FDC_TRACK FDC Track No
const FDC_TRACK = 0x21

// FDC_SECT FDC Sector
const FDC_SECT = 0x22

// FDC_DATA FDC Data
const FDC_DATA = 0x23

// FDC_WAIT FDC Wait
const FDC_WAIT = 0x24

/*
 * Floppy Controller port
 */

// FLOPPY Floppy Controller port
const FLOPPY = 0x25 // WR: 5-SSEN, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT

/*
 * КР580ВВ55 DD78  Keyboard
 */

// KBD_DD78PA Port A - Keyboard Data
const KBD_DD78PA = 0x40

// KBD_DD78PB Port B - JST3,SHFT,CTRL,ACK,TAPE5,TAPE4,GK,GC
const KBD_DD78PB = 0x41

// KBD_DD78PC Port C - [PC7:5],[KBD_ACK],[PC3:0]
const KBD_DD78PC = 0x42

// KBD_DD78CTR Control port
const KBD_DD78CTR = 0x43 //	Сonfig: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]

/*
 * КР580ВИ53 DD70
 */

// TMR_DD70C1 Timer load 1
const TMR_DD70C1 = 0x60

// TMR_DD70C2 Timer load 2
const TMR_DD70C2 = 0x61

// TMR_DD70C3 Timer load 3
const TMR_DD70C3 = 0x62

/*
	 TMR_DD70CTR
	    Timer config: [sc1,sc0][rl1,rl0][m2,m1,m0][bcd]
		sc - timer, rl=01-LSB, 10-MSB, 11-LSB+MSB
		mode 000 - intr on fin,
		     001 - one shot,
		     x10 - rate gen,
		     x11-sq wave
*/
const TMR_DD70CTR = 0x63

/*
 * Programmable Interrupt controller PIC  KR580VV59
 */

// PIC_DD75RS RS Port
const PIC_DD75RS = 0x80

const Rst0SysFlag = 0x01 // System interrupt
const Rst1KbdFlag = 0x02 // Keyboard interrupt
const Rst2SerFlag = 0x04 // Serial interface interrupt
const RstЗLptFlag = 0x08 // Printer ready
const Rst4TmrFlag = 0x10 // System timer
const Rst5PwrFlag = 0x20 // Power intr
const Rst6UsrFlag = 0x40 // User device 1 interrupt
const Rst7UsrFlag = 0x80 // User device 1 interrupt

// PIC_DD75RM RM Port
const PIC_DD75RM = 0x81

/*
 * КР580ВВ51 DD72
 */

// UART_DD72RD Serial data
const UART_DD72RD = 0xA0

// UART_DD72RR Serial status [RST,RQ_RX,RST_ERR,PAUSE,RX_EN,RX_RDY,TX_RDY]
const UART_DD72RR = 0xA1

/*
 * КР580ВВ55 DD17 System port
 */

// Port A - VShift[8..1] Vertical shift
const SYS_DD17PA = 0xC0

// Port B - Memory mapper [ROM14,13][REST][ENROM-][A18,17,16][32k]
const SYS_DD17PB = 0xC1

// Port C - HShift[HS5..1,SB3..1] Horisontal shift
const SYS_DD17PC = 0xC2

/*
 * SYS_DD17CTR
 * Сonfig: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
 * Set bit: [0][xxx][bbb][0|1]
 */
const SYS_DD17CTR = 0xC3

/*
 * КР580ВВ55 DD67
 */

// LPT_DD67PA Port A - Printer Data
const LPT_DD67PA = 0xE0

// VID_DD67PB Port B - Video control [VSU,C/M,FL3:1,COL3:1]
const VID_DD67PB = 0xE1

// DD67PC Port C - [USER3:1, STB-LP, BELL, TAPE3:1]
const DD67PC = 0xE2

/*
 * DD67CTR
 * Сonfig: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
 * Set bit: [0][xxx][bbb][0|1]
 */
const DD67CTR = 0xE3

var RemapKey = map[fyne.KeyName]byte{
	fyne.KeyEscape:       0x1B,
	fyne.KeyReturn:       0x0A,
	fyne.KeyTab:          0x09,
	fyne.KeyBackspace:    0x08,
	fyne.KeyInsert:       0x00,
	fyne.KeyDelete:       0x08,
	fyne.KeyRight:        0x18,
	fyne.KeyLeft:         0x08,
	fyne.KeyDown:         0x0A,
	fyne.KeyUp:           0x19,
	fyne.KeyPageUp:       0x00,
	fyne.KeyPageDown:     0x00,
	fyne.KeyHome:         0x0C,
	fyne.KeyEnd:          0x1A,
	fyne.KeyF1:           0x00,
	fyne.KeyF2:           0x00,
	fyne.KeyF3:           0x00,
	fyne.KeyF4:           0x00,
	fyne.KeyF5:           0x00,
	fyne.KeyF6:           0x00,
	fyne.KeyF7:           0x00,
	fyne.KeyF8:           0x00,
	fyne.KeyF9:           0x00,
	fyne.KeyF10:          0x00,
	fyne.KeyF11:          0x00,
	fyne.KeyF12:          0x00,
	fyne.KeyEnter:        0x0D,
	fyne.Key0:            0x30,
	fyne.Key1:            0x31,
	fyne.Key2:            0x32,
	fyne.Key3:            0x33,
	fyne.Key4:            0x34,
	fyne.Key5:            0x35,
	fyne.Key6:            0x36,
	fyne.Key7:            0x37,
	fyne.Key8:            0x38,
	fyne.Key9:            0x39,
	fyne.KeyA:            0x61,
	fyne.KeyB:            0x62,
	fyne.KeyC:            0x63,
	fyne.KeyD:            0x64,
	fyne.KeyE:            0x65,
	fyne.KeyF:            0x66,
	fyne.KeyG:            0x67,
	fyne.KeyH:            0x68,
	fyne.KeyI:            0x69,
	fyne.KeyJ:            0x6a,
	fyne.KeyK:            0x6b,
	fyne.KeyL:            0x6c,
	fyne.KeyM:            0x6d,
	fyne.KeyN:            0x6e,
	fyne.KeyO:            0x6f,
	fyne.KeyP:            0x70,
	fyne.KeyQ:            0x71,
	fyne.KeyR:            0x72,
	fyne.KeyS:            0x73,
	fyne.KeyT:            0x74,
	fyne.KeyU:            0x75,
	fyne.KeyV:            0x76,
	fyne.KeyW:            0x77,
	fyne.KeyX:            0x78,
	fyne.KeyY:            0x79,
	fyne.KeyZ:            0x7A,
	fyne.KeySpace:        0x20,
	fyne.KeyApostrophe:   0x27,
	fyne.KeyComma:        0x2c,
	fyne.KeyMinus:        0x2d,
	fyne.KeyPeriod:       0x2E,
	fyne.KeySlash:        0x2F,
	fyne.KeyBackslash:    0x5C,
	fyne.KeyLeftBracket:  0x5B,
	fyne.KeyRightBracket: 0x5D,
	fyne.KeySemicolon:    0x3B,
	fyne.KeyEqual:        0x3D,
	fyne.KeyAsterisk:     0x2A,
	fyne.KeyPlus:         0x2B,
	fyne.KeyBackTick:     0x60,
	fyne.KeyUnknown:      0x00,
}

var RemapKeyShift = map[fyne.KeyName]byte{
	fyne.KeyEscape:    0x1B,
	fyne.KeyReturn:    0x0A,
	fyne.KeyTab:       0x09,
	fyne.KeyBackspace: 0x08,
	fyne.KeyInsert:    0x00,
	fyne.KeyDelete:    0x08,
	fyne.KeyRight:     0x18,
	fyne.KeyLeft:      0x08,
	fyne.KeyDown:      0x0A,
	fyne.KeyUp:        0x19,
	fyne.KeyPageUp:    0x00,
	fyne.KeyPageDown:  0x00,
	fyne.KeyHome:      0x0C,
	fyne.KeyEnd:       0x1A,
	fyne.KeyF1:        0x00,
	fyne.KeyF2:        0x00,
	fyne.KeyF3:        0x00,
	fyne.KeyF4:        0x00,
	fyne.KeyF5:        0x00,
	fyne.KeyF6:        0x00,
	fyne.KeyF7:        0x00,
	fyne.KeyF8:        0x00,
	fyne.KeyF9:        0x00,
	fyne.KeyF10:       0x00,
	fyne.KeyF11:       0x00,
	fyne.KeyF12:       0x00,
	fyne.KeyEnter:     0x0D,

	fyne.Key0:            0x29,
	fyne.Key1:            0x21,
	fyne.Key2:            0x40,
	fyne.Key3:            0x23,
	fyne.Key4:            0x24,
	fyne.Key5:            0x25,
	fyne.Key6:            0x5E,
	fyne.Key7:            0x26,
	fyne.Key8:            0x2A,
	fyne.Key9:            0x28,
	fyne.KeyA:            0x41,
	fyne.KeyB:            0x42,
	fyne.KeyC:            0x43,
	fyne.KeyD:            0x44,
	fyne.KeyE:            0x45,
	fyne.KeyF:            0x46,
	fyne.KeyG:            0x47,
	fyne.KeyH:            0x48,
	fyne.KeyI:            0x49,
	fyne.KeyJ:            0x4a,
	fyne.KeyK:            0x4b,
	fyne.KeyL:            0x4c,
	fyne.KeyM:            0x4d,
	fyne.KeyN:            0x4e,
	fyne.KeyO:            0x4f,
	fyne.KeyP:            0x50,
	fyne.KeyQ:            0x51,
	fyne.KeyR:            0x52,
	fyne.KeyS:            0x53,
	fyne.KeyT:            0x54,
	fyne.KeyU:            0x55,
	fyne.KeyV:            0x56,
	fyne.KeyW:            0x57,
	fyne.KeyX:            0x58,
	fyne.KeyY:            0x59,
	fyne.KeyZ:            0x5A,
	fyne.KeySpace:        0x20,
	fyne.KeyApostrophe:   0x22,
	fyne.KeyComma:        0x3C,
	fyne.KeyMinus:        0x5F,
	fyne.KeyPeriod:       0x3E,
	fyne.KeySlash:        0x3F,
	fyne.KeyBackslash:    0x7C,
	fyne.KeyLeftBracket:  0x7B,
	fyne.KeyRightBracket: 0x7D,
	fyne.KeySemicolon:    0x3A,
	fyne.KeyEqual:        0x2B,
	fyne.KeyAsterisk:     0x7E,
	fyne.KeyPlus:         0x7E,
	fyne.KeyBackTick:     0x60,
	fyne.KeyUnknown:      0x00,
}
