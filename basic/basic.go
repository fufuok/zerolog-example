package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// 示例配置
const (
	// 文件日志
	LogFileName = "./app-ff.log"
	// 每个日志文件最大 MB
	LogFileMaxSize = 100
	// 保留日志文件个数
	LogFileMaxBackups = 10
	// 保留日志最大天数
	LogFileMaxAge = 30

	// LogSampled 配置: 每 1 秒最多输出 3 条日志
	LogPeriod = time.Second
	LogBurst  = 3

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel = 0
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

	// 日志格式规范, 路径脱敏
	zerolog.MessageFieldName = "msg"
	zerolog.DurationFieldInteger = true
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}

// 配置热加载等场景调用, 重载日志环境
func InitLogger() error {
	if err := LogConfig(); err != nil {
		return err
	}

	// 抽样的日志记录器
	LogSampled = Log.Sample(&zerolog.LevelSampler{
		InfoSampler: &zerolog.BurstSampler{
			Burst:  LogBurst,
			Period: LogPeriod,
		},
	})

	// 子记录器
	LogSubDemo = Log.With().Strs("Prefix", []string{"***", "FUFU", "***"}).Logger()

	return nil
}

// 加载日志配置
// 1. 开发环境时, 日志高亮输出到控制台
// 2. 生产环境时, 日志输出到文件(可选关闭高亮, 保存最近 10 个 30 天内的日志)
func LogConfig() error {
	basicLog := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}

	if !Debug {
		basicLog.NoColor = NoColor
		basicLog.Out = &lumberjack.Logger{
			Filename:   LogFileName,
			MaxSize:    LogFileMaxSize,
			MaxAge:     LogFileMaxAge,
			MaxBackups: LogFileMaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
	}

	Log = zerolog.New(basicLog).With().Timestamp().Caller().Logger()
	Log = Log.Level(LogLevel)

	return nil
}

func main() {
	start := time.Now()

	Log.Trace().Msg("test Trace")
	Log.Debug().Msg("test DEBUG")
	Log.Info().Msg("test INFO")
	Log.Warn().Msg("test WARN")

	err := errors.New("fake error")
	Log.Error().Err(err).Int("k1", 123).Msg("test ERROR with Field")
	// Log.Fatal().Err(err).Send()

	for i := 0; i < 10; i++ {
		LogSampled.Info().Msgf(">>>test LogSampled: %d", i)
		time.Sleep(200 * time.Millisecond)
	}

	LogSubDemo.Warn().Dur("dur.us", 3*time.Second).Msg("test LogSubDemo")

	Log.Info().TimeDiff("cost", time.Now(), start).Msg("test TimeDiff")
}
