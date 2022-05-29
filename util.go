package main

import (
	"fmt"
	"os"
	"regexp"

	"go.uber.org/zap"
)

var (
	reTabEscape   = regexp.MustCompile("[\x00\x01\t\n]")
	reTabUnescape = regexp.MustCompile("\x01.")
)

func Close(message string, logger *zap.Logger, closeFunc func() error) {
	err := closeFunc()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Failed to %s", message), zap.Error(err))
		} else {
			_, _ = os.Stderr.WriteString(
				fmt.Sprintf("Failed to %s: %s", message, err),
			)
		}
	}
}

func TabEscape(s string) string {
	return reTabEscape.ReplaceAllStringFunc(s, func(m string) string {
		switch m {
		case "\x00":
			return "\x010"
		case "\x01":
			return "\x011"
		case "\t":
			return "\x01t"
		case "\r":
			return "\x01r"
		case "\n":
			return "\x01l"
		default:
			return m
		}
	})
}

func TabUnescape(s string) string {
	return reTabUnescape.ReplaceAllStringFunc(s, func(m string) string {
		switch m {
		case "\x010":
			return "\x00"
		case "\x011":
			return "\x01"
		case "\x01t":
			return "\t"
		case "\x01r":
			return "\r"
		case "\x01l":
			return "\n"
		default:
			return m[1:]
		}
	})
}
