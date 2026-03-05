package pit

/*
    Programmable Interval Timer
	i8053, MSM82C53, КР580ВИ53

	By Romych, 2025.03.04
*/

// Timer work modes
const (
	TimerModeIntOnFin = iota
	TimerModeOneShot
	TimerModeRateGen
	TimerModeSqWave
)

// Timer load counter modes
const (
	TimerRLMsbLsb = iota
	TimerRLLsb
	TimerRLMsb
	TimerRLLsbMsb
)

type Timer8253Ch struct {
	rl      byte   // load mode
	mode    byte   // counter mode
	bcd     bool   // decimal/BCD load mode
	load    uint16 // value to count from
	counter uint16 // timer counter
	fb      bool   // first byte load flag
	started bool   // true if timer started
	fired   bool
}
type I8253 struct {
	//chNo    byte
	channel [3]Timer8253Ch
}

type I8253Interface interface {
	//Init()
	Configure(value byte)
	Load(chNo int, value byte)
	Counter(chNo int) uint16
	Fired(chNo int) bool
	Start(chNo int) bool
}

func NewI8253() *I8253 {
	return &I8253{
		//chNo: 0,
		channel: [3]Timer8253Ch{
			{0, 0, false, 0, 0, true, false, false},
			{0, 0, false, 0, 0, true, false, false},
			{0, 0, false, 0, 0, true, false, false},
		},
	}
}

func (t *I8253) Tick(chNo int) {
	tmr := &t.channel[chNo]
	if tmr.started {
		tmr.counter--
		if tmr.counter == 0 {
			switch tmr.mode {
			case TimerModeIntOnFin:
				{
					tmr.started = false
					tmr.fired = true
				}
			case TimerModeOneShot:
				tmr.started = false
			case TimerModeRateGen:
				tmr.started = false
			case TimerModeSqWave:
				{
					tmr.started = true
					tmr.counter = tmr.load
					tmr.fired = true
				}
			}
		}
	}
}

func (t *I8253) Counter(chNo int) uint16 {
	return t.channel[chNo].counter
}

func (t *I8253) Fired(chNo int) bool {
	f := t.channel[chNo].fired
	if f {
		t.channel[chNo].fired = false
	}
	return f
}

func (t *I8253) Start(chNo int) bool {
	return t.channel[chNo].started
}

/*
	Timer config byte: [sc1:0][rl1:0][m2:0][bcd]
	sc1:0 - timer No
    rl=01-LSB, 10-MSB, 11-LSB+MSB
	mode 000 - intRq on fin,
		 001 - one shot,
		 x10 - rate gen,
		 x11 - sq wave
*/

func (t *I8253) Configure(value byte) {
	chNo := (value & 0xC0) >> 6
	rl := value & 0x30 >> 4
	t.channel[chNo].started = false
	t.channel[chNo].rl = rl
	t.channel[chNo].mode = (value & 0x0E) >> 1
	t.channel[chNo].fb = true
	t.channel[chNo].bcd = value&0x01 == 1
	t.channel[chNo].load = 0
}

func (t *I8253) Load(chNo byte, value byte) {
	timer := &t.channel[chNo]
	switch timer.rl {
	case TimerRLMsbLsb:
		// MSB+LSB
		if timer.fb {
			// MSB
			timer.load = uint16(value) << 8
			timer.fb = false
		} else {
			// LSB
			timer.load |= uint16(value)
			timer.started = true
		}
	case TimerRLLsb:
		// LSB Only
		timer.load = (timer.load & 0xff00) | uint16(value)
		timer.started = true
	case TimerRLMsb:
		// MSB Only
		timer.load = (timer.load & 0x00ff) | (uint16(value) << 8)
		timer.started = true
	case TimerRLLsbMsb:
		// LSB+MSB
		if timer.fb {
			// LSB
			timer.load = uint16(value)
			timer.fb = false
		} else {
			// MSB
			timer.load = (uint16(value) << 8) | (timer.load & 0x00ff)
			timer.started = true
			timer.counter = timer.load
		}
	}
}
