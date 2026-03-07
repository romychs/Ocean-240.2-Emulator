package okean240

/*
 * КР580ВВ55 DD79  USER PORT
 */

// USR_DD79PA User port A
//const USR_DD79PA = 0x00

// USR_DD79PB User port B
//const USR_DD79PB = 0x01

// USR_DD79PC User port C
//const USR_DD79PC = 0x02

// USR_DD79CTR Config
// Configure: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]
//const USR_DD79CTR = 0x03

/*
 * КР1818ВГ93 FDC Controller
 */

// FdcCmd FDC Command
const FdcCmd = 0x20

// FdcTrack FDC Track No
const FdcTrack = 0x21

// FdcSect FDC Sector
const FdcSect = 0x22

// FdcData FDC Data
const FdcData = 0x23

// FdcDrq Read DRQ state from FDC
const FdcDrq = 0x24

/*
 * Floppy Controller port
 */

// Floppy Controller port
// WR: 5-SSEN, 4-#DDEN, 3-INIT, 2-DRSEL, 1-MOT1, 0-MOT0
// RD: 7-MOTST, 6-SSEL, 5,4-x , 3-DRSEL, 2-MOT1, 1-MOT0, 0-INT
const Floppy = 0x25

/*
 * КР580ВВ55 DD78  Keyboard
 */

// KbdDd78pa Port A - Keyboard Data
const KbdDd78pa = 0x40

// KbdDd78pb Port B - JST3,SHFT,CTRL,ACK,TAPE5,TAPE4,GK,GC
const KbdDd78pb = 0x41

// KBD_DD78PC Port C - [PC7:5],[KBD_ACK],[PC3:0]
//const KBD_DD78PC = 0x42

// KBD_DD78CTR Control port
// Configure: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]
// const KBD_DD78CTR = 0x43

/*
 * КР580ВИ53 DD70
 */

// TmrDd70c1 Timer load 1
const TmrDd70c1 = 0x60

// TmrDd70c2 Timer load 2
const TmrDd70c2 = 0x61

// TmrDd70c3 Timer load 3
const TmrDd70c3 = 0x62

// TmrDd70ctr
// Timer config: [sc1,sc0][rl1,rl0][m2,m1,m0][bcd]
//
//	   sc - timer, rl=01-LSB, 10-MSB, 11-LSB+MSB
//		  mode 000 - intRq on fin,
//			   001 - one shot,
//			   x10 - rate gen,
//			   x11-sq wave
const TmrDd70ctr = 0x63

/*
 * Programmable Interrupt controller PIC  КР580ВН59
 */

const RstKbdNo = 1
const RstTimerNo = 4

//const Rst0Mask = 0x01 // System interrupt
//const Rst1Mask = 0x02 // Keyboard interrupt
//const Rst2Mask = 0x04 // Serial interface interrupt
//const Rst3Mask = 0x08 // Printer ready
//const Rst4Mask = 0x10
//const Rst5Mask = 0x20 // Power intRq
//const Rst6Mask = 0x40 // User device 1 interrupt
//const Rst7Mask = 0x80 // User device 1 interrupt

// PicDd75a Port A (a0=0)
//const PicDd75a = 0x80

// PIC_DD75B Port B (a0=1)
//const PIC_DD75B = 0x81

/*
 * КР580ВВ51 DD72
 */

// UartDd72rd Serial data
const UartDd72rd = 0xA0

// UartDd72rr Serial status [RST,RQ_RX,RST_ERR,PAUSE,RX_EN,RX_RDY,TX_RDY]
const UartDd72rr = 0xA1

/*
 * КР580ВВ55 DD17 System port
 */

// SysDd17pa Port A - VShift[8..1] Vertical shift
const SysDd17pa = 0xC0

// SysDd17pb Port B - Memory mapper [ROM14,13][REST][ENROM-][A18,17,16][32k]
const SysDd17pb = 0xC1

// SysDd17pc Port C - HShift[HS5..1,SB3..1] Horisontal shift
const SysDd17pc = 0xC2

// SysDd17ctr  Configure: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]
const SysDd17ctr = 0xC3

/*
 * КР580ВВ55 DD67
 */

// LPT_DD67PA Port A - Printer Data
//const LPT_DD67PA = 0xE0

// VID_DD67PB Port B - Video control [VSU,C/M,FL3:1,COL3:1]
//const VID_DD67PB = 0xE1

// DD67PC Port C - [USER3:1, STB-LP, BELL, TAPE3:1]
//const DD67PC = 0xE2

// DD67CTR
// Configure: [1][ma1,ma0][0-aO|1-aI],[0-chO,1-chI],[mb],[0-bO|1-bI],[0-clO,1-clI]
// Set bit: [0][xxx][bbb][0|1]
// const DD67CTR = 0xE3
