package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 示例配置
const (
	// LogFileName 文件日志
	LogFileName = "./app-ff.log"
	// LogFileMaxSize 每个日志文件最大 MB
	LogFileMaxSize = 100
	// LogFileMaxBackups 保留日志文件个数
	LogFileMaxBackups = 10
	// LogFileMaxAge 保留日志最大天数
	LogFileMaxAge = 30

	// LogPeriod LogSampled 配置: 每 1 秒最多输出 3 条日志
	LogPeriod = time.Second
	LogBurst  = 3

	// LogLevel 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel     = 0
	LogHookLevel = 2
)

var (
	Log        zerolog.Logger
	LogSampled zerolog.Logger
	LogSubDemo zerolog.Logger
	Debug      bool
	NoColor    bool
)

func init() {
	flag.BoolVar(&Debug, "debug", false, "true 控制台日志, false 文件记录日志")
	flag.BoolVar(&NoColor, "nocolor", false, "生产环境中文本格式日志是否关闭高亮")
	flag.Parse()

	if err := InitLogger(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 路径脱敏, 日志格式规范, 避免与自定义字段名冲突: {"E":"is Err(error)","error":"is Str(error)"}
	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.ErrorFieldName = "E"
	zerolog.CallerFieldName = "F"
	zerolog.ErrorStackFieldName = "S"
	zerolog.DurationFieldInteger = true
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}

// InitLogger 配置热加载等场景调用, 重载日志环境
func InitLogger() error {
	if err := LogConfig(); err != nil {
		return err
	}

	// 抽样的日志记录器
	sampler := &zerolog.BurstSampler{
		Burst:  LogBurst,
		Period: LogPeriod,
	}
	LogSampled = Log.Sample(&zerolog.LevelSampler{
		TraceSampler: sampler,
		DebugSampler: sampler,
		InfoSampler:  sampler,
		WarnSampler:  sampler,
		ErrorSampler: sampler,
	})

	// 子记录器
	LogSubDemo = Log.With().Strs("Prefix", []string{"***", "FUFU", "***"}).Logger()

	return nil
}

// LogConfig 加载日志配置
func LogConfig() error {
	var (
		writers  []io.Writer
		basicLog = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	)

	if Debug {
		// 1. 开发环境时, 日志高亮输出到控制台
		writers = []io.Writer{basicLog}
	} else {
		// 2. 生产环境时, 日志输出到文件(可选关闭高亮, 保存最近 10 个 30 天内的日志), 并发送 JSON 日志到 ES
		basicLog.NoColor = NoColor
		basicLog.Out = &lumberjack.Logger{
			Filename:   LogFileName,
			MaxSize:    LogFileMaxSize,
			MaxAge:     LogFileMaxAge,
			MaxBackups: LogFileMaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
		writers = []io.Writer{basicLog, NewESWriter(zerolog.WarnLevel)}
	}

	Log = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Caller().Logger()
	Log = Log.Hook(logHookDemo{}).Level(LogLevel)

	return nil
}

// Hook 示例
type logHookDemo struct{}

func (h logHookDemo) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level >= LogHookLevel {
		e.Bool("HOOK", true)
	}
}

// ESWriter 自定义日志接收器
// 只接收指定级别及以上日志
type ESWriter struct {
	lv zerolog.Level
}

func NewESWriter(lv zerolog.Level) *ESWriter {
	return &ESWriter{
		lv: lv,
	}
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

func main() {
	start := time.Now()

	Log.Trace().Msg("test Trace")
	Log.Debug().Msg("test DEBUG")
	Log.Info().Msg("test INFO")
	Log.Warn().Msg("test WARN")

	err := errors.New("fake error")
	Log.Error().Err(err).Int("k1", 123).Msg("test ERROR with Field")

	// {"L":"error","E":"fake error","error":"my error msg",
	// "T":"2021-04-06T16:00:27+08:00","F":"main.go:153","HOOK":true,"M":"test ERROR json.key"}
	Log.Error().Err(err).Str("error", "my error msg").Msg("test ERROR json.key")

	// Log.Fatal().Err(err).Send()

	for i := 0; i < 10; i++ {
		LogSampled.Info().Msgf(">>>test LogSampled: %d", i)
		time.Sleep(200 * time.Millisecond)
	}

	LogSubDemo.Warn().Dur("dur.us", 3*time.Second).Msg("test LogSubDemo")

	Log.Info().TimeDiff("cost", time.Now(), start).Msg("test TimeDiff")
}
