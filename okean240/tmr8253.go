package okean240

/*
	Timer config: [sc1,sc0][rl1,rl0][m2,m1,m0][bcd]
	sc - timer, rl=01-LSB, 10-MSB, 11-LSB+MSB
	mode 000 - int on fin,
		 001 - one shot,
		 x10 - rate gen,
		 x11 - sq wave
*/

const (
	TimerModeIntOnFin = iota
	TimerModeOneShot
	TimerModeRateGen
	TimerModeSqWave
)

const (
	TimerRLMsbLsb = iota
	TimerRLLsbLsb
	TimerRLMsb
	TimerRLLsbMsb
)

type Timer8253Ch struct {
	rl      byte
	mode    byte
	bcd     bool
	load    uint16
	counter uint16
	fb      bool
	start   bool
	fired   bool
}
type Timer8253 struct {
	//chNo    byte
	channel [3]Timer8253Ch
}

type Timer8253Interface interface {
	//Init()
	Configure(value byte)
	Load(chNo int, value byte)
	Counter(chNo int) uint16
	Fired(chNo int) bool
	Start(chNo int) bool
}

func NewTimer8253() *Timer8253 {
	return &Timer8253{
		//chNo: 0,
		channel: [3]Timer8253Ch{
			{0, 0, false, 0, 0, true, false, false},
			{0, 0, false, 0, 0, true, false, false},
			{0, 0, false, 0, 0, true, false, false},
		},
	}
}

func (t *Timer8253) Tick(chNo int) {
	tmr := &t.channel[chNo]
	if tmr.start {
		tmr.counter--
		if tmr.counter == 0 {
			switch tmr.mode {
			case TimerModeIntOnFin:
				{
					tmr.start = false
					tmr.fired = true
				}
			case TimerModeOneShot:
				tmr.start = false
			case TimerModeRateGen:
				tmr.start = false
			case TimerModeSqWave:
				{
					tmr.start = true
					tmr.counter = tmr.load
					tmr.fired = true
				}
			}
		}
	}
}

func (t *Timer8253) Counter(chNo int) uint16 {
	return t.channel[chNo].counter
}

func (t *Timer8253) Fired(chNo int) bool {
	f := t.channel[chNo].fired
	if f {
		t.channel[chNo].fired = false
	}
	return f
}

func (t *Timer8253) Start(chNo int) bool {
	return t.channel[chNo].start
}

func (t *Timer8253) Configure(value byte) {
	chNo := (value & 0xC0) >> 6
	rl := value & 0x30 >> 4
	t.channel[chNo].start = false
	t.channel[chNo].rl = rl
	t.channel[chNo].mode = (value & 0x0E) >> 1
	t.channel[chNo].fb = true
	t.channel[chNo].bcd = value&0x01 == 1
	t.channel[chNo].load = 0
}

func (t *Timer8253) Load(chNo byte, value byte) {
	timer := &t.channel[chNo]
	if timer.rl == 0 {
		// MSB+LSB
		if timer.fb {
			// MSB
			timer.load = uint16(value) << 8
			timer.fb = false
		} else {
			// LSB
			timer.load |= uint16(value)
			timer.start = true
		}
	} else if timer.rl == 1 {
		// LSB Only
		timer.load = (timer.load & 0xff00) | uint16(value)
		timer.start = true
	} else if timer.rl == 2 {
		// MSB Only
		timer.load = (timer.load & 0x00ff) | (uint16(value) << 8)
		timer.start = true
	} else {
		// LSB+MSB
		if timer.fb {
			// LSB
			timer.load = uint16(value)
			timer.fb = false
		} else {
			// MSB
			timer.load = (uint16(value) << 8) | (timer.load & 0x00ff)
			timer.start = true
			timer.counter = timer.load
		}
	}
}
