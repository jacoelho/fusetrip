package main

import (
	"fmt"
	"time"
)

const (
	StateClosed = iota
	StateOpen
)

type Fuse struct {
	State   int
	TimeOut time.Duration
}

func NewFuse() *Fuse {
	return &Fuse{
		State:   StateClosed,
		TimeOut: 3,
	}
}

func (f *Fuse) IsOpen() bool {
	return f.State == StateOpen
}

func (f *Fuse) Connected(fn func() error) *Fuse {
	if !f.IsOpen() {
		wait := make(chan error, 1)

		go func() {
			wait <- fn()
		}()

		select {
		case res := <-wait:
			if res != nil {
				f.State = StateOpen
			}
		case <-time.After(time.Second * f.TimeOut):
			f.State = StateOpen
		}
	}
	return f
}

func (f *Fuse) Tripped(fn func()) *Fuse {
	if f.IsOpen() {
		fn()
	}
	return f
}

func weather(locations string) string {
	time.Sleep(2 * time.Second)
	return "raining"
}

func fallback() string {
	return "sunny"
}

func main() {
	fuse := NewFuse()

	var myWeather string

	fuse.Connected(func() error {
		myWeather = weather("leeds")
		return nil
	}).Tripped(func() {
		myWeather = fallback()
	})

	fmt.Println(myWeather)
}
