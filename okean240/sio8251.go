package okean240

type Sio8251 struct {
	counter uint64
}

type Sio8251Interface interface {
	Tick()
}

func NewSio8251() *Sio8251 {
	return &Sio8251{
		counter: 0,
	}
}

func (s *Sio8251) Tick() {
	s.counter++
}
