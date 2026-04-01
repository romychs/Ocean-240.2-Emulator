package dzrp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/breakpoint"
	"okemu/okean240"
	"okemu/z80/dis"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type DZRP struct {
	port         string
	config       *config.OkEmuConfig
	debugger     *debug.Debugger
	disassembler *dis.Disassembler
	computer     *okean240.ComputerType
	conn         net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
	cmd          *Command
}

// SetupTcpHandler Setup TCP listener, handle connections

func NewDZRP(config *config.OkEmuConfig, debug *debug.Debugger, dissasm *dis.Disassembler, comp *okean240.ComputerType) *DZRP {
	return &DZRP{
		port:         config.Debugger.Host + ":" + strconv.Itoa(config.Debugger.Port),
		debugger:     debug,
		disassembler: dissasm,
		computer:     comp,
	}
}

func (p *DZRP) SetupTcpHandler() {

	var err error
	var l net.Listener

	l, err = net.Listen("tcp4", p.port)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			log.Warnf("Error closing listener connection %v", err)
		}
	}(l)

	log.Infof("Ready for debugger connections on %s", p.port)
	for {
		p.conn, err = l.Accept()
		if err != nil {
			log.Errorf("Accept connection: %v", err)
			return
		}
		go p.handleConnection()
	}

}

// Receive messages, split to strings and parse
func (p *DZRP) handleConnection() {
	p.reader = bufio.NewReader(p.conn)
	p.writer = bufio.NewWriter(p.conn)
	var command Command
	n := 0
	for {
		// receive command packet byte by byte
		b, err := p.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Errorf("TCP error: %v", err)
				p.debugger.SetStepMode(false)
				return
			}
		}
		switch n {
		case 0:
			command.Len = uint32(b)
		case 1:
			command.Len |= uint32(b) << 8
		case 2:
			command.Len |= uint32(b) << 16
		case 3:
			command.Len |= uint32(b) << 24
		case 4:
			command.Sn = b
		case 5:
			command.Id = b
		default:
			command.Payload = append(command.Payload, b)
		}
		if n >= 5 && int(command.Len) == len(command.Payload) {
			p.cmd = &command
			if p.handleCommand() {
				break
			}
			command.Len = 0
			command.Payload = []uint8{}
			n = 0
		} else {
			n++
		}
	}
	log.Debug("Closing connection")

	p.debugger.SetStepMode(false)
	_ = p.writer.Flush()
	p.writer = nil
	p.reader = nil
	err := p.conn.Close()
	if err != nil {
		log.Warnf("Can not close TCP socket: %v", err)
	}
}

// writeResponse Write response to command back to client
func (p *DZRP) writeResponse(response *Response) error {
	//	log.Infof("Ready for debugger connections on %s", port)
	// Send Len
	l := response.Len
	for c := 0; c < 4; c++ {
		e := p.writer.WriteByte(byte(l))
		if e != nil {
			log.Warnf("Error writing Len: %v", e)
		}
		l = l >> 8
	}
	// Send Sn
	e := p.writer.WriteByte(response.Sn)
	if e != nil {
		log.Warnf("Error writing Sn: %v", e)
		return e
	}
	// Send Payload
	for _, b := range response.Payload {
		e := p.writer.WriteByte(b)
		if e != nil {
			log.Warnf("Error writing payload: %v", e)
			return e
		}
	}
	e = p.writer.Flush()
	if e != nil {
		log.Warnf("Error flushing response: %v", e)
		return e
	}
	return nil
}

// BreakpointHit Send notification message to Dezog on BP Hit
func (p *DZRP) BreakpointHit(number uint16, typ byte) {
	// turn off temporary breakpoints on any hit
	p.debugger.SetBreakpointEnabled(breakpoint.BpTmp1, false)
	p.debugger.SetBreakpointEnabled(breakpoint.BpTmp2, false)
	if p.writer != nil {
		err := p.writeResponse(p.buildBpHitResponse(number, typ))
		if err != nil {
			log.Warnf("NTF_PAUSE, write response err: %v", err)
		}
	} else {
		log.Warn("NTF_PAUSE, writer is nil")
	}
}

func (p *DZRP) buildBpHitResponse(number uint16, typ byte) *Response {
	bpReason, bpAddr, mBank := p.getBpReasonAndAddr(number, typ)
	rsn := getBpReason(bpReason, bpAddr)
	// msg - ASCIIZ bytes
	msg := []byte(rsn)
	msg = append(msg, byte(0))

	rep := []byte{
		1, // NTF_PAUSE
		bpReason,
		byte(bpAddr), byte(bpAddr >> 8),
		mBank, // bank
	}
	rep = append(rep, msg...)

	log.Debugf("NTF_PAUSE %s resp: %v", rsn, PayloadToString(rep))

	return &Response{
		Len:     uint32(len(rep) + 1),
		Sn:      0,
		Payload: rep,
	}
}

