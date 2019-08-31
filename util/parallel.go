package util

import "github.com/hashicorp/go-multierror"

func RunInParallel(targets ...func() error) error {
	errorCh := make(chan error)
	panicCh := make(chan interface{})
	for _, target := range targets {
		go func(target func() error) {
			var err error
			defer func() {
				errorCh <- err
				panicCh <- recover()
			}()
			err = target()
		}(target)
	}
	var errors error
	var firstPanic interface{}
	for range targets {
		err := <-errorCh
		if err != nil {
			if errors == nil {
				errors = err
			} else {
				errors = multierror.Append(errors, err)
			}
		}
		panicInfo := <-panicCh
		if panicInfo != nil {
			if firstPanic == nil {
				// print this here just in case something prevents us from finishing this function
				if s, ok := panicInfo.(string); ok {
					println("hit panic:", s)
				} else {
					println("hit panic:", panicInfo)
				}
				firstPanic = panicInfo
			} else {
				println("dropping additional panic:", panicInfo)
			}
		}
	}
	if firstPanic != nil {
		panic(firstPanic)
	}
	return errors
}
