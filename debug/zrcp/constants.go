package zrcp

const welcomeMessage = "Welcome to Ocean-240.2 remote command protocol (ZRCP partial implementation)\nWrite help for available commands\n\ncommand> "
const emptyResponse = "\ncommand> "
const aboutResponse = "ZEsarUX remote command protocol"
const getVersionResponse = "12.1"
const getRegistersResponse = "PC=%04x SP=%04x AF=%04x BC=%04x HL=%04x DE=%04x IX=%04x IY=%04x AF'=%04x BC'=%04x HL'=%04x DE'=%04x I=%02x R=%02x  F=%s F'=%s MEMPTR=%04x IM0 IFF%s VPS: 0 MMU=00000000000000000000000000000000"
const getStateResponse = "PC=%04x SP=%04x AF=%04x BC=%04x HL=%04x DE=%04x IX=%04x IY=%04x AF'=%04x BC'=%04x HL'=%04x DE'=%04x I=%02x R=%02x IM0 IFF%s (PC)=%s (SP)=%s MMU=00000000000000000000000000000000"

const inCpuStepResponse = "\ncommand@cpu-step> "
const getMachineResponse = "64K RAM, no ZX\n"
const respErrorLoading = "ERROR loading file"
const quitResponse = "Sayonara baby\n"
const runUntilBPMessage = "Running until a breakpoint, key press or data sent, menu opening or other event\n"
