package util

import (
	"fmt"
	"time"
)

type entry struct {
	Fn   func()
	Name string
	Done chan<- struct{}
}

type SingleThread struct {
	txmit chan<- entry
	mon   <-chan string
}

func NewSingleThread() *SingleThread {
	ch := make(chan entry)
	mon := make(chan string)
	st := &SingleThread{
		txmit: ch,
		mon:   mon,
	}
	st.monitor()
	go func() {
		for f := range ch {
			mon <- f.Name
			f.Fn()
			f.Done <- struct{}{}
			mon <- ""
		}
	}()
	return st
}

// makes sure that the thread doesn't get stuck
func (st *SingleThread) monitor() {
	ticker := time.Tick(time.Second)
	go func() {
		var running string
		warning := false
		stuck := false
		for {
			select {
			case active := <-st.mon:
				running = active
				if running == "" {
					warning = false
					stuck = false
				}
			case <-ticker:
				if running != "" {
					if !warning {
						warning = true
					} else {
						// we've had two seconds elapse since the last time we heard from the thread, and we know it's
						// actively trying to run something. so we must have stalled!
						if stuck {
							fmt.Printf("*** still stuck while running SingleThread func: %v\n", running)
						} else {
							stuck = true
							fmt.Printf("*** got stuck while running SingleThread func: %v\n", running)
						}
					}
				}
			}
		}
	}()
}

// runs the command on the single thread
func (st *SingleThread) Run(name string, f func()) {
	done := make(chan struct{})
	st.txmit <- entry{
		Fn:   f,
		Name: name,
		Done: done,
	}
	<-done
}
