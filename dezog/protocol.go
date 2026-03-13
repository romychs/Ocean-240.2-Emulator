package dezog

// Command - DZRP protocol command
type Command struct {
	length   uint32 // Length of the payload data. (little endian)
	sequence uint8  // Sequence number, 1-255. Increased with each command
	command  uint8  // Command ID
	data     []byte // Payload
}

// Response - response on DZRP protocol command
type Response struct {
	length   uint32 // Length of the following data beginning with the sequence number. (little endian)
	sequence uint8  // Sequence number, same as command.
	data     []byte // Payload
}

// Notification - message from emulator to DEZOG
type Notification struct {
	seqNo       uint8 // Instead of Seq No.
	command     uint8 // NTF_PAUSE = 1
	breakReason uint8 // Break reason: 0 = no reason (e.g. a step-over), 1 = manual break, 2 = breakpoint hit,
	// 3 = watchpoint hit read access, 4 = watchpoint hit write access, 255 = some other reason:
	//the reason string might have useful information for the user
	address uint16 // Breakpoint or watchpoint address.
	bank    uint8  // The bank+1 of the breakpoint or watchpoint address.
	reason  string // 	Null-terminated break reason string. Might in theory have almost 2^32 byte length.
	// In practice, it will be normally less than 256. If reason string is empty it will contain at
	// least a 0.
}

const NTF_PAUSE = 1

// Notification, Break reasons
const (
	BR_REASON_MANUAL   = 1
	BR_REASON_BP_HIT   = 2
	BR_REASON_WP_HIT_R = 3
	BR_REASON_WP_HIT_W = 4
	BR_REASON_OTHER    = 255
)

// DEZOG Commands to emulator
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
	CMD_GET_SPRITES                         = 18
	CMD_GET_SPRITE_PATTERNS                 = 19
	CMD_READ_PORT                           = 20
	CMD_WRITE_PORT                          = 21
	CMD_EXEC_ASM                            = 22
	CMD_INTERRUPT_ON_OFF                    = 23
	CMD_ADD_BREAKPOINT                      = 40
	CMD_REMOVE_BREAKPOINT                   = 41
	CMD_ADD_WATCHPOINT                      = 42
	CMD_REMOVE_WATCHPOINT                   = 43
	CMD_READ_STATE                          = 50
	CMD_WRITE_STATE                         = 51
)