// handleCommand Handle command received from client, return true if last command is CMD_CLOSE
func (p *DZRP) handleCommand() bool {
	log.Debugf("Handling command: %s", p.cmd.toString())
	var err error
	var resp *Response
	handler, ok := commandHandlers[int(p.cmd.Id)]
	if ok {
		resp, err = handler(p)
	} else {
		//resp = NewResponse(p.cmd, nil)
		err = errors.New("unknown command Id: " + strconv.Itoa(int(p.cmd.Id)))
	}
	if err == nil {
		err = p.writeResponse(resp)
	}
	if err != nil {
		log.Errorf("Error handling command: %v", err)
	}
	return p.cmd.Id == CMD_CLOSE
}

type CommandHandler func(*DZRP) (*Response, error)

var commandHandlers = map[int]CommandHandler{
	CMD_INIT:              (*DZRP).handleCmdInit,             //1
	CMD_CLOSE:             (*DZRP).handleCmdClose,            //2
	CMD_GET_REGISTERS:     (*DZRP).handleCmdGetRegisters,     //3
	CMD_SET_REGISTER:      (*DZRP).handleCmdSetRegister,      //4
	CMD_CONTINUE:          (*DZRP).handleCmdContinue,         //6
	CMD_PAUSE:             (*DZRP).handleCmdPause,            // 7
	CMD_READ_MEM:          (*DZRP).handleCmdReadMem,          //8
	CMD_WRITE_MEM:         (*DZRP).handleCmdWriteMem,         //9
	CMD_LOOPBACK:          (*DZRP).handleCmdLoopback,         // 15
	CMD_READ_PORT:         (*DZRP).handleCmdReadPort,         // 20
	CMD_WRITE_PORT:        (*DZRP).handleCmdWritePort,        // 21
	CMD_INTERRUPT_ON_OFF:  (*DZRP).handleCmdInterruptOnOff,   // 23
	CMD_ADD_BREAKPOINT:    (*DZRP).handleCmdAddBreakpoint,    //40
	CMD_REMOVE_BREAKPOINT: (*DZRP).handleCmdRemoveBreakpoint, //41

}

func (p *DZRP) handleCmdInit() (*Response, error) {
	if len(p.cmd.Payload) < 4 {
		return nil, errors.New("too short payload")
	}
	p.debugger.SetStepMode(true)
	p.debugger.SetBreakpointsEnabled(true)
	p.debugger.ClearBreakpoints()
	p.debugger.ClearMemBreakpoints()

	app := string(p.cmd.Payload[3 : len(p.cmd.Payload)-1])
	log.Debugf("CMD_INIT: client version %d.%d.%d App: %s", p.cmd.Payload[0], p.cmd.Payload[1], p.cmd.Payload[2], app)

	payload := []byte{0, VersionMajor, VersionMinor, VersionPatch, MachineZX128K}
	payload = append(payload, []byte(AppName)...)
	payload = append(payload, 0)
	return NewResponse(p.cmd, payload), nil
}

func (p *DZRP) handleCmdClose() (*Response, error) {
	log.Debug("CMD_CLOSE")
	return NewResponse(p.cmd, nil), nil
}

func (p *DZRP) handleCmdPause() (*Response, error) {
	log.Debug("CMD_PAUSE")
	p.debugger.SetStepMode(true)
	//return NewResponse(p.cmd, []byte{p.cmd.Sn}), nil
	return NewResponse(p.cmd, nil), nil
}

func (p *DZRP) handleCmdReadMem() (*Response, error) {
	if len(p.cmd.Payload) < 5 {
		return nil, errors.New("too short payload")
	}
	addr := uint16(p.cmd.Payload[2])<<8 + uint16(p.cmd.Payload[1])
	size := uint16(p.cmd.Payload[4])<<8 + uint16(p.cmd.Payload[3])

	log.Debugf("CMD_READ_MEM[0x%04X] len: 0x%04X", addr, size)
	mem := make([]byte, size)
	//mem[0] = cmd.Sn
	for i := 0; i < int(size); i++ {
		mem[i] = p.computer.MemRead(addr)
		addr++
	}
	return NewResponse(p.cmd, mem), nil
}

func (p *DZRP) handleCmdWriteMem() (*Response, error) {
	if len(p.cmd.Payload) < 4 {
		return nil, errors.New("too short payload")
	}
	addr := uint16(p.cmd.Payload[1]) + uint16(p.cmd.Payload[2])<<8
	log.Debugf("CMD_WRITE_MEM[0x%04X] len: 0x%04X", addr, len(p.cmd.Payload)-3)
	for i := 3; i < len(p.cmd.Payload); i++ {
		p.computer.MemWrite(addr, p.cmd.Payload[i])
		addr++
	}
	return NewResponse(p.cmd, []byte{p.cmd.Sn}), nil
}

