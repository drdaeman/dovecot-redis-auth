package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const CmdHello = 'H'
const CmdLookup = 'L'
const CmdIterate = 'I'

const ResponseOk = 'O'
const ResponseNotFound = 'N'
const ResponseFailure = 'F'

const FlagRecurse = 0x01
const FlagSortByKey = 0x02
const FlagSortByValue = 0x04
const FlagNoValue = 0x08
const FlagExactKey = 0x10
const FlagAsync = 0x20

func handleRequest(ctx context.Context, req *Request) error {
	log := req.Log

	switch req.Command {
	case CmdHello:
		// HELLO: major, minor, value type, obsolete user, dict name
		if len(req.Args) != 5 {
			req.Fatal(fmt.Errorf("protocol error: expected 5 arguments, received %d", len(req.Args)))
			return nil
		}
		if req.Args[0] != "3" {
			req.Fatal(fmt.Errorf("incompatible major protocol version '%s'", req.Args[0]))
			return nil
		}
		return nil // No response!
	case CmdLookup:
		// LOOKUP: key, user
		if len(req.Args) != 2 {
			return fmt.Errorf("expected 2 arguments, received %d", len(req.Args))
		}
		key := req.Args[0]
		log.Info("Received a lookup request", zap.String("Key", key))

		if strings.HasPrefix(key, "shared/") {
			log.Debug("Stripped the 'shared/' prefix", zap.String("Key", key))
			key = key[7:]
		}

		value, err := Lookup(ctx, key)
		if err != nil {
			return err
		}
		if value == "" {
			return req.Respond(ResponseNotFound)
		}
		return req.Respond(ResponseOk, value)
	case CmdIterate:
		// ITERATE: flags, max rows, path, user
		if len(req.Args) != 4 {
			return fmt.Errorf("expected 4 arguments, received %d", len(req.Args))
		}

		flags, err := ParseBitFlags(req.Args[0])
		if err != nil {
			return fmt.Errorf("malformed flags value: %w", err)
		}

		if flags.Has(FlagAsync) {
			return errors.New("asynchronous iteration is not supported")
		}
		if flags.Has(FlagRecurse) {
			log.Info("Recursion to all sub-hierarchies was requested")
		}
		if flags.Has(FlagSortByKey) && flags.Has(FlagSortByValue) {
			return errors.New("must request either sorting by key or value, not both")
		}

		maxRows, err := strconv.Atoi(req.Args[1])
		if err != nil {
			return fmt.Errorf("malformed max rows value: %w", err)
		}

		var match string
		var prefix string
		if flags.Has(FlagExactKey) {
			match = req.Args[2]
		} else {
			match = fmt.Sprintf("%s*", req.Args[2])
		}
		if strings.HasPrefix(match, "shared/") {
			prefix = "shared/"
			log.Debug("Stripped the 'shared/' prefix", zap.String("Match", match))
			match = match[7:]
		}

		log.Info("Received an iterate request", zap.String("Match", match))

		result, err := Iterate(ctx, match, !flags.Has(FlagNoValue), maxRows)
		if err != nil {
			return err
		}

		if flags.Has(FlagSortByKey) {
			By(ByKey).Sort(result)
		}
		if flags.Has(FlagSortByValue) {
			By(ByValue).Sort(result)
		}

		for _, item := range result {
			key := item.Key
			if prefix != "" {
				key = fmt.Sprintf("%s%s", prefix, key)
			}
			if flags.Has(FlagNoValue) {
				err = req.Respond(ResponseOk, key)
			} else {
				err = req.Respond(ResponseOk, key, item.Value)
			}
			if err != nil {
				return err
			}
		}
		return req.RespondEmptyLine()
	default:
		return req.Respond(ResponseFailure, "unsupported command")
	}
}
