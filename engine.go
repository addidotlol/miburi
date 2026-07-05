package main

import (
	"log"
	"slices"
	"time"

	"github.com/holoplot/go-evdev"
)

type engine struct {
	mouse     *evdev.InputDevice
	keyboard  *evdev.InputDevice
	button    evdev.EvCode
	threshold int32
	overview  []evdev.EvCode
	left      []evdev.EvCode
	right     []evdev.EvCode

	holding bool
	fired   bool
	dx, dy  int32
}

func (e *engine) handle(ev *evdev.InputEvent) error {
	if ev.Type == evdev.EV_KEY && ev.Code == e.button {
		return e.handleButton(ev.Value)
	}
	if e.holding && !e.fired && ev.Type == evdev.EV_REL && (ev.Code == evdev.REL_X || ev.Code == evdev.REL_Y) {
		if err := e.accumulate(ev); err != nil {
			return err
		}
	}
	return e.mouse.WriteOne(ev)
}

func (e *engine) handleButton(value int32) error {
	switch value {
	case 1:
		e.holding = true
		e.fired = false
		e.dx, e.dy = 0, 0
	case 0:
		wasClick := e.holding && !e.fired
		e.holding = false
		if wasClick {
			return e.clickThrough()
		}
	}
	return nil
}

func (e *engine) accumulate(ev *evdev.InputEvent) error {
	if ev.Code == evdev.REL_X {
		e.dx += ev.Value
	} else {
		e.dy += ev.Value
	}

	if abs(e.dx) >= e.threshold && abs(e.dx) >= abs(e.dy) {
		e.fired = true
		if e.dx > 0 {
			log.Print("gesture: workspace right")
			return e.tap(e.right...)
		}
		log.Print("gesture: workspace left")
		return e.tap(e.left...)
	}

	if -e.dy >= e.threshold && abs(e.dy) > abs(e.dx) {
		e.fired = true
		log.Print("gesture: overview")
		return e.tap(e.overview...)
	}

	return nil
}

func (e *engine) clickThrough() error {
	if err := e.writeKey(e.mouse, e.button, 1); err != nil {
		return err
	}
	return e.writeKey(e.mouse, e.button, 0)
}

func (e *engine) tap(codes ...evdev.EvCode) error {
	for _, c := range codes {
		if err := e.writeKey(e.keyboard, c, 1); err != nil {
			return err
		}
		time.Sleep(3 * time.Millisecond)
	}
	for _, code := range slices.Backward(codes) {
		if err := e.writeKey(e.keyboard, code, 0); err != nil {
			return err
		}
		time.Sleep(3 * time.Millisecond)
	}
	return nil
}

func (e *engine) writeKey(dev *evdev.InputDevice, code evdev.EvCode, value int32) error {
	if err := dev.WriteOne(&evdev.InputEvent{Type: evdev.EV_KEY, Code: code, Value: value}); err != nil {
		return err
	}
	return dev.WriteOne(&evdev.InputEvent{Type: evdev.EV_SYN, Code: evdev.SYN_REPORT})
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
