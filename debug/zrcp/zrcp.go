package zrcp

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
	"okemu/z80"
	"okemu/z80/dis"
	"os"
	"strings"
	//"okemu/logger"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type ZRCP struct {
	port         string
	config       *config.OkEmuConfig
	debugger     *debug.Debugger
	disassembler *dis.Disassembler
	computer     *okean240.ComputerType
	conn         net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
	params       string
	//cmd          *Command
}

func NewZRCP(config *config.OkEmuConfig, debug *debug.Debugger, disassm *dis.Disassembler, comp *okean240.ComputerType) *ZRCP {
	return &ZRCP{
		port:         config.Debugger.Host + ":" + strconv.Itoa(config.Debugger.Port),
		debugger:     debug,
		disassembler: disassm,
		computer:     comp,
	}
}

// SetupTcpHandler Setup TCP listener, handle connections
func (p *ZRCP) SetupTcpHandler() {
	l, err := net.Listen("tcp4", p.port)
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
		var err error
		p.conn, err = l.Accept()
		if err != nil {
			log.Errorf("Accept connection: %v", err)
			return
		}
		go p.handleConnection()
	}

}

// Receive messages, split to strings and parse
func (p *ZRCP) handleConnection() {
	p.reader = bufio.NewReader(p.conn)
	p.writer = bufio.NewWriter(p.conn)

	if !p.writeWelcomeMessage() {
		return
	}
	for {
		str, err := p.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Errorf("TCP error: %v", err)
				p.debugger.SetStepMode(false)
				return
			}
		}
		if !p.handleCommand(str) {
			log.Debug("Closing connection")
			p.writeResponseMessage(quitResponse)
			break
		}
		//byteBuffer.WriteByte(b)
	}
	p.debugger.SetStepMode(false)
	err := p.conn.Close()
	if err != nil {
		log.Warnf("Can not close socket: %v", err)
	}
}

//var activeWriter *bufio.Writer = nil

func (p *ZRCP) writeWelcomeMessage() bool {
	return p.writeResponseMessage(welcomeMessage)
}

func (p *ZRCP) writeResponseMessage(message string) bool {
	if message == "-" {
		return true
	}
	prompt := emptyResponse
	if p.debugger.StepMode() {
		prompt = inCpuStepResponse
	}

	_, err := p.writer.WriteString(message + prompt)
	if err != nil {
		log.Errorf("TCP write error: %v", err)
		return false
	}
	err = p.writer.Flush()
	if err != nil {
		log.Errorf("TCP flush error: %v", err)
		return false
	}
	return true
}

func (p *ZRCP) writeMessage(message string) bool {
	_, err := p.writer.WriteString(message)
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	err = p.writer.Flush()
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	return true
}

// HandleCommand HandleLogLine Parse log line(s) and send it to redis
func (p *ZRCP) handleCommand(str string) bool {
	str = strings.TrimSpace(str)
	if str == "" {
		return false
	}
	log.Debugf("Command: '%s'", str)

	pos := strings.Index(str, " ")
	cmd := str
	p.params = ""

	if pos > 1 {
		cmd = str[:pos]
		p.params = strings.TrimSpace(str[pos+1:])
	}
	var err error
	var resp string

	if cmd == "quit" {
		return false
	}

	handler, ok := commandHandlers[cmd]
	if ok {
		resp, err = handler(p)
		if err != nil {
			log.Errorf("%v", err)
			p.writeResponseMessage(err.Error())
		} else {
			p.writeResponseMessage(resp)
		}
	} else {
		log.Debugf("Unhandled Command: %s", str)
		p.writeResponseMessage("")
	}
	return true
}

type CommandHandler func(zrcp *ZRCP) (string, error)

