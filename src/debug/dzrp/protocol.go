package dzrp

import "fmt"

const (
	VersionMajor = 2
	VersionMinor = 1
	VersionPatch = 0
	AppName      = "okemu v1.0.0"
)

type Command struct {
	Len     uint32  // Length of the payload data. (little endian)
	Sn      uint8   // Sequence number, 1-255. Increased with each command
	Id      uint8   // Command ID
	Payload []uint8 // Payload: Data[0]..Data[n-1]
}

type Notification struct {
	Len     uint32  // Length of the following data beginning with the sequence number. (little endian)
	Sn      uint8   // Sequence number = 0
	Payload []uint8 // Payload: Data[0]..Data[n-1]
}

type Response struct {
	Len     uint32  // Length of the following data beginning with the sequence number. (little endian)
	Sn      uint8   // Sequence number, same as command.
	Payload []uint8 // Payload: Data[0]..Data[n-1]
}

// The DRZP commands and responses.
// The response contains the command with the bit 7 set.
const (
	CMD_INIT                                = 1
	CMD_CLOSE                               = 2
	CMD_GET_REGISTERS                       = 3
	CMD_SET_REGISTER                        = 4
	CMD_WRITE_BANK                          = 5
	CMD_CONTINUE                            = 6
	CMD_PAUSE                               = 7
	CMD_READ_MEM                            = 8
	CMD_WRITE_MEM                           = 9
	CMD_SET_SLOT                            = 10
	CMD_GET_TBBLUE_REG                      = 11
	CMD_SET_BORDER                          = 12
	CMD_SET_BREAKPOINTS                     = 13
	CMD_RESTORE_MEM                         = 14
	CMD_LOOPBACK                            = 15
	CMD_GET_SPRITES_PALETTE                 = 16
	CMD_GET_SPRITES_CLIP_WINDOW_AND_CONTROL = 17

	// Sprites
	CMD_GET_SPRITES         = 18
	CMD_GET_SPRITE_PATTERNS = 19

	CMD_READ_PORT        = 20
	CMD_WRITE_PORT       = 21
	CMD_EXEC_ASM         = 22
	CMD_INTERRUPT_ON_OFF = 23

	// Breakpoint
	CMD_ADD_BREAKPOINT    = 40
	CMD_REMOVE_BREAKPOINT = 41

	CMD_ADD_WATCHPOINT    = 42
	CMD_REMOVE_WATCHPOINT = 43

	// State
	CMD_READ_STATE  = 50
	CMD_WRITE_STATE = 51
)

// DZRP notifications.
const NTF_PAUSE = 1

// Machine type that is returned in CMD_INIT.
// It is required to determine the memory model
const (
	MachineUnknown = 0
	MachineZX16K   = 1
	MachineZX48K   = 2
	MachineZX128K  = 3
	MachineZXNEXT  = 4
)

const (
	RegPC = iota
	RegSP
	RegAF
	RegBC
	RegDE
	RegHL
	RegIX
	RegIY
	RegAF_
	RegBC_
	RegDE_
	RegHL_
	RegUnk
	RegIM
	RegF
	RegA
	RegC
	RegB
	RegE
	RegD
	RegL
	RegH
	RegIXL
	RegIXH
	RegIYL
	RegIYH
	RegF_
	RegA_
	RegC_
	RegB_
	RegE_
	RegD_
	RegL_
	RegH_
	RegR
	RegI
)

type CmdInitCommand struct {
	Major   uint8 // Version (of the command sender): 3 bytes, big endian: Major.Minor.Patch
	Minor   uint8
	Patch   uint8
	AppName string // 0-terminated string	The program name + version as a string. E.g. "DeZog v1.4.0"
}

type CmdInitResponse struct {
	Sn      uint8 // Same seq no
	Error   uint8 // Error: 0=no error, 1=general (unknown) error.
	Major   uint8 // Version (of the response sender) : 3 bytes, big endian: Major.Minor.Patch
	Minor   uint8
	Patch   uint8
	Machine uint8  // Machine type (memory model): 0 = UNKNOWN, 1 = ZX16K, 2 = ZX48K, 3 = ZX128K, 4 = ZXNEXT.
	AppName string //	0-terminated string	The responding program name + version as a string. E.g. "dbg_uart_if v2.0.0"}
}

const (
	BprStepOver = 0
	BprManual   = 1
	BprHit      = 2
	BprMemRead  = 3
	BprMemWrite = 4
	BprOther    = 255
)

var BprReasons = map[int]string{
	BprStepOver: "Step-over",
	BprManual:   "Manual break",
	BprHit:      "Hit",
	BprMemRead:  "WP read",
	BprMemWrite: "WP Write",
	BprOther:    "Other",
}

func (c *Command) toString() string {
	return fmt.Sprintf("Len: %d, Sn: %d, Id: %d, Payload: %s", c.Len, c.Sn, c.Id, PayloadToString(c.Payload))
}

func NewResponse(cmd *Command, payload []byte) *Response {
	return &Response{
		Len:     uint32(len(payload) + 1),
		Sn:      cmd.Sn,
		Payload: payload,
	}
}

func PayloadToString(payload []byte) string {
	res := "["
	for _, b := range payload {
		res += fmt.Sprintf("%02X ", b)
	}
	return res + "]"
}
