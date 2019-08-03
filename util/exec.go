package util

type SingleThread struct {
	txmit chan<- func()
}

func NewSingleThread() *SingleThread {
	ch := make(chan func())
	st := &SingleThread{
		txmit: ch,
	}
	go func() {
		for f := range ch {
			f()
		}
	}()
	return st
}

// runs the command on the single thread
func (st *SingleThread) Run(f func()) {
	done := make(chan struct{})
	st.txmit <- func() {
		f()
		done <- struct{}{}
	}
	<-done
}