var commandHandlers = map[string]CommandHandler{
	"about":                   (*ZRCP).handleAbout,
	"clear-membreakpoints":    (*ZRCP).handleClearMemBreakpoints,
	"close-all-menus":         (*ZRCP).handleEmptyHandler,
	"cpu-code-coverage":       (*ZRCP).handleCPUCodeCoverage,
	"cpu-history":             (*ZRCP).handleCPUHistory,
	"cpu-step":                (*ZRCP).handleCpuStep,
	"disable-breakpoint":      (*ZRCP).handleDisableBreakpoint,
	"disable-breakpoints":     (*ZRCP).handleDisableBreakpoints,
	"disassemble":             (*ZRCP).handleDisassemble,
	"enable-breakpoint":       (*ZRCP).handleEnableBreakpoint,
	"enable-breakpoints":      (*ZRCP).handleEnableBreakpoints,
	"enter-cpu-step":          (*ZRCP).handleEnterCPUStep,
	"exit-cpu-step":           (*ZRCP).handleExitCPUStep,
	"extended-stack":          (*ZRCP).handleExtendedStack,
	"get-cpu-frequency":       (*ZRCP).handleGetCPUFrequency,
	"get-current-machine":     (*ZRCP).handleGetCurrentMachine,
	"get-registers":           (*ZRCP).handleGetRegisters,
	"get-tstates-partial":     (*ZRCP).handleGetTStatesPartial,
	"get-version":             (*ZRCP).handleGetVersion,
	"hard-reset-cpu":          (*ZRCP).handleHardResetCPU,
	"load-binary":             (*ZRCP).handleLoadBinary,
	"read-memory":             (*ZRCP).handleReadMemory,
	"reset-tstates-partial":   (*ZRCP).handleResetTStatesPartial,
	"run":                     (*ZRCP).handleRun,
	"set-breakpoint":          (*ZRCP).handleSetBreakpoint,
	"set-breakpointaction":    (*ZRCP).handleEmptyHandler,
	"set-breakpointpasscount": (*ZRCP).handleSetBreakpointPassCount,
	"set-debug-settings":      (*ZRCP).handleEmptyHandler,
	"set-membreakpoint":       (*ZRCP).handleSetMemBreakpoint,
	"set-register":            (*ZRCP).handleSetRegister,
	"snapshot-load":           (*ZRCP).handleSnapshotLoad,
	"snapshot-save":           (*ZRCP).handleSnapshotSave,
}

func (p *ZRCP) handleCpuStep() (string, error) {
	p.debugger.SetDoStep(true) // computer.Do()
	text := p.disassembler.Disassm(p.computer.CPUState().PC)
	return registersResponse(p.computer.CPUState()) + " TSTATES: " + strconv.Itoa(int(p.computer.TStatesPartial())) + "\n" + text, nil
}

func (p *ZRCP) handleRun() (string, error) {
	p.writeMessage(runUntilBPMessage)
	p.debugger.SetRunMode(true)
	return "-", nil
}

func (p *ZRCP) handleDisassemble() (string, error) {
	return p.disassemble(p.params), nil
}

func convertToUint16(s string) (uint16, error) {
	v := strings.TrimSpace(strings.ToUpper(s))
	base := 0
	if strings.HasSuffix(v, "H") {
		v = strings.TrimSuffix(v, "H")
		base = 16
	}
	a, e := strconv.ParseUint(v, base, 16)
	return uint16(a), e
}

func (p *ZRCP) SetMemBreakpoint(param string) string {
	param = strings.TrimSpace(param)
	params := strings.Split(param, " ")
	if len(params) < 1 {
		return "error, not enough parameters"
	}
	address, err := convertToUint16(params[0])
	if err != nil {
		return "error, illegal address: '" + params[0] + "'"
	}
	t := uint16(3)
	// if has type
	if len(params) > 1 {
		t, err = convertToUint16(params[1])
		if err != nil || t > 3 {
			return "error, illegal access type: '" + params[1] + "'"
		}
	}

	s := uint16(1)
	if len(params) > 2 {
		s, err = convertToUint16(params[2])
		if err != nil {
			return "error, illegal memory size: '" + params[2] + "'"
		}
	}
	if p.debugger != nil {
		p.debugger.SetMemBreakpoint(address, byte(t), s)
	}
	return ""
}

