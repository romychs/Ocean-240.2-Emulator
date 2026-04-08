package zrcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/breakpoint"
	"okemu/okean240"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/romychs/z80go"
	"github.com/romychs/z80go/dis"
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
	params       []string
}

type CommandHandler struct {
	fn   func(zrcp *ZRCP) (string, error)
	desc string
}

/*
hard-reset-cpu                Hard resets the machine
help                          Shows help screen or command help
hexdump                       Dumps memory at address, showing hex and ascii
*/

var commandHandlers = map[string]CommandHandler{
	"about":                   {(*ZRCP).handleAbout, "Shows about message"},
	"clear-membreakpoints":    {(*ZRCP).handleClearMemBreakpoints, "Clear all memory breakpoints"},
	"close-all-menus":         {(*ZRCP).handleEmptyHandler, "Close all visible dialogs"},
	"cpu-code-coverage":       {(*ZRCP).handleCPUCodeCoverage, "Sets cpu code coverage parameters"},
	"cpu-history":             {(*ZRCP).handleCPUHistory, "Runs cpu history actions"},
	"cpu-step":                {(*ZRCP).handleCpuStep, "Run single opcode cpu step"},
	"disable-breakpoint":      {(*ZRCP).handleDisableBreakpoint, "Disable specific breakpoint"},
	"disable-breakpoints":     {(*ZRCP).handleDisableBreakpoints, "Disable all breakpoints"},
	"disassemble":             {(*ZRCP).handleDisassemble, "Disassemble at address"},
	"enable-breakpoint":       {(*ZRCP).handleEnableBreakpoint, "Enable specific breakpoint"},
	"enable-breakpoints":      {(*ZRCP).handleEnableBreakpoints, "Enable breakpoints"},
	"enter-cpu-step":          {(*ZRCP).handleEnterCPUStep, "Enter cpu step to step mode"},
	"evaluate":                {(*ZRCP).handleEvaluate, "Evaluate expression"},
	"exit-cpu-step":           {(*ZRCP).handleExitCPUStep, "Exit cpu step to step mode"},
	"extended-stack":          {(*ZRCP).handleExtendedStack, "Sets extended stack parameters, which allows you to see what kind of values are in the stack"},
	"get-cpu-frequency":       {(*ZRCP).handleGetCPUFrequency, "Get cpu frequency in HZ"},
	"get-current-machine":     {(*ZRCP).handleGetCurrentMachine, "Returns current machine name"},
	"get-machines":            {(*ZRCP).handleGetMachines, "Returns list of emulated machines"},
	"get-membreakpoints":      {(*ZRCP).handleGetMemBreakpoints, "Get memory breakpoints list"},
	"get-memory-pages":        {(*ZRCP).handleGetMemoryPages, "Returns current state of memory pages"},
	"get-os":                  {(*ZRCP).handleGetOs, "Shows emulator operating system"},
	"get-registers":           {(*ZRCP).handleGetRegisters, "Get CPU registers"},
	"get-tstates":             {(*ZRCP).handleGetTStates, "Get the t-states counter"},
	"get-tstates-partial":     {(*ZRCP).handleGetTStatesPartial, "Get the t-states partial counter"},
	"get-version":             {(*ZRCP).handleGetVersion, "Shows emulator version"},
	"hard-reset-cpu":          {(*ZRCP).handleHardResetCPU, "Hard resets the machine"},
	"help":                    {(*ZRCP).handleEmptyHandler, "Shows help screen or command help"},
	"hexdump":                 {(*ZRCP).handleHexDump, "Dumps memory at address, showing hex and ascii"},
	"load-binary":             {(*ZRCP).handleLoadBinary, "Load binary file \"file\" at address \"addr\" with length \"len\", on the current memory zone"},
	"quit":                    {(*ZRCP).handleEmptyHandler, "Closes connection"},
	"reset-cpu":               {(*ZRCP).handleResetCPU, "Resets CPU"},
	"read-memory":             {(*ZRCP).handleReadMemory, "Dumps memory at address"},
	"reset-tstates-partial":   {(*ZRCP).handleResetTStatesPartial, "Resets the t-states partial counter"},
	"run":                     {(*ZRCP).handleRun, "Run cpu when on cpu step mode"},
	"save-binary":             {(*ZRCP).handleSaveBinary, "Save binary file \"file\" from address \"addr\" with length \"len\", from the current memory zone"},
	"set-breakpoint":          {(*ZRCP).handleSetBreakpoint, "Sets a breakpoint at desired index entry with condition"},
	"set-breakpointaction":    {(*ZRCP).handleEmptyHandler, "Sets a breakpoint action at desired index entry"},
	"set-breakpointpasscount": {(*ZRCP).handleSetBreakpointPassCount, "Set pass count for breakpoint"},
	"set-debug-settings":      {(*ZRCP).handleEmptyHandler, "Set debug settings on remote command protocol"},
	"set-membreakpoint":       {(*ZRCP).handleSetMemBreakpoint, "Sets a memory breakpoint starting at desired address entry for type"},
	"set-machine":             {(*ZRCP).handleEmptyHandler, "Set machine"},
	"set-register":            {(*ZRCP).handleSetRegister, "Changes register value"},
	"snapshot-load":           {(*ZRCP).handleSnapshotLoad, "Loads a snapshot"},
	"snapshot-save":           {(*ZRCP).handleSnapshotSave, "Saves a snapshot"},
	"write-memory":            {(*ZRCP).handleWriteMemory, "Writes a sequence of bytes starting at desired address on memory"},
	"write-port":              {(*ZRCP).handleWritePort, "Writes value at port"},
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

// handleConnection Receive and handle commands
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

func (p *ZRCP) writeWelcomeMessage() bool {
	return p.writeResponseMessage(welcomeMessage)
}

// writeResponseMessage send response and prompt message to client
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

// writeMessage send message back to clien
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

// handleCommand route client command
func (p *ZRCP) handleCommand(str string) bool {
	str = strings.TrimSpace(str)
	if str == "" {
		return false
	}

	log.Debugf("Command: '%s'", str)

	pos := strings.Index(str, " ")
	cmd := str

	p.params = []string{}
	if pos > 1 {
		cmd = str[:pos]
		if len(str) >= pos+1 {
			p.params = strings.Split(strings.TrimSpace(str[pos+1:]), " ")
		}
	}

	var err error
	var resp string

	if cmd == "quit" {
		return false
	} else if cmd == "help" {
		resp, _ := p.handleHelp()
		p.writeResponseMessage(resp)
		return true
	}

	handler, ok := commandHandlers[cmd]
	if ok {
		resp, err = handler.fn(p)
		if err != nil {
			//log.Errorf("%v", err)
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

func (p *ZRCP) handleCpuStep() (string, error) {
	p.debugger.SetDoStep(true) // computer.Do()
	text := p.disassembler.Disassm(p.computer.CPUState().PC)
	return p.registersResponse(p.computer.CPUState()) + " TSTATES: " + strconv.Itoa(int(p.computer.TStatesPartial())) + "\n" + text, nil
}

func (p *ZRCP) handleRun() (string, error) {
	p.writeMessage(runUntilBPMessage)
	p.debugger.SetRunMode(true)
	return "-", nil
}

func (p *ZRCP) handleSetMemBreakpoint() (string, error) {
	if len(p.params) < 1 {
		return "", errors.New("error, not enough parameters")
	}
	address, err := parseUint16(p.params[0])
	if err != nil {
		return "", errors.New("error, illegal address")
	}
	typ := uint16(3)
	if len(p.params) > 1 {
		typ, err = parseUint16(p.params[1])
		if err != nil || typ > 3 {
			return "", errors.New("error, illegal access type")
		}
	}

	size := uint16(1)
	if len(p.params) > 2 {
		size, err = parseUint16(p.params[2])
		if err != nil {
			return "", errors.New("error, illegal memory size")
		}
	}
	if p.debugger != nil {
		p.debugger.SetMemBreakpoint(address, byte(typ), size)
	}
	return "", nil
}

func (p *ZRCP) handleCPUHistory() (string, error) {
	if len(p.params) < 1 {
		return "", errors.New("error, no parameters")
	}

	cmd := p.params[0]
	nspe := errors.New("error, no second parameter")

	switch cmd {

	case "enabled":
		if len(p.params) < 2 {
			return "", nspe
		}
		p.debugger.SetCpuHistoryEnabled(p.params[1] == "yes")

	case "clear":
		p.debugger.CpuHistoryClear()

	case "started":
		if len(p.params) < 2 {
			return "", nspe
		}
		p.debugger.SetCpuHistoryStarted(p.params[1] == "yes")
	case "set-max-size":
		if len(p.params) != 2 {
			return "", nspe
		}
		size, err := parseUint64(p.params[1])
		if err != nil {
			return "", errors.New("error, illegal number")
		}
		p.debugger.SetCpuHistoryMaxSize(int(size))
	case "get":
		if len(p.params) != 2 {
			return "", nspe
		}
		index, err := parseUint64(p.params[1])
		if err != nil {
			return "", errors.New("error, illegal number")
		}
		history := p.debugger.CpuHistory(int(index))
		if history != nil {
			return p.stateResponse(history), nil
		}
		return "", errors.New("ERROR: index out of range")
	case "get-size":
		return strconv.Itoa(p.debugger.CpuHistorySize()), nil
	case "ignrephalt", "ignrepldxr":
		// ignore
	default:
		return "", errors.New("error: unknown history command: " + cmd)
	}

	return "", nil
}

func (p *ZRCP) handleLoadBinary() (string, error) {
	loadError := errors.New(respErrorLoading)
	if len(p.params) < 2 {
		return "", loadError
	}
	fn := strings.Trim(p.params[0], " \"\t")
	offset, e := parseUint16(p.params[1])
	length := 0
	if e != nil || offset < 0 || offset > 65535 || len(fn) == 0 {
		return "", loadError
	}
	if len(p.params) > 2 {
		l, e := parseUint64(p.params[2])
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
		p.computer.MemWrite(addr+offset, data[addr])
	}
	return "", nil
}

// registersResponse Build string
// PC=%4x SP=%4x AF=%4x BC=%4x HL=%4x DE=%4x IX=%4x IY=%4x AF'=%4x BC'=%4x HL'=%4x DE'=%4x I=%2x
// R=%2x  F=%s F'=%s MEMPTR=%4x IM0 IFF-- VPS: 0 MMU=00000000000000000000000000000000
func (p *ZRCP) registersResponse(state *z80go.CPU) string {
	resp := fmt.Sprintf(getRegistersResponse,
		state.PC,
		state.SP,
		toW(state.A, state.Flags.AsByte()),
		toW(state.B, state.C),
		toW(state.H, state.L),
		toW(state.D, state.E),
		state.IX,
		state.IY,
		toW(state.AAlt, state.FlagsAlt.AsByte()),
		toW(state.BAlt, state.CAlt),
		toW(state.HAlt, state.LAlt),
		toW(state.DAlt, state.EAlt),
		state.I,
		state.R,
		state.Flags.String(),
		state.FlagsAlt.String(),
		state.MemPtr,
		iifStr(state.Iff1, state.Iff2),
		p.getMMU(),
	)
	log.Debug(resp)
	return resp
}

// getNBytes  return hex string of n bytes from memory starts at addr
func (p *ZRCP) getNBytes(addr uint16, n int) string {
	var res strings.Builder
	for i := 0; i < n; i++ {
		res.WriteString(fmt.Sprintf("%02X", p.computer.MemRead(addr)))
		addr++
	}
	return res.String()
}

// stateResponse build string, represent history state
// PC=003a SP=ff46 AF=005c BC=174b HL=107f DE=0006 IX=ffff IY=5c3a AF'=0044 BC'=ffff HL'=ffff DE'=5cb9 I=3f R=78
// IM0 IFF-- (PC)=2a785c23 (SP)=107f MMU=00000000000000000000000000000000
func (p *ZRCP) stateResponse(state *z80go.CPU) string {
	resp := fmt.Sprintf(getStateResponse,
		state.PC,
		state.SP,
		toW(state.A, state.Flags.AsByte()),
		toW(state.B, state.C),
		toW(state.H, state.L),
		toW(state.D, state.E),
		state.IX,
		state.IY,
		toW(state.AAlt, state.FlagsAlt.AsByte()),
		toW(state.BAlt, state.CAlt),
		toW(state.HAlt, state.LAlt),
		toW(state.DAlt, state.EAlt),
		state.I,
		state.R,
		iifStr(state.Iff1, state.Iff2),
		p.getNBytes(state.PC, 4),
		p.getNBytes(state.SP, 2),
		p.getMMU(),
	)
	log.Trace(resp)
	return resp
}

func (p *ZRCP) handleSetRegister() (string, error) {
	state := p.computer.CPUState()
	if len(p.params) != 1 {
		return "", errors.New("error, expected REG=val")
	}
	regPar := strings.Split(p.params[0], "=")
	if len(regPar) != 2 {
		return "error", errors.New("error, illegal set register parameter: '" + regPar[0] + "'")
	}
	val, e := parseUint16(regPar[1])
	if e != nil {
		return "error", errors.New("invalid register value: '" + regPar[1] + "'")
	}
	switch regPar[0] {
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
		state.SP = val
	case "PC":
		state.PC = val
	case "IX":
		state.IX = val
	case "IY":
		state.IY = val
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
		log.Errorf("Unsupported set register parameter: %v", p.params)
	}
	p.computer.SetCPUState(state)
	return p.registersResponse(p.computer.CPUState()), nil
}

func (p *ZRCP) handleReadMemory() (string, error) {
	addr, size, err := p.getAddrValue64()
	if err != nil {
		return "", err
	}
	if size > 65536 {
		return "", errors.New("error, too many bytes")
	}
	return p.getNBytes(addr, int(size)), nil
}

func (p *ZRCP) getExtendedStack() (string, error) {
	if len(p.params) < 2 {
		return "", errors.New("error, will be 2 or 3 params")
	}

	size, err := parseUint16(p.params[1])
	if err != nil {
		return "", errors.New("error, invalid size parameter")
	}

	sp := p.computer.CPUState().SP
	if len(p.params) == 3 {
		var err error
		sp, err = parseUint16(p.params[2])
		if err != nil {
			return "", errors.New("error, illegal number for SP")
		}
	}

	var resp strings.Builder
	spEnd := sp - size*2
	es, err := p.computer.ExtendedStack()

	if err == nil {
		for i := sp; i > spEnd; i -= 2 {
			pvt, ok := es[i]
			if !ok {
				pvt = z80go.PushValueTypeDefault
			}
			resp.WriteString(fmt.Sprintf("%04XH %s\n", p.computer.MemRead(i), PushValueTypeName[pvt]))
		}
	}
	log.Tracef("extended-stack get: %s", resp)
	return resp.String(), err
}

func (p *ZRCP) handleSetBreakpoint() (string, error) {
	if len(p.params) < 2 {
		return "", errors.New("error, invalid parameters")
	}

	no, e := parseUint16(p.params[0])
	if e != nil || no > breakpoint.MaxBreakpoints || no < 1 {
		return "", errors.New("error, invalid breakpoint number")
	}
	exp := strings.Join(p.params[1:], " ")
	e = p.debugger.SetBreakpoint(no, exp, 1)
	if e != nil {
		return "", errors.New("error, " + e.Error())
	}
	return "", nil
}

// BreakpointHit handle breakpoint hit, called from program cycle
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

func (p *ZRCP) handleSetBreakpointPassCount() (string, error) {
	bpNo, passCount, err := p.getAddrValue16()
	if err != nil {
		return "", err
	}
	p.debugger.SetBreakpointPassCount(bpNo, passCount)
	return "", nil
}

func (p *ZRCP) handleDisassemble() (string, error) {
	var addr uint16
	var size uint64
	if len(p.params) == 0 {
		addr = p.computer.CPUState().PC
	} else {
		var e error
		addr, e = parseUint16(p.params[0])
		if e != nil {
			return "", fmt.Errorf("error, illegal address: %s", p.params[0])
		}
		if len(p.params) == 2 {
			size, e = parseUint64(p.params[1])
			if e != nil {
				return "", fmt.Errorf("error, illegal size: %s", p.params[1])
			}
		} else {
			size = 1
		}
	}
	res := p.disassembler.Disassm(addr)
	log.Tracef("DISASSM[0x%04X, %d]: %s", addr, size, res)
	return res, nil
}

func (p *ZRCP) handleSnapshotSave() (string, error) {
	if len(p.params) < 1 {
		return "", errors.New("error, no parameter")
	}
	e := p.computer.SaveSnapshot(strings.TrimSpace(p.params[0]))
	if e != nil {
		return "", errors.New("error: " + e.Error())
	}
	return "", nil
}

func (p *ZRCP) handleSnapshotLoad() (string, error) {
	if len(p.params) < 1 {
		return "", errors.New("error, no parameter")
	}
	e := p.computer.LoadSnapshot(strings.TrimSpace(p.params[0]))
	if e != nil {
		return "", errors.New("error: " + e.Error())
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
	return p.registersResponse(p.computer.CPUState()), nil
}

func (p *ZRCP) handleHardResetCPU() (string, error) {
	p.computer.HardReset()
	return "", nil
}

func (p *ZRCP) handleResetCPU() (string, error) {
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

func (p *ZRCP) handleEnableBreakpoints() (string, error) {
	p.debugger.SetBreakpointsEnabled(true)
	return "", nil
}

func (p *ZRCP) handleDisableBreakpoints() (string, error) {
	p.debugger.SetBreakpointsEnabled(false)
	return "", nil
}

func (p *ZRCP) setBreakpointState(enable bool) string {
	if len(p.params) == 0 {
		return "error, no bp number"
	}
	no, e := parseUint16(p.params[0])
	if e != nil {
		return "error, illegal bp number"
	}
	if enable && !p.debugger.BreakpointsEnabled() {
		return "Error. You must enable breakpoints first"
	}
	p.debugger.SetBreakpointEnabled(no, enable)
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

func (p *ZRCP) handleExtendedStack() (string, error) {
	if len(p.params) < 1 {
		return "", errors.New("error, not enough params")
	}
	cmd := p.params[0]
	if cmd == "get" {
		return p.getExtendedStack()
	} else if cmd == "enabled" {
		if len(p.params) < 2 {
			return "", errors.New("error, expected yes|no")
		}
		p.computer.SetExtendedStack(p.params[1] == "yes")
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
	if len(p.params) < 1 {
		return "", errors.New("error, not enough params")
	}

	resp := ""

	switch p.params[0] {
	case "enabled":
		if len(p.params) < 2 {
			return "", errors.New("error, not arguments for enabled [yas|no]")
		}
		p.computer.SetCodeCoverage(p.params[1] == "yes")
	case "clear":
		p.computer.ClearCodeCoverage()
	case "get":
		for addr, _ := range p.computer.CodeCoverage() {
			resp += fmt.Sprintf("%04X ", addr)
		}
	}
	return resp, nil
}

func (p *ZRCP) getAddrValue64() (uint16, uint64, error) {
	if len(p.params) != 2 {
		return 0, 0, errors.New("error, not enough params")
	}
	addr, e := parseUint16(p.params[0])
	if e != nil {
		return 0, 0, errors.New("error, invalid first number: " + e.Error())
	}

	val, e := parseUint64(p.params[1])
	if e != nil {
		return 0, 0, errors.New("error, invalid second number: " + e.Error())
	}
	return addr, val, nil
}

func (p *ZRCP) getAddrValue16() (uint16, uint16, error) {
	addr, val, err := p.getAddrValue64()
	return addr, uint16(val), err
}

func (p *ZRCP) getAddrValue8() (uint16, uint8, error) {
	addr, val, err := p.getAddrValue64()
	return addr, uint8(val), err
}

func (p *ZRCP) handleWriteMemory() (string, error) {
	addr, val, e := p.getAddrValue8()
	if e != nil {
		return "", e
	}
	p.computer.MemWrite(addr, val)
	log.Tracef("0x%02X=>MEM[0x%04X]", val, addr)
	return "", nil
}

func (p *ZRCP) handleWritePort() (string, error) {
	addr, val, e := p.getAddrValue8()
	if e != nil {
		return "", e
	}
	p.computer.IOWrite(addr, val)
	log.Tracef("0x%02X=>IO[0x%04X]", val, addr)
	return "", nil
}

func (p *ZRCP) handleEvaluate() (string, error) {
	if len(p.params) == 0 {
		return "0", nil
	}
	return p.computer.Evaluate(strings.Join(p.params, " "))
}

func (p *ZRCP) getMMU() string {
	var res strings.Builder
	for _, id := range p.computer.MemoryPages() {
		res.WriteString(fmt.Sprintf("%04X", id))
	}
	return res.String()
}

func (p *ZRCP) handleGetMemoryPages() (string, error) {
	var res strings.Builder
	for _, id := range p.computer.MemoryPages() {
		if id < 0xf0 {
			res.WriteString(fmt.Sprintf("RA%d ", id))
		} else {
			res.WriteString(fmt.Sprintf("RO%d ", id-0xf0))
		}
	}
	return res.String(), nil
}

func (p *ZRCP) handleGetOs() (string, error) {
	return runtime.GOOS, nil
}

func (p *ZRCP) handleGetTStates() (string, error) {
	return strconv.FormatUint(p.computer.Cycles(), 10), nil
}

func (p *ZRCP) handleSaveBinary() (string, error) {
	if len(p.params) != 3 {
		return "", errors.New("error, need 3 parameters")
	}
	fn := strings.Trim(p.params[0], " \t\"")
	addr, e := parseUint16(p.params[1])
	if e != nil {
		return "", errors.New("error, invalid address")
	}
	size, e := parseUint16(p.params[2])
	if e != nil {
		return "", errors.New("error, invalid size")
	}
	var block []byte
	for c := uint16(0); c < size; c++ {
		block = append(block, p.computer.MemRead(addr))
		addr++
	}
	err := os.WriteFile(fn, block, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return "", nil
}

func (p *ZRCP) handleGetMachines() (string, error) {
	return "OK240.2\t" + getMachineResponse, nil
}

func (p *ZRCP) handleGetMemBreakpoints() (string, error) {
	var res strings.Builder
	for _, bp := range p.debugger.GetMemBreakpoints() {
		res.WriteString(bp.String())
		res.WriteByte(0x0a)
	}
	return res.String(), nil
}

func (p *ZRCP) handleHelp() (string, error) {
	var res strings.Builder
	res.WriteString("Available commands:\n")
	commands := slices.Collect(maps.Keys(commandHandlers))
	slices.Sort(commands)
	for _, cmd := range commands {
		res.WriteString(fmt.Sprintf("%-*s%s\n", 24, cmd, commandHandlers[cmd].desc))
	}
	res.WriteString("\nTotal commands: " + strconv.Itoa(len(commandHandlers)) + "\n")
	return res.String(), nil
}

func (p *ZRCP) handleHexDump() (string, error) {
	addr, size, err := p.getAddrValue64()
	if err != nil {
		return "", err
	}
	ctr := 0
	fakeSize := size / 16 * 16
	if fakeSize != size {
		fakeSize += 16
	}
	var resB strings.Builder
	var resA strings.Builder
	for c := uint64(0); c <= fakeSize; c++ {
		if ctr%16 == 0 {
			if resB.Len() > 0 {
				resB.WriteString(resA.String())
				resB.WriteString("|\n")
			}
			if c == fakeSize {
				break
			}
			resA.Reset()
			resA.WriteString(" |")
			resB.WriteString(fmt.Sprintf("  %04XH ", addr))
		}
		if c < size {
			b := p.computer.MemRead(addr)
			resB.WriteString(fmt.Sprintf("%02X ", b))
			if b >= 32 && b < 127 {
				resA.WriteString(string(b))
			} else {
				resA.WriteByte(0x2e) // .
			}
		} else {
			resB.WriteString("   ")
			resA.WriteByte(0x20)
		}
		addr++
		ctr++
	}
	return resB.String(), nil
}
