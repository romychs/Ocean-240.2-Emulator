package usart

import log "github.com/sirupsen/logrus"

/**
Universal Serial Asynchronous Receiver/Transmitter
i8051, MSM82C51, КР580ВВ51

By Romych, 2025.03.04
*/

const I8251DSRFlag = 0x80
const I8251SynDetFlag = 0x40
const I8251FrameErrorFlag = 0x20
const I8251OverrunErrorFlag = 0x10
const I8251ParityErrorFlag = 0x08
const I8251TxEnableFlag = 0x04
const I8251RxReadyFlag = 0x02
const I8251TxReadyFlag = 0x01
const I8251TxBuffMaxLen = 16

const (
	Sio8251Reset = iota
	Sio8251LoadSyncChar1
	Sio8251LoadSyncChar2
	Sio8251LoadCommand
)

type I8251 struct {
	counter   uint64
	mode      byte
	initState byte
	syncChar1 byte
	syncChar2 byte
	bufferRx  []byte
	bufferTx  []byte
	rxe       bool
	txe       bool
}

type I8251Interface interface {
	Tick()
	Status() byte
	Reset()
	Command(value byte)
	Send(value byte)
	Receive() byte
}

func NewI8251() *I8251 {
	return &I8251{
		counter:   0,
		mode:      0,
		initState: 0,
		rxe:       false,
		txe:       false,
		bufferRx:  []byte{},
		bufferTx:  []byte{},
	}
}

func (s *I8251) Tick() {
	s.counter++
}

// Status i8251 status [RST,RQ_RX,RST_ERR,PAUSE,RX_EN,RX_RDY,TX_RDY]
func (s *I8251) Status() byte {
	var status byte = 0
	if len(s.bufferRx) > 0 {
		status |= I8251RxReadyFlag
	}
	if len(s.bufferTx) < I8251TxBuffMaxLen {
		status |= I8251TxReadyFlag
	}
	if s.txe {
		status |= I8251TxEnableFlag
	}
	return status
}

func (s *I8251) Reset() {
	s.counter = 0
	s.mode = 0
	s.initState = 0
	s.bufferRx = make([]byte, 8)
	s.bufferTx = make([]byte, I8251TxBuffMaxLen)
	s.rxe = false
	s.txe = false
}

func (s *I8251) Command(value byte) {
	switch s.initState {
	case Sio8251Reset:
		s.mode = value
		if s.mode&0x03 > 0 {
			// SYNC
			s.initState = Sio8251LoadSyncChar1
		}
		// ASYNC
		s.initState = Sio8251LoadCommand
	case Sio8251LoadSyncChar1:
		s.mode = value
		if s.mode&0x80 == 0 { // SYNC DOUBLE
			s.initState = Sio8251LoadSyncChar2
		}
	case Sio8251LoadSyncChar2:
		s.mode = value
		s.initState = Sio8251LoadCommand
	case Sio8251LoadCommand:
		// value = command
		if value&0x40 != 0 {
			// RESET CMD
			s.Reset()
		} else {
			// Set RXE, TXE
			if value&0x04 != 0 {
				s.rxe = true
			} else {
				s.rxe = false
			}
			if value&0x01 != 0 {
				s.txe = true
			} else {
				s.txe = false
			}
		}
	}

}

func (s *I8251) Send(value byte) {
	if s.txe {
		s.bufferTx = append(s.bufferTx, value)
	}
}

func (s *I8251) Receive() byte {

	if s.rxe {
		if len(s.bufferRx) > 0 {
			res := s.bufferRx[0]
			s.bufferRx = s.bufferRx[1:]
			log.Debugf("ReceiveByte: %x", res)
			return res
		}
	}
	log.Debugf("ReceiveByte: empty buffer")
	return 0
}

func (s *I8251) SetRxBytes(bytes []byte) {
	s.bufferRx = append(s.bufferRx, bytes...)
}