func (p *ZRCP) handleCPUHistory() (string, error) {
	params := strings.Split(p.params, " ")
	if len(params) < 1 {
		return "", errors.New("error, no parameters")
	}

	cmd := params[0]
	nspe := errors.New("error, no second parameter")

	switch cmd {

	case "enabled":
		if len(params) < 2 {
			return "", nspe
		}
		p.debugger.SetCpuHistoryEnabled(params[1] == "yes")

	case "clear":
		p.debugger.CpuHistoryClear()

	case "started":
		if len(params) < 2 {
			return "", nspe
		}
		p.debugger.SetCpuHistoryStarted(params[1] == "yes")
	case "set-max-size":
		if len(params) != 2 {
			return "", nspe
		}
		size, err := strconv.Atoi(params[1])
		if err != nil {
			return "", errors.New("error, illegal number")
		}
		p.debugger.SetCpuHistoryMaxSize(size)
	case "get":
		if len(params) != 2 {
			return "", nspe
		}
		index, err := strconv.Atoi(params[1])
		if err != nil {
			return "", errors.New("error, illegal number")
		}
		history := p.debugger.CpuHistory(index)
		if history != nil {
			return p.stateResponse(history), nil
		}
		return "", errors.New("ERROR: index out of range")
	case "ignrephalt":
		// ignore
	default:
		return "", errors.New("error: unknown history command: " + cmd)
	}

	return "", nil
}

func (p *ZRCP) handleLoadBinary() (string, error) {
	params := strings.Split(p.params, " ")
	loadError := errors.New(respErrorLoading)
	if len(params) < 2 {
		return "", loadError
	}
	fn := strings.TrimSpace(params[0])
	if strings.HasPrefix(fn, "\"") {
		fn = fn[1:]
	}
	if strings.HasSuffix(fn, "\"") && len(fn) > 1 {
		fn = fn[:len(fn)-1]
	}
	offset, e := strconv.Atoi(params[1])
	length := 0
	if e != nil || offset < 0 || offset > 65535 || len(fn) == 0 {
		return "", loadError
	}
	if len(params) > 2 {
		l, e := strconv.ParseInt(params[2], 0, 32)
		if e != nil {
			length = 0
		} else {
			length = int(l)
		}
	}
	data, err := os.ReadFile(fn)
	if err != nil {
		return "", loadError
	}
	if length != 0 && len(data) != length {
		log.Warnf("File size does not match the specified length. Expected %d bytes, got %d.", length, len(data))
		//return respErrorLoading
		length = len(data)
	}
	if length == 0 {
		length = len(data)
	}
	// Loaded Ok, move file to memory
	for addr := uint16(0); addr < uint16(length); addr++ {
		p.computer.MemWrite(addr+uint16(offset), data[addr])
	}
	return "", nil
}

func toW(hi, lo byte) uint16 {
	return uint16(lo) | (uint16(hi) << 8)
}

func iifStr(iif1, iif2 bool) string {
	flags := []byte{'-', '-'}
	if iif1 {
		flags[0] = '1'
	}
	if iif2 {
		flags[1] = '2'
	}
	return string(flags)
}

// registersResponse Build string
// PC=%4x SP=%4x AF=%4x BC=%4x HL=%4x DE=%4x IX=%4x IY=%4x AF'=%4x BC'=%4x HL'=%4x DE'=%4x I=%2x
// R=%2x  F=%s F'=%s MEMPTR=%4x IM0 IFF-- VPS: 0 MMU=00000000000000000000000000000000
func registersResponse(state *z80.CPU) string {
	resp := fmt.Sprintf(getRegistersResponse,
		state.PC,
		state.SP,
		toW(state.A, state.Flags.GetFlags()),
		toW(state.B, state.C),
		toW(state.H, state.L),
		toW(state.D, state.E),
		state.IX,
		state.IY,
		toW(state.AAlt, state.FlagsAlt.GetFlags()),
		toW(state.BAlt, state.CAlt),
		toW(state.HAlt, state.LAlt),
		toW(state.DAlt, state.EAlt),
		state.I,
		state.R,
		state.Flags.GetFlagsStr(),
		state.FlagsAlt.GetFlagsStr(),
		state.MemPtr,
		iifStr(state.Iff1, state.Iff2),
	)
	log.Trace(resp)
	return resp
}

func (p *ZRCP) getNBytes(addr uint16, n uint16) string {
	res := ""
	for i := uint16(0); i < n; i++ {
		b := p.computer.MemRead(addr + i)
		res += fmt.Sprintf("%02X", b)
	}
	return res
}

