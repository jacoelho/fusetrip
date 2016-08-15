package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const (
	StateClosed uint32 = iota // circuit is working
	StateOpen                 // circuit have failed
)

var (
	ErrorFuseTripped = errors.New("fused tripped")
)

type Fuse struct {
	TimeOut        time.Duration // how long to wait for the execution
	FailThreshold  uint32        // how many fails until circuit is tripped
	RetryThreshold uint32        // how many failed requests until try again
	failCounter    uint32
	retries        uint32
}

// circuit is working again
func (f *Fuse) reset() {
	f.failCounter = 0
	f.retries = 0
}

// circuit failed
func (f *Fuse) increment() {
	f.failCounter++
	f.retries++
}

// should we check if circuit is working again?
func (f *Fuse) shouldRetry() bool {
	if f.retries > f.RetryThreshold {
		f.retries = 0
		return true
	}
	return false
}

func (f *Fuse) isOpen() bool {
	if f.shouldRetry() {
		return false
	} else {
		return f.failCounter > f.FailThreshold
	}
}

func (f *Fuse) connected(fn func() error) error {
	if !f.isOpen() {
		wait := make(chan error, 1)

		// run function with timeout
		go func() { wait <- fn() }()

		select {
		case err := <-wait:
			if err != nil {
				break
			} else {
				f.reset()
				return nil
			}
		case <-time.After(time.Second * f.TimeOut):
			break
		}
	}
	f.increment()
	return ErrorFuseTripped
}

func (f *Fuse) WithCircuit(regular func() error, tripped func()) error {
	err := f.connected(regular)

	if err != nil {
		tripped()
	}

	return err
}

func fetchWeather(location string) string {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return "sunny"
}

func fallbackWeather() string {
	return "raining"
}

func main() {
	fuse := &Fuse{
		TimeOut:        3,
		FailThreshold:  5,
		RetryThreshold: 5,
	}

	var localWeather string

	for {
		fuse.WithCircuit(func() error {
			localWeather = fetchWeather("London")
			return nil
		},
			func() {
				localWeather = fallbackWeather()
			})

		fmt.Println(localWeather)
	}
}
