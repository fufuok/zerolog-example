package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

var (
	Log zerolog.Logger
)

// ESWriter 自定义日志接收器
// 只接收指定级别及以上日志
type ESWriter struct {
	lv zerolog.Level
}

// Write 发送日志到 ES
func (w *ESWriter) Write(p []byte) (n int, err error) {
	fmt.Print("__TO_ES_:", string(p))
	return len(p), nil
}

// WriteLevel 日志级别过滤
func (w *ESWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= w.lv {
		return w.Write(p)
	}
	return len(p), nil
}

// LevelFileWriter 指定级别的日志写文件
type LevelFileWriter struct {
	lw io.Writer
	lv zerolog.Level
}

func (w *LevelFileWriter) Write(p []byte) (n int, err error) {
	return w.lw.Write(p)
}

func (w *LevelFileWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l == w.lv {
		return w.lw.Write(p)
	}
	return len(p), nil
}

// LevelConsoleWriter 特定级别日志写到控制台
type LevelConsoleWriter struct {
	lw zerolog.ConsoleWriter
	lv []zerolog.Level
}

func (w *LevelConsoleWriter) Write(p []byte) (n int, err error) {
	return w.lw.Write(p)
}

func (w *LevelConsoleWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	for _, v := range w.lv {
		if v == l {
			return w.lw.Write(p)
		}
	}
	return len(p), nil
}

func InitLogger() {
	errorFile, _ := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	warnFile, _ := os.OpenFile("warn.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	Log = zerolog.New(zerolog.MultiLevelWriter(
		// Warn 级别及以上的日志发往 ES
		&ESWriter{lv: zerolog.WarnLevel},
		// Warn 日志写入 warn.log
		&LevelFileWriter{lw: warnFile, lv: zerolog.WarnLevel},
		// Error 日志写入 error.log
		&LevelFileWriter{lw: errorFile, lv: zerolog.ErrorLevel},
		// Debug, Error 日志显示在控制台
		&LevelConsoleWriter{
			lw: zerolog.ConsoleWriter{Out: os.Stdout},
			lv: []zerolog.Level{zerolog.DebugLevel, zerolog.ErrorLevel},
		},
	)).With().Timestamp().Caller().Logger()
}

func main() {
	InitLogger()
	Log.Trace().Msg("test TRACE")
	Log.Debug().Msg("test DEBUG")
	Log.Info().Msg("test INFO")
	Log.Warn().Msg("test WARN")
	Log.Error().Msg("test ERROR")
	Log.Fatal().Msg("test FATAL")
}