func (p *DZRP) handleCmdSetRegister() (*Response, error) {
	lo := p.cmd.Payload[1]
	hi := p.cmd.Payload[2]
	word := uint16(hi)<<8 | uint16(lo)
	s := p.computer.CPUState()

	switch p.cmd.Payload[0] {
	case RegPC:
		s.PC = word
		log.Debugf("Set PC=0x%04X", s.PC)
	case RegSP:
		s.SP = word
		log.Debugf("Set SP=0x%04X", s.PC)
	case RegAF:
		s.A = hi
		s.Flags.SetFlags(lo)
	case RegBC:
		s.B = hi
		s.C = lo
	case RegDE:
		s.D = hi
		s.E = lo
	case RegHL:
		s.H = hi
		s.L = lo
	case RegIX:
		s.IX = word
	case RegIY:
		s.IY = word
	case RegAF_:
		s.AAlt = hi
		s.FlagsAlt.SetFlags(lo)
	case RegBC_:
		s.BAlt = hi
		s.CAlt = lo
	case RegDE_:
		s.DAlt = hi
		s.EAlt = lo
	case RegHL_:
		s.HAlt = hi
		s.LAlt = lo
	case RegIM:
		s.IMode = lo
	case RegF:
		s.Flags.SetFlags(lo)
	case RegA:
		s.A = lo
	case RegC:
		s.C = lo
	case RegB:
		s.B = lo
	case RegE:
		s.E = lo
	case RegD:
		s.D = lo
	case RegL:
		s.L = lo
	case RegH:
		s.H = lo
	case RegIXL:
		s.IX = s.IX&0xff00 | uint16(lo)
	case RegIXH:
		s.IX = (uint16(lo) << 8) | (s.IX & 0x00ff)
	case RegIYL:
		s.IY = s.IY&0xff00 | uint16(lo)
	case RegIYH:
		s.IY = (uint16(lo) << 8) | (s.IY & 0x00ff)
	case RegF_:
		s.FlagsAlt.SetFlags(lo)
	case RegA_:
		s.AAlt = lo
	case RegC_:
		s.CAlt = lo
	case RegB_:
		s.BAlt = lo
	case RegE_:
		s.EAlt = lo
	case RegD_:
		s.DAlt = lo
	case RegL_:
		s.LAlt = lo
	case RegH_:
		s.HAlt = lo
	case RegR:
		s.R = lo
	case RegI:
		s.I = lo
	default:
		return nil, errors.New("unknown register no: " + strconv.Itoa(int(p.cmd.Payload[0])))
	}
	p.computer.SetCPUState(s)
	//return NewResponse(p.cmd, []byte{p.cmd.Sn}), nil
	return NewResponse(p.cmd, nil), nil
}

func (p *DZRP) handleCmdGetRegisters() (*Response, error) {
	s := p.computer.CPUState()
	resp := []byte{
		//p.cmd.Sn,
		byte(s.PC), byte(s.PC >> 8),
		byte(s.SP), byte(s.SP >> 8),
		s.Flags.GetFlags(), s.A,
		s.C, s.B,
		s.E, s.D,
		s.L, s.H,
		byte(s.IX), byte(s.IX >> 8),
		byte(s.IY), byte(s.IY >> 8),
		s.FlagsAlt.GetFlags(), s.AAlt,
		s.CAlt, s.BAlt,
		s.EAlt, s.DAlt,
		s.LAlt, s.HAlt,
		s.R,
		s.I,
		s.IMode,
		0, // reserved
		0, // Nslots. The number of slots that will follow.
	}
	log.Debugf("CMD_GET_REGISTERS resp: %v", resp)
	return NewResponse(p.cmd, resp), nil
}

func (p *DZRP) handleCmdContinue() (*Response, error) {

	eb1 := p.cmd.Payload[0] != 0
	ab1 := (uint16(p.cmd.Payload[2]) << 8) | uint16(p.cmd.Payload[1])
	p.setTempBreakpoint(eb1, breakpoint.BpTmp1, ab1)
	eb2 := p.cmd.Payload[3] != 0
	ab2 := (uint16(p.cmd.Payload[5]) << 8) | uint16(p.cmd.Payload[4])
	p.setTempBreakpoint(eb2, breakpoint.BpTmp2, ab2)

	log.Debugf("CMD_CONTINUE BP1 en: %v, addr: 0x%04X; BP2  en: %v, addr: 0x%04X; AC: 0x%02X",
		eb1, ab1, eb2, ab2, p.cmd.Payload[6])

	p.debugger.SetRunMode(true)

	//return NewResponse(p.cmd, []byte{p.cmd.Sn}), nil
	return NewResponse(p.cmd, nil), nil
}

