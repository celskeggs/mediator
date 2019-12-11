package util

import (
	"github.com/hashicorp/go-multierror"
	"sync"
)

type ParallelElements struct {
	count      uint
	stopped    bool
	countMutex sync.Mutex
	results    chan parallelResult
}

func NewParallel() *ParallelElements {
	return &ParallelElements{
		results: make(chan parallelResult),
	}
}

type parallelResult struct {
	Err   error
	Panic interface{}
}

func (p *ParallelElements) incrementCount() {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()
	if p.stopped {
		panic("attempt to Add new target after ParallelElements has stopped!")
	}
	p.count += 1
}

func (p *ParallelElements) checkFinished(count uint) bool {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()
	if count > p.count {
		panic("should never have more finished units than total count!")
	} else if count < p.count {
		return false // not yet
	} else {
		// we got them all, so we shouldn't get any more jobs
		p.stopped = true
		return true
	}
}

func (p *ParallelElements) Add(target func() error) {
	p.incrementCount()
	go func() {
		var err error
		defer func() {
			p.results <- parallelResult{
				Err:   err,
				Panic: recover(),
			}
		}()
		err = target()
	}()
}

func (p *ParallelElements) Join() error {
	var received uint
	var errors error
	var firstPanic interface{}
	for !p.checkFinished(received) {
		result := <-p.results
		received += 1
		if result.Err != nil {
			if errors == nil {
				errors = result.Err
			} else {
				errors = multierror.Append(errors, result.Err)
			}
		}
		if result.Panic != nil {
			if firstPanic == nil {
				// print this here just in case something prevents us from finishing this function
				if s, ok := result.Panic.(string); ok {
					println("hit panic:", s)
				} else {
					println("hit panic:", result.Panic)
				}
				firstPanic = result.Panic
			} else {
				println("dropping additional panic:", result.Panic)
			}
		}
	}
	if firstPanic != nil {
		panic(firstPanic)
	}
	return errors
}

func RunInParallel(targets ...func() error) error {
	pe := NewParallel()
	for _, target := range targets {
		pe.Add(target)
	}
	return pe.Join()
}
