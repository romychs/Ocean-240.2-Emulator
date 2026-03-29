package listener

import (
	"bufio"
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

// Receive messages, split to strings and parse
func handleConnection(c net.Conn) {
	reader := bufio.NewReader(c)
	writer := bufio.NewWriter(c)
	if !writeWelcomeMessage(writer) {
		return
	}
	activeWriter = writer
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Errorf("TCP error: %v", err)
				debugger.SetStepMode(false)
				return
			}
		}
		if !HandleCommand(str, writer) {
			log.Debug("Closing connection")
			writeResponseMessage(writer, quitResponse)
			break
		}
		//byteBuffer.WriteByte(b)
	}
	debugger.SetStepMode(false)
	activeWriter = nil
	err := c.Close()
	if err != nil {
		log.Warnf("Can not close socket: %v", err)
	}

}

var activeWriter *bufio.Writer = nil

func writeWelcomeMessage(writer *bufio.Writer) bool {
	return writeResponseMessage(writer, welcomeMessage)
}

//command@cpu-step

func writeResponseMessage(writer *bufio.Writer, message string) bool {
	prompt := emptyResponse
	if debugger.StepMode() {
		prompt = inCpuStepResponse
	}

	_, err := writer.WriteString(message + prompt)
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	err = writer.Flush()
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	return true
}

func writeMessage(writer *bufio.Writer, message string) bool {
	_, err := writer.WriteString(message)
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	err = writer.Flush()
	if err != nil {
		log.Errorf("TCP error: %v", err)
		return false
	}
	return true
}

// var
var debugger *debug.Debugger
var disassembler *dis.Disassembler
var computer *okean240.ComputerType

// SetupTcpHandler Setup TCP listener, handle connections
func SetupTcpHandler(config *config.OkEmuConfig, debug *debug.Debugger, disasm *dis.Disassembler, comp *okean240.ComputerType) {
	port := config.Debugger.Host + ":" + strconv.Itoa(config.Debugger.Port)
	debugger = debug
	disassembler = disasm
	computer = comp

	log.Infof("Ready for debugger connections on %s", port)

	l, err := net.Listen("tcp4", port)
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

	for {
		c, err := l.Accept()
		if err != nil {
			log.Errorf("Accept connection: %v", err)
			return
		}
		go handleConnection(c)
	}

}

// HandleCommand HandleLogLine Parse log line(s) and send it to redis
func HandleCommand(str string, writer *bufio.Writer) bool {
	quit := false
	str = strings.TrimSpace(str)
	if str == "" {
		return false
	}
	log.Debugf("Command: '%s'", str)

	pos := strings.Index(str, " ")
	cmd := str
	params := ""

	if pos > 1 {
		cmd = str[:pos]
		params = strings.TrimSpace(str[pos+1:])
	}

	switch cmd {
	case "cpu-step":
		debugger.SetDoStep(true) // computer.Do()
		text := disassembler.Disassm(computer.CPUState().PC)
		writeResponseMessage(writer, registersResponse(computer.CPUState())+" TSTATES: "+strconv.Itoa(int(computer.TStatesPartial()))+"\n"+text)
	case "run":
		writeMessage(writer, runUntilBPMessage)
		debugger.SetRunMode(true)
	case "disassemble":
		writeResponseMessage(writer, disassemble(params))
	case "get-tstates-partial":
		writeResponseMessage(writer, strconv.FormatUint(computer.TStatesPartial(), 10))
	case "reset-tstates-partial":
		computer.ResetTStatesPartial()
		writeResponseMessage(writer, "")
	case "close-all-menus":
		writeResponseMessage(writer, "")
	case "about":
		writeResponseMessage(writer, aboutResponse)
	case "get-version":
		writeResponseMessage(writer, getVersionResponse)
	case "get-registers":
		writeResponseMessage(writer, registersResponse(computer.CPUState()))
	case "set-register":
		writeResponseMessage(writer, setRegister(params))
	case "hard-reset-cpu":
		computer.Reset()
		writeResponseMessage(writer, "")
	case "enter-cpu-step":
		debugger.SetStepMode(true)
		writeResponseMessage(writer, "")
	case "exit-cpu-step":
		debugger.SetStepMode(false)
		writeResponseMessage(writer, "")
	case "set-debug-settings":
		log.Debugf("Set debug settings to %s", params)
		writeResponseMessage(writer, "")
	case "get-current-machine":
		writeResponseMessage(writer, getMachineResponse)
	case "clear-membreakpoints":
		debugger.ClearMemBreakpoints()
		writeResponseMessage(writer, "")
	case "set-membreakpoint": // addr type size
		writeResponseMessage(writer, SetMemBreakpoint(params))
	case "enable-breakpoints":
		debugger.SetBreakpointsEnabled(true)
		writeResponseMessage(writer, "")
	case "disable-breakpoints":
		debugger.SetBreakpointsEnabled(false)
		writeResponseMessage(writer, "")
	case "enable-breakpoint":
		writeResponseMessage(writer, setBreakpointState(params, true))
	case "disable-breakpoint":
		writeResponseMessage(writer, setBreakpointState(params, false))
	case "get-cpu-frequency":
		writeResponseMessage(writer, strconv.Itoa(int(computer.CPUFrequency())))
	case "set-breakpoint":
		// 1 PC=0010Bh
		writeResponseMessage(writer, setBreakpoint(params))
	case "set-breakpointpasscount":
		setBreakpointPassCount(params)
		writeResponseMessage(writer, "")
	case "cpu-code-coverage":
		//"enabled no"
		writeResponseMessage(writer, "")
	case "cpu-history":
		// "enabled yes"
		// "set-max-size 1000"
		// "clear"
		// "started yes"
		// "ignrephalt yes"
		// "ignrepldxr yes"

		writeResponseMessage(writer, doCpuHistory(params))
	case "extended-stack":
		// "enabled no"
		// "enabled yes"
		if strings.HasPrefix(params, "get") {
			writeResponseMessage(writer, getExtendedStack(params))
		} else {
			writeResponseMessage(writer, "")
		}
	case "load-binary":
		writeResponseMessage(writer, loadBinary(params))
	case "read-memory":
		writeResponseMessage(writer, readMemory(params))
	case "quit":
		quit = true
	case "snapshot-save":
		writeResponseMessage(writer, snapshotSave(params))
	case "snapshot-load":
		writeResponseMessage(writer, snapshotLoad(params))
	case "set-breakpointaction":
		// now do nothing
		writeResponseMessage(writer, "")
	default:
		log.Debugf("Unhandled Command: %s", str)
		writeResponseMessage(writer, "")
	}
	return !quit
}

