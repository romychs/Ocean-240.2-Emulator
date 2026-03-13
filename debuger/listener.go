package debuger

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"okemu/config"
	"okemu/okean240"
	"os"
	"strings"
	//"okemu/logger"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const welcomeMessage = "Welcome to ZEsarUX remote command protocol (ZRCP)\nWrite help for available commands\n\ncommand> "
const emptyResponse = "\ncommand> "
const aboutResponse = "ZEsarUX remote command protocol"
const getVersionResponse = "12.1"
const getRegistersResponse = "PC=%04x SP=%04x AF=%04x BC=%04x HL=%04x DE=%04x IX=%04x IY=%04x AF'=%04x BC'=%04x HL'=%04x DE'=%04x I=%02x R=%02x  F=%s F'=%s MEMPTR=%04x IM0 IFF%s VPS: 0 MMU=00000000000000000000000000000000"
const inCpuStepResponse = "\ncommand@cpu-step> "
const getMachineResponse = "64K RAM, no ZX\n"
const respErrorLoading = "ERROR loading file"
const quitResponse = "Sayonara baby\n"

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
	activeWriter = nil
	//log.Trace("TCP Connection closed")
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
	if computer.IsStepMode() {
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

var computer *okean240.ComputerType

// SetupTcpHandler Setup TCP listener, handle connections
func SetupTcpHandler(config *config.OkEmuConfig, comp *okean240.ComputerType) {
	port := config.Host + ":" + strconv.Itoa(config.Port)
	computer = comp
	log.Infof("Serve TCP connections on %s", port)

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
		computer.Do()
		writeResponseMessage(writer, "  "+fmt.Sprintf("%04X", computer.GetCPUState().PC))
	case "run":
		_, e := writer.WriteString("Running until a breakpoint, key press or data sent, menu opening or other event\n")
		if e != nil {
			log.Warnf("Error writing to buffer: %v", e)
		}
		e = writer.Flush()
		if e != nil {
			log.Warnf("Error flushing the buffer: %v", e)
		}
		computer.SetRunMode(true)
	case "get-tstates-partial":
		writeResponseMessage(writer, strconv.FormatUint(computer.Cycles(), 10))
	case "close-all-menus":
		writeResponseMessage(writer, "")
	case "about":
		writeResponseMessage(writer, aboutResponse)
	case "get-version":
		writeResponseMessage(writer, getVersionResponse)
	case "get-registers":
		writeResponseMessage(writer, registersResponse())
	case "set-register":
		writeResponseMessage(writer, setRegister(params))
	case "hard-reset-cpu":
		computer.Reset()
		writeResponseMessage(writer, "")
	case "enter-cpu-step":
		computer.SetStepMode(true)
		writeResponseMessage(writer, "")
	case "exit-cpu-step":
		computer.SetStepMode(false)
		writeResponseMessage(writer, "")
	case "set-debug-settings":
		log.Debugf("Set debug settings to %s", params)
		writeResponseMessage(writer, "")
	case "get-current-machine":
		writeResponseMessage(writer, getMachineResponse)
	case "clear-membreakpoints":
		computer.ClearMemBreakpoints()
		writeResponseMessage(writer, "")
	case "enable-breakpoints":
		computer.SetBreakpointsEnabled(true)
		writeResponseMessage(writer, "")
	case "disable-breakpoints":
		computer.SetBreakpointsEnabled(false)
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
		writeResponseMessage(writer, "")
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
	default:
		log.Debugf("Unhandled Command: %s", str)
		writeResponseMessage(writer, "")
	}
	return !quit
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
		length, e = strconv.Atoi(params[1])
		if e != nil {
			length = 0
		}
	}
	data, err := os.ReadFile(fn)
	if err != nil {
		log.Errorf("Error reading file: %v", err)
		return respErrorLoading
	}
	if length != 0 && len(data) < length {
		log.Errorf("File too short. Expected %d bytes, got %d", len(data), length)
		return respErrorLoading
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
func registersResponse() string {
	state := computer.GetCPUState()
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
	log.Debug(resp)
	return resp
}

func setRegister(param string) string {
	state := computer.GetCPUState()
	params := strings.Split(param, "=")
	if len(params) != 2 {
		log.Errorf("Invalid set register parameter: %s", param)
		return registersResponse()
	}
	val, e := strconv.Atoi(params[1])
	if e != nil {
		log.Errorf("Invalid set register parameter value: %s", params[1])
		return registersResponse()
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
	return registersResponse()
}

func readMemory(param string) string {
	params := strings.Split(param, " ")
	if len(params) != 2 {
		log.Errorf("Invalid read memory parameter: %s", param)
		return registersResponse()
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
	log.Tracef("ReadMemory[%d,%d]:\n%s", offset, size, resp)
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

	sp := computer.GetCPUState().SP
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
	log.Debugf("Stack[%d,%d]:\n%s", sp, size, resp)
	return resp
}

func setBreakpointState(param string, enable bool) string {
	no, e := strconv.Atoi(param)
	if e != nil {
		log.Errorf("Invalid breakpoint parameter: %s", param)
		return ""
	}
	if enable && !computer.IsBreakpointsEnabled() {
		return "Error. You must enable breakpoints first"
	}
	computer.SetBreakpointEnabled(uint16(no), enable)
	return ""
}

func setBreakpoint(param string) string {
	// 1 PC=0010Bh
	params := strings.Split(param, " ")
	if len(params) != 2 {
		log.Errorf("Invalid set breakpoint parameters: %s", param)
		return ""
	}
	no, e := strconv.Atoi(params[0])
	if e != nil || no > okean240.MaxBreakpoints || no < 1 {
		log.Errorf("Invalid breakpoint number: %s", params[0])
		return ""
	}

	regv := strings.Split(params[1], "=")
	if len(regv) != 2 {
		log.Errorf("Invalid breakpoint parameter: %s", params[1])
		return ""
	}
	addr, e := strconv.ParseUint(strings.TrimSuffix(regv[1], "h"), 16, 32)
	if e != nil || addr < 0 || addr >= 65535 {
		log.Errorf("Invalid breakpoint address: %s", regv[1])
		return ""
	}
	if regv[0] == "PC" {
		computer.SetBreakpoint(uint16(no), uint16(addr))
	} else {
		log.Errorf("Unsupported BP: %s", params[1])
	}
	return ""
}

func BreakpointHit(no uint16) {
	if activeWriter != nil {
		pc := computer.GetCPUState().PC
		rep := fmt.Sprintf("Breakpoint fired: PC=%XH\n  %04X NOP", pc, pc)
		log.Debug(rep)
		writeResponseMessage(activeWriter, rep)
	}
}
