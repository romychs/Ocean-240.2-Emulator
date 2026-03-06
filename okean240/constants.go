package okean240

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

// FDC_DRQ Read DRQ state from FDC
const FDC_DRQ = 0x24

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
		mode 000 - intRq on fin,
		     001 - one shot,
		     x10 - rate gen,
		     x11-sq wave
*/
const TMR_DD70CTR = 0x63

/*
 * Programmable Interrupt controller PIC  KR580VV59
 */

const RstKbdNo = 1
const RstTimerNo = 4

const Rst0Mask = 0x01 // System interrupt
const Rst1Mask = 0x02 // Keyboard interrupt
const Rst2Mask = 0x04 // Serial interface interrupt
const RstЗMask = 0x08 // Printer ready
const Rst4Mask = 0x10
const Rst5Mask = 0x20 // Power intRq
const Rst6Mask = 0x40 // User device 1 interrupt
const Rst7Mask = 0x80 // User device 1 interrupt

// PIC_DD75A Port A (a0=0)
const PIC_DD75A = 0x80

// PIC_DD75B Port B (a0=1)
const PIC_DD75B = 0x81

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