func convertToUint16(s string) (uint16, error) {
	v := strings.TrimSpace(strings.ToUpper(s))
	base := 0
	if strings.HasSuffix(v, "h") || strings.HasSuffix(v, "H") {
		v = strings.TrimSuffix(v, "H")
		v = strings.TrimSuffix(v, "h")
		base = 16
	}
	a, e := strconv.ParseUint(v, base, 16)
	return uint16(a), e
}

func SetMemBreakpoint(param string) string {
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
	if debugger != nil {
		debugger.SetMemBreakpoint(address, byte(t), s)
	}
	return ""
}

func doCpuHistory(param string) string {
	param = strings.TrimSpace(param)
	params := strings.Split(param, " ")
	if len(params) == 0 {
		return "error"
	}
	cmd := params[0]
	switch cmd {
	case "enabled":
		if len(params) != 2 {
			return "error"
		}
		debugger.SetCpuHistoryEnabled(params[1] == "yes")
	case "clear":
		debugger.CpuHistoryClear()
	case "started":
		if len(params) != 2 {
			return "error"
		}
		debugger.SetCpuHistoryStarted(params[1] == "yes")
	case "set-max-size":
		if len(params) != 2 {
			return "error"
		}
		size, err := strconv.Atoi(params[1])
		if err != nil {
			return "error"
		}
		debugger.SetCpuHistoryMaxSize(size)
	case "get":
		if len(params) != 2 {
			return "error"
		}
		index, err := strconv.Atoi(params[1])
		if err != nil {
			return "error"
		}
		history := debugger.CpuHistory(index)
		if history != nil {
			return stateResponse(history)
		}
		return "ERROR: index out of range"
	}
	return ""
}

func loadBinary(param string) string {
	params := strings.Split(param, " ")
	if len(params) < 2 {
		log.Errorf("Invalid load parameters: %s", param)
		return respErrorLoading
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
		log.Errorf("Invalid load parameters: %s", param)
		return respErrorLoading
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
		log.Errorf("Error reading file: %v", err)
		return respErrorLoading
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
		computer.MemWrite(addr+uint16(offset), data[addr])
	}

	return ""
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
	//state := computer.GetCPUState()
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

func getNBytes(addr uint16, n uint16) string {
	res := ""
	for i := uint16(0); i < n; i++ {
		b := computer.MemRead(addr + i)
		res += fmt.Sprintf("%02X", b)
	}
	return res
}

// stateResponse build string, represent history state
// PC=003a SP=ff46 AF=005c BC=174b HL=107f DE=0006 IX=ffff IY=5c3a AF'=0044 BC'=ffff HL'=ffff DE'=5cb9 I=3f R=78
// IM0 IFF-- (PC)=2a785c23 (SP)=107f MMU=00000000000000000000000000000000
func stateResponse(state *z80.CPU) string {
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
		getNBytes(state.PC, 4),
		getNBytes(state.SP, 2),
	)
	log.Trace(resp)
	return resp
}

