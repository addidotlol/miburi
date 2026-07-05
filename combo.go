package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/holoplot/go-evdev"
)

var modifierAliases = map[string]evdev.EvCode{
	"ctrl":  evdev.KEY_LEFTCTRL,
	"shift": evdev.KEY_LEFTSHIFT,
	"alt":   evdev.KEY_LEFTALT,
	"super": evdev.KEY_LEFTMETA,
	"meta":  evdev.KEY_LEFTMETA,
}

func parseCombo(s string) ([]evdev.EvCode, error) {
	var codes []evdev.EvCode
	for part := range strings.SplitSeq(s, "+") {
		token := strings.ToLower(strings.TrimSpace(part))
		if token == "" {
			continue
		}
		if code, ok := modifierAliases[token]; ok {
			codes = append(codes, code)
			continue
		}
		if code, ok := evdev.KEYFromString["KEY_"+strings.ToUpper(token)]; ok {
			codes = append(codes, code)
			continue
		}
		if code, ok := evdev.KEYFromString[strings.ToUpper(token)]; ok {
			codes = append(codes, code)
			continue
		}
		return nil, fmt.Errorf("unknown key %q", part)
	}
	if len(codes) == 0 {
		return nil, errors.New("empty combo")
	}
	return codes, nil
}

func comboUnion(combos ...[]evdev.EvCode) []evdev.EvCode {
	seen := map[evdev.EvCode]bool{}
	var union []evdev.EvCode
	for _, combo := range combos {
		for _, code := range combo {
			if !seen[code] {
				seen[code] = true
				union = append(union, code)
			}
		}
	}
	return union
}
