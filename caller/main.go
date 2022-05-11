package main

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		i := strings.LastIndexByte(file, '/')
		if i == -1 {
			return file
		}
		i = strings.LastIndexByte(file[:i], '/')
		if i == -1 {
			return file
		}
		return file[i+1:] + ":" + strconv.Itoa(line)
	}
	log.Info().Caller().Msg("caller is like package/file:line")
}