func setRegister(param string) string {
	state := computer.CPUState()
	params := strings.Split(param, "=")
	if len(params) != 2 {
		log.Errorf("Invalid set register parameter: %s", param)
		return "error"
	}
	val, e := strconv.Atoi(params[1])
	if e != nil {
		log.Errorf("Invalid set register parameter value: %s", params[1])
		return "error"
	}
	switch params[0] {
	case "SP":
		state.SP = uint16(val)
	case "PC":
		state.PC = uint16(val)
	case "IX":
		state.IX = uint16(val)
	case "IY":
		state.IY = uint16(val)
	case "A":
		state.A = uint8(val)
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

	case "I":
		state.I = uint8(val)
	case "R":
		state.R = uint8(val)
	default:
		log.Errorf("Unsupported set register parameter: %s", param)
	}
	computer.SetCPUState(state)
	return registersResponse(computer.CPUState())
}

func readMemory(param string) string {
	params := strings.Split(param, " ")
	if len(params) != 2 {
		log.Errorf("Invalid read memory parameter: %s", param)
		return "error" //registersResponse(computer.GetCPUState())
	}
	offset, e := strconv.Atoi(params[0])
	if e != nil {
		log.Errorf("Invalid read memory parameter offset: %s", params[0])
	}

	size, e := strconv.Atoi(params[1])
	if e != nil {
		log.Errorf("Invalid read memory parameter size: %s", params[1])
	}
	resp := ""
	for i := 0; i < size; i++ {
		resp += fmt.Sprintf("%02X", computer.MemRead(uint16(offset)+uint16(i)))
	}
	return resp
}

func getExtendedStack(param string) string {
	params := strings.Split(param, " ")
	if len(params) < 2 {
		log.Errorf("Will be 2 or 3 params: %s", param)
		return ""
	}
	size, err := strconv.Atoi(params[1])
	if err != nil || size < 0 || size > 65636 {
		log.Errorf("Invalid size param: %s", param)
	}

	sp := computer.CPUState().SP
	if len(params) == 3 {
		psp, err := strconv.ParseUint(params[2], 10, 16)
		if err != nil {
			log.Errorf("Invalid SP param: %s", params[2])
		} else {
			sp = uint16(psp)
		}
	}

	resp := ""
	spEnd := sp - uint16(size*2)
	for i := sp; i > spEnd; i -= 2 {
		resp += fmt.Sprintf("%04XH default\n", computer.MemRead(i))
	}
	//log.Debugf("Stack[%d,%d]:\n%s", sp, size, resp)
	return resp
}

func setBreakpointState(param string, enable bool) string {
	no, e := strconv.Atoi(param)
	if e != nil {
		log.Errorf("Invalid breakpoint parameter: %s", param)
		return ""
	}
	if enable && !debugger.BreakpointsEnabled() {
		return "Error. You must enable breakpoints first"
	}
	debugger.SetBreakpointEnabled(uint16(no), enable)
	return ""
}

func setBreakpoint(param string) string {
	// 1 PC=0010Bh
	params := strings.Split(param, " ")
	if len(params) < 2 {
		log.Errorf("Invalid set breakpoint parameters: %s", param)
		return "Error, invalid parameters"
	}
	no, e := strconv.ParseUint(params[0], 0, 16)
	if e != nil || no > breakpoint.MaxBreakpoints || no < 1 {
		log.Errorf("Invalid breakpoint number: %s", params[0])
		return "Error, invalid breakpoint number"
	}

	e = debugger.SetBreakpoint(uint16(no), param[len(params[0]):])
	if e != nil {
		return "Error: " + e.Error()
	}
	return ""
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

func BreakpointHit(number uint16, typ byte) {
	if activeWriter != nil {
		pc := computer.CPUState().PC
		res := disassembler.Disassm(pc)
		msg := ""
		if typ == 0 {
			msg = debugger.BPExpression(number)
		} else {
			msg = fmt.Sprintf("MEM[%04X] %s", number, typToString(typ))
		}
		rep := fmt.Sprintf("Breakpoint fired: %s\n%s", msg, res)
		log.Debug(rep)
		writeResponseMessage(activeWriter, rep)
	}
}

func setBreakpointPassCount(param string) {
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
	debugger.SetBreakpointPassCount(uint16(bpNo), uint16(passCount))
}

func disassemble(param string) string {
	addr, e := strconv.ParseUint(param, 0, 16)
	if e != nil {
		log.Errorf("Invalid disassemble address: %s", param)
	}
	res := disassembler.Disassm(uint16(addr))
	log.Debug(res)
	return res
}

func snapshotSave(params string) string {
	e := computer.SaveSnapshot(strings.TrimSpace(params))
	if e != nil {
		return fmt.Sprintf("Error saving snapshot: %s", e)
	}
	return ""
}

func snapshotLoad(params string) string {
	e := computer.LoadSnapshot(strings.TrimSpace(params))
	if e != nil {
		return fmt.Sprintf("Error load snapshot: %s", e)
	}
	return ""
}