func (p *DZRP) setTempBreakpoint(ena bool, no uint16, addr uint16) {
	if ena {
		e := p.debugger.SetBreakpoint(no, fmt.Sprintf("PC=%04Xh", addr), 0)
		if e != nil {
			log.Debugf("setTmpBreakpoint err: %v", e)
		}
		p.debugger.SetBreakpointEnabled(no, true)
	} else {
		p.debugger.SetBreakpointEnabled(no, false)
	}
}

func (p *DZRP) getBpReasonAndAddr(number uint16, typ byte) (reason byte, addr uint16, mBank uint8) {
	reason = BprStepOver // default StepOver
	addr = number
	mBank = uint8(1)
	if typ >= 1 && typ <= 2 { // 1-rd, 2-wr
		reason = typ + 2
	} else {
		if number != breakpoint.BpTmp1 && number != breakpoint.BpTmp2 {
			// bp hit
			mBank = p.debugger.BreakpointMBank(addr)
			reason = BprHit
		}
		addr = p.computer.CPUState().PC
	}
	return reason, addr, mBank
}

func (p *DZRP) handleCmdAddBreakpoint() (*Response, error) {
	if len(p.cmd.Payload) < 4 {
		return nil, errors.New("too short payload")
	}
	addr := (uint16(p.cmd.Payload[1]) << 8) | uint16(p.cmd.Payload[0])
	mBank := p.cmd.Payload[2]
	isCond := p.cmd.Payload[3] != 0
	var expr string
	if isCond {
		expr = string(p.cmd.Payload[3 : len(p.cmd.Payload)-3])
	} else {
		expr = fmt.Sprintf("PC=%04Xh", addr)
	}
	log.Debugf("CMD_ADD_BREAKPOINT addr: 0x%04X, bank: %d, cond: '%s'", addr, mBank, expr)
	bpNum, err := p.debugger.AddBreakpoint(expr, mBank)
	if err != nil {
		log.Debugf("SetBreakpoint err: %v", err)
		return nil, err
	}
	if bpNum == breakpoint.MaxBreakpoints {
		bpNum = 0
	}
	//bpNum := p.debugger.GetBreakpointNum()
	//if bpNum < breakpoint.MaxBreakpoints {
	//	err := p.debugger.SetBreakpoint(bpNum, cond, mBank)
	//	p.debugger.SetBreakpointEnabled(bpNum, true)
	//	if err != nil {
	//		log.Debugf("SetBreakpoint err: %v", err)
	//		return nil, err
	//	}
	//} else {
	//	bpNum = 0
	//}
	return NewResponse(p.cmd, []byte{byte(bpNum), byte(bpNum >> 8)}), nil
}

func getBpReason(reason byte, addr uint16) string {
	rsn, ok := BprReasons[int(reason)]
	if !ok {
		rsn = strconv.Itoa(int(reason))
	}
	return fmt.Sprintf("BP: '%s', at: 0x%04X", rsn, addr)
}

func (p *DZRP) handleCmdRemoveBreakpoint() (*Response, error) {
	if len(p.cmd.Payload) < 2 {
		return nil, errors.New("too short payload")
	}
	bpNum := (uint16(p.cmd.Payload[1]) << 8) | uint16(p.cmd.Payload[0])
	p.debugger.RemoveBreakpoint(bpNum)
	log.Debugf("CMD_REMOVE_BREAKPOINT no: %d", bpNum)
	return NewResponse(p.cmd, nil), nil
}

// handleCmdLoopback send back received data
func (p *DZRP) handleCmdLoopback() (*Response, error) {
	return NewResponse(p.cmd, p.cmd.Payload), nil
}

func (p *DZRP) handleCmdReadPort() (*Response, error) {
	if len(p.cmd.Payload) < 2 {
		return nil, errors.New("too short payload")
	}
	addr := (uint16(p.cmd.Payload[1]) << 8) | uint16(p.cmd.Payload[0])
	return NewResponse(p.cmd, []byte{p.computer.IORead(addr)}), nil
}

func (p *DZRP) handleCmdWritePort() (*Response, error) {
	if len(p.cmd.Payload) < 3 {
		return nil, errors.New("too short payload")
	}
	addr := (uint16(p.cmd.Payload[1]) << 8) | uint16(p.cmd.Payload[0])
	p.computer.IOWrite(addr, p.cmd.Payload[2])
	return NewResponse(p.cmd, nil), nil
}

func (p *DZRP) handleCmdInterruptOnOff() (*Response, error) {
	if len(p.cmd.Payload) == 0 {
		return nil, errors.New("too short payload")
	}
	on := p.cmd.Payload[0] != 0
	p.debugger.SetBreakpointsEnabled(on)
	log.Debugf("CMD_INTERRUPT_ONOFF on: %t", on)
	return NewResponse(p.cmd, nil), nil
}