// stateResponse build string, represent history state
// PC=003a SP=ff46 AF=005c BC=174b HL=107f DE=0006 IX=ffff IY=5c3a AF'=0044 BC'=ffff HL'=ffff DE'=5cb9 I=3f R=78
// IM0 IFF-- (PC)=2a785c23 (SP)=107f MMU=00000000000000000000000000000000
func (p *ZRCP) stateResponse(state *z80.CPU) string {
	resp := fmt.Sprintf(getStateResponse,
		state.PC,
		state.SP,
		toW(state.A, state.Flags.GetFlags()),
		toW(state.B, state.C),
		toW(state.H, state.L),
		toW(state.D, state.E),
		state.IX,
		state.IY,
		toW(state.AAlt, state.FlagsAlt.GetFlags()),
		toW(state.BAlt, state.CAlt),
		toW(state.HAlt, state.LAlt),
		toW(state.DAlt, state.EAlt),
		state.I,
		state.R,
		iifStr(state.Iff1, state.Iff2),
		p.getNBytes(state.PC, 4),
		p.getNBytes(state.SP, 2),
	)
	log.Trace(resp)
	return resp
}

func (p *ZRCP) handleSetRegister() (string, error) {
	state := p.computer.CPUState()
	params := strings.Split(p.params, "=")
	if len(params) != 2 {
		return "error", errors.New("invalid set register parameter")
	}
	val, e := strconv.Atoi(params[1])
	if e != nil {
		return "error", errors.New("invalid set register value")
	}
	switch params[0] {
	case "AF":
		state.A = uint8(val >> 8)
		state.Flags.SetFlags(uint8(val))
	case "BC":
		state.B = uint8(val >> 8)
		state.C = uint8(val)
	case "DE":
		state.D = uint8(val >> 8)
		state.E = uint8(val)
	case "HL":
		state.H = uint8(val >> 8)
		state.L = uint8(val)
	// ------------------------------
	case "SP":
		state.SP = uint16(val)
	case "PC":
		state.PC = uint16(val)
	case "IX":
		state.IX = uint16(val)
	case "IY":
		state.IY = uint16(val)
	// ------------------------------
	case "AF'":
		state.AAlt = uint8(val >> 8)
		state.FlagsAlt.SetFlags(uint8(val))
	case "BC'":
		state.BAlt = uint8(val >> 8)
		state.CAlt = uint8(val)
	case "DE'":
		state.DAlt = uint8(val >> 8)
		state.EAlt = uint8(val)
	case "HL'":
		state.HAlt = uint8(val >> 8)
		state.LAlt = uint8(val)

	// ------------------------------
	case "A":
		state.A = uint8(val)
	case "F":
		state.Flags.SetFlags(uint8(val))
	case "B":
		state.B = uint8(val)
	case "C":
		state.C = uint8(val)
	case "D":
		state.D = uint8(val)
	case "E":
		state.E = uint8(val)
	case "H":
		state.H = uint8(val)
	case "L":
		state.L = uint8(val)
	// ------------------------------
	case "A'":
		state.AAlt = uint8(val)
	case "F'":
		state.FlagsAlt.SetFlags(uint8(val))
	case "B'":
		state.BAlt = uint8(val)
	case "C'":
		state.CAlt = uint8(val)
	case "D'":
		state.DAlt = uint8(val)
	case "E'":
		state.EAlt = uint8(val)
	case "H'":
		state.HAlt = uint8(val)
	case "L'":
		state.LAlt = uint8(val)
	// ------------------------------
	case "I":
		state.I = uint8(val)
	case "R":
		state.R = uint8(val)
	default:
		log.Errorf("Unsupported set register parameter: %s", p.params)
	}
	p.computer.SetCPUState(state)
	return registersResponse(p.computer.CPUState()), nil
}

func (p *ZRCP) handleReadMemory() (string, error) {
	params := strings.Split(p.params, " ")
	if len(params) != 2 {
		return "", errors.New("error, invalid read memory parameter")
	}
	offset, e := strconv.Atoi(params[0])
	if e != nil {
		return "", errors.New("error, invalid number for address")
	}

	size, e := strconv.Atoi(params[1])
	if e != nil {
		return "", errors.New("error, invalid number for size")
	}
	resp := ""
	for i := 0; i < size; i++ {
		resp += fmt.Sprintf("%02X", p.computer.MemRead(uint16(offset)+uint16(i)))
	}
	log.Tracef("read memory 0x%04X, 0x%04X: %s", offset, size, resp)
	return resp, nil
}

