package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/holoplot/go-evdev"
)

const virtualPrefix = "miburi"

type app struct {
	devicePath string
	button     evdev.EvCode
	threshold  int32
	overview   []evdev.EvCode
	left       []evdev.EvCode
	right      []evdev.EvCode

	mu           sync.Mutex
	physical     *evdev.InputDevice
	virtualMouse *evdev.InputDevice
	keyboard     *evdev.InputDevice
}

func main() {
	list := flag.Bool("list", false, "list input devices")
	devicePath := flag.String("device", "", "input device path, auto-detected when empty")
	buttonName := flag.String("button", "BTN_EXTRA", "trigger button name")
	threshold := flag.Int("threshold", 250, "movement in evdev units needed to fire a gesture")
	overviewCombo := flag.String("overview", "super", "keys for the up gesture")
	leftCombo := flag.String("left", "super+pageup", "keys for the left gesture")
	rightCombo := flag.String("right", "super+pagedown", "keys for the right gesture")
	flag.Parse()

	if *list {
		if err := listDevices(); err != nil {
			log.Fatal(err)
		}
		return
	}

	button, ok := evdev.KEYFromString[*buttonName]
	if !ok {
		log.Fatalf("unknown button %q", *buttonName)
	}

	overview, err := parseCombo(*overviewCombo)
	if err != nil {
		log.Fatalf("overview combo: %v", err)
	}
	left, err := parseCombo(*leftCombo)
	if err != nil {
		log.Fatalf("left combo: %v", err)
	}
	right, err := parseCombo(*rightCombo)
	if err != nil {
		log.Fatalf("right combo: %v", err)
	}

	keyboard, err := evdev.CreateDevice(
		virtualPrefix+" virtual keyboard",
		evdev.InputID{BusType: 0x03, Vendor: 0x6d69, Product: 0x6275, Version: 1},
		map[evdev.EvType][]evdev.EvCode{
			evdev.EV_KEY: comboUnion(overview, left, right),
		},
	)
	if err != nil {
		log.Fatalf("create virtual keyboard: %v", err)
	}

	a := &app{
		devicePath: *devicePath,
		button:     button,
		threshold:  int32(*threshold),
		keyboard:   keyboard,
		overview:   overview,
		left:       left,
		right:      right,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		a.shutdown()
	}()

	for {
		if err := a.session(); err != nil {
			log.Printf("session ended: %v", err)
		}
		time.Sleep(2 * time.Second)
	}
}

func (a *app) session() error {
	path := a.devicePath
	if path == "" {
		found, err := findMouse(a.button)
		if err != nil {
			return err
		}
		path = found
	}

	dev, err := evdev.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}

	name, _ := dev.Name()

	if err := dev.Grab(); err != nil {
		dev.Close()
		return fmt.Errorf("grab %s: %w", path, err)
	}

	virtualMouse, err := evdev.CloneDevice(virtualPrefix+" virtual mouse", dev)
	if err != nil {
		dev.Ungrab()
		dev.Close()
		return fmt.Errorf("clone %s: %w", path, err)
	}

	a.mu.Lock()
	a.physical = dev
	a.virtualMouse = virtualMouse
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.physical = nil
		a.virtualMouse = nil
		a.mu.Unlock()
		dev.Ungrab()
		dev.Close()
		evdev.DestroyDevice(virtualMouse)
		virtualMouse.Close()
	}()

	log.Printf("grabbed %q (%s)", name, path)

	e := &engine{
		mouse:     virtualMouse,
		keyboard:  a.keyboard,
		button:    a.button,
		threshold: a.threshold,
		overview:  a.overview,
		left:      a.left,
		right:     a.right,
	}

	for {
		ev, err := dev.ReadOne()
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := e.handle(ev); err != nil {
			return fmt.Errorf("handle event: %w", err)
		}
	}
}

func (a *app) shutdown() {
	a.mu.Lock()
	if a.physical != nil {
		a.physical.Ungrab()
		a.physical.Close()
	}
	if a.virtualMouse != nil {
		evdev.DestroyDevice(a.virtualMouse)
		a.virtualMouse.Close()
	}
	if a.keyboard != nil {
		evdev.DestroyDevice(a.keyboard)
		a.keyboard.Close()
	}
	a.mu.Unlock()
	os.Exit(0)
}

func findMouse(button evdev.EvCode) (string, error) {
	paths, err := evdev.ListDevicePaths()
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		if strings.Contains(p.Name, virtualPrefix) {
			continue
		}
		dev, err := evdev.Open(p.Path)
		if err != nil {
			continue
		}
		usable := hasCode(dev, evdev.EV_REL, evdev.REL_X) && hasCode(dev, evdev.EV_KEY, button)
		dev.Close()
		if usable {
			return p.Path, nil
		}
	}

	return "", errors.New("no mouse with the trigger button found")
}

func hasCode(dev *evdev.InputDevice, t evdev.EvType, code evdev.EvCode) bool {
	return slices.Contains(dev.CapableEvents(t), code)
}

func listDevices() error {
	paths, err := evdev.ListDevicePaths()
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		fmt.Println("no readable input devices, try running as root")
		return nil
	}
	for _, p := range paths {
		fmt.Printf("%s\t%s\n", p.Path, p.Name)
	}
	return nil
}
