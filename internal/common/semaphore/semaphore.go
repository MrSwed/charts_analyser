package semaphore

type S struct {
	semaCh chan struct{}
}

func New(maxReq int) *S {
	if maxReq <= 0 {
		maxReq = 1
	}
	return &S{
		semaCh: make(chan struct{}, maxReq),
	}
}

func (s *S) Acquire() {
	s.semaCh <- struct{}{}
}

func (s *S) Release() {
	<-s.semaCh
}