func (p *ZRCP) getExtendedStack() (string, error) {
	params := strings.Split(p.params, " ")
	if len(params) < 2 {
		return "", errors.New("error, will be 2 or 3 params")
	}
	size, err := strconv.Atoi(params[1])
	if err != nil || size < 0 || size > 65636 {
		return "", errors.New("error, invalid size parameter")
	}

	sp := p.computer.CPUState().SP
	if len(params) == 3 {
		psp, err := strconv.ParseUint(params[2], 10, 16)
		if err != nil {
			return "", errors.New("error, illegal number for SP")
		}
		sp = uint16(psp)
	}

	resp := ""
	spEnd := sp - uint16(size*2)
	es, err := p.computer.ExtendedStack()
	if err == nil {
		for i := sp; i > spEnd; i -= 2 {
			resp += fmt.Sprintf("%04XH %s\n", p.computer.MemRead(i), PushValueTypeName[es[i]])
		}
	}
	log.Tracef("extended-stack get: %s", resp)
	return resp, err
}

func (p *ZRCP) handleSetBreakpoint() (string, error) {
	// 1 PC=0010Bh
	params := strings.Split(p.params, " ")
	if len(params) < 2 {
		return "", errors.New("error, invalid parameters")
	}
	no, e := strconv.ParseUint(params[0], 0, 16)
	if e != nil || no > breakpoint.MaxBreakpoints || no < 1 {
		return "", errors.New("error, invalid breakpoint number")
	}

	e = p.debugger.SetBreakpoint(uint16(no), p.params[len(params[0]):], 1)
	if e != nil {
		return "", errors.New("error, " + e.Error())
	}
	return "", nil
}

func typToString(typ uint8) string {
	switch typ {
	case 0:
		return "D"
	case 1:
		return "R"
	case 2:
		return "W"
	case 3:
		return "R/W"
	default:
		return "x"
	}
}

func (p *ZRCP) BreakpointHit(number uint16, typ byte) {
	if p.writer != nil {
		pc := p.computer.CPUState().PC
		res := p.disassembler.Disassm(pc)
		msg := ""
		if typ == 0 {
			msg = p.debugger.BPExpression(number)
		} else {
			msg = fmt.Sprintf("MEM[%04X] %s", number, typToString(typ))
		}
		rep := fmt.Sprintf("Breakpoint fired: %s\n%s", msg, res)
		log.Debug(rep)
		p.writeResponseMessage(rep)
	}
}

func (p *ZRCP) setBreakpointPassCount(param string) {
	params := strings.Split(param, " ")
	if len(params) != 2 {
		log.Errorf("Set breakpoint passCount failed, expected 2 params, got %d", len(params))
	}
	bpNo, err := strconv.Atoi(params[0])
	if err != nil || bpNo < 0 || bpNo > breakpoint.MaxBreakpoints {
		log.Errorf("Invalid BP no.: %v", err)
	}
	passCount, err := strconv.Atoi(params[1])
	if err != nil || passCount < 0 || passCount > 65535 {
		log.Errorf("Invalid BP passCount: %v", err)
	}
	p.debugger.SetBreakpointPassCount(uint16(bpNo), uint16(passCount))
}

func (p *ZRCP) disassemble(param string) string {
	addr, e := strconv.ParseUint(param, 0, 16)
	if e != nil {
		log.Errorf("Invalid disassemble address: %s", param)
	}
	res := p.disassembler.Disassm(uint16(addr))
	log.Debug(res)
	return res
}

func (p *ZRCP) handleSnapshotSave() (string, error) {
	e := p.computer.SaveSnapshot(strings.TrimSpace(p.params))
	if e != nil {
		return "", errors.New("Error saving snapshot: " + e.Error())
	}
	return "", nil
}

func (p *ZRCP) handleSnapshotLoad() (string, error) {
	e := p.computer.LoadSnapshot(strings.TrimSpace(p.params))
	if e != nil {
		return "", errors.New("Error loading snapshot: " + e.Error())
	}
	return "", nil
}

func (p *ZRCP) handleGetTStatesPartial() (string, error) {
	return strconv.FormatUint(p.computer.TStatesPartial(), 10), nil
}

