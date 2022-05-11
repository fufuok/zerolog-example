package main

import (
	"os"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var Log zerolog.Logger

func init() {
	// 配合 pkg/errors 包
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Caller().Logger()
}

func main() {
	err := outer()
	log.Error().Stack().Err(err).Msg("")

	// 输出到控制台
	Log.Error().Stack().Err(err).Msg("")

	R()
}

func inner() error {
	return errors.New("seems we have an error here")
}

func middle() error {
	err := inner()
	if err != nil {
		return err
	}
	return nil
}

func outer() error {
	err := middle()
	if err != nil {
		return err
	}
	return nil
}

func R() {
	defer func() {
		if r := recover(); r != nil {
			Log.Error().Interface("r", r).Bytes("stack", debug.Stack()).Msg("Recovery from panic")
		}
	}()
	P()
}

func P() {
	panic([]int{1, 2, 3})
}