func (p *ZRCP) handleResetTStatesPartial() (string, error) {
	p.computer.ResetTStatesPartial()
	return "", nil
}

func (p *ZRCP) handleEmptyHandler() (string, error) {
	return "", nil
}

func (p *ZRCP) handleAbout() (string, error) {
	return aboutResponse, nil
}

func (p *ZRCP) handleGetVersion() (string, error) {
	return getVersionResponse, nil
}

func (p *ZRCP) handleGetRegisters() (string, error) {
	return registersResponse(p.computer.CPUState()), nil
}

func (p *ZRCP) handleHardResetCPU() (string, error) {
	p.computer.Reset()
	return "", nil
}

func (p *ZRCP) handleEnterCPUStep() (string, error) {
	p.debugger.SetStepMode(true)
	return "", nil
}

func (p *ZRCP) handleExitCPUStep() (string, error) {
	p.debugger.SetStepMode(false)
	return "", nil
}

func (p *ZRCP) handleGetCurrentMachine() (string, error) {
	return getMachineResponse, nil
}

func (p *ZRCP) handleClearMemBreakpoints() (string, error) {
	p.debugger.ClearMemBreakpoints()
	return "", nil
}

func (p *ZRCP) handleSetMemBreakpoint() (string, error) {
	resp := p.SetMemBreakpoint(p.params)
	var err error
	if len(resp) != 0 {
		err = errors.New(resp)
	}
	return "", err
}

func (p *ZRCP) handleEnableBreakpoints() (string, error) {
	p.debugger.SetBreakpointsEnabled(true)
	return "", nil
}

func (p *ZRCP) handleDisableBreakpoints() (string, error) {
	p.debugger.SetBreakpointsEnabled(false)
	return "", nil
}

func (p *ZRCP) setBreakpointState(enable bool) string {
	no, e := strconv.Atoi(p.params)
	if e != nil {
		return fmt.Sprintf("Invalid breakpoint parameter: %s", p.params)
	}
	if enable && !p.debugger.BreakpointsEnabled() {
		return "Error. You must enable breakpoints first"
	}
	p.debugger.SetBreakpointEnabled(uint16(no), enable)
	return ""
}

func (p *ZRCP) handleEnableBreakpoint() (string, error) {
	resp := p.setBreakpointState(true)
	var err error
	if len(resp) != 0 {
		err = errors.New(resp)
	}
	return "", err
}

func (p *ZRCP) handleDisableBreakpoint() (string, error) {
	resp := p.setBreakpointState(false)
	var err error
	if len(resp) != 0 {
		err = errors.New(resp)
	}
	return "", err
}

func (p *ZRCP) handleGetCPUFrequency() (string, error) {
	return strconv.Itoa(int(p.computer.CPUFrequency())), nil
}

func (p *ZRCP) handleSetBreakpointPassCount() (string, error) {
	p.setBreakpointPassCount(p.params)
	return "", nil
}

func (p *ZRCP) handleExtendedStack() (string, error) {
	params := strings.Split(p.params, " ")
	if len(params) < 1 {
		return "", errors.New("error, not enough params")
	}
	cmd := params[0]
	if cmd == "get" {
		return p.getExtendedStack()
	} else if cmd == "enabled" {
		if len(params) < 2 {
			return "", errors.New("error, expected yes|no")
		}
		p.computer.SetExtendedStack(params[1] == "yes")
	} else {
		return "", errors.New("error, unknown sub-command: " + cmd)
	}
	return "", nil
}

// handleCPUCodeCoverage Handle commands:
// cpu-code-coverage enabled yes
// cpu-code-coverage enabled no
// cpu-code-coverage clear
func (p *ZRCP) handleCPUCodeCoverage() (string, error) {
	command := strings.Split(p.params, " ")
	if len(command) < 1 {
		return "", errors.New("error, not enough arguments")
	}
	cmd := command[0]
	resp := ""
	switch cmd {
	case "enabled":
		if len(command) < 2 {
			return "", errors.New("error, not arguments for enabled [yas|no]")
		}
		p.computer.SetCodeCoverage(command[1] == "yes")
	case "clear":
		p.computer.ClearCodeCoverage()
	case "get":
		for addr, _ := range p.computer.CodeCoverage() {
			resp += fmt.Sprintf("%04X ", addr)
		}
	}
	return resp, nil
}
