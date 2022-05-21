package log

import (
	"os"
	"path/filepath"
	"time"

	"projtmpl/env"

	"github.com/phuslu/log"
)

var Logger *log.Logger

// Setup 配置日志
func Setup() {
	tier := env.Envs.Tier
	var writer log.Writer
	switch tier {
	case env.Testing, env.Staging, env.Prod:
		writer = newFileLogger()
	default:
		writer = newConsoleWriter()
	}
	Logger = &log.Logger{
		Level:      log.ParseLevel("info"),
		Writer:     writer,
		TimeFormat: "2006-01-02 15:04:05",
		Context: log.NewContext(nil).
			Str("tier", env.Envs.Tier).
			Value(),
	}
}

// newConsoleWriter 打印日志到控制台，适合在本地开发环境使用
func newConsoleWriter() log.Writer {
	return &log.ConsoleWriter{
		ColorOutput:    true,
		QuoteString:    true,
		EndWithMessage: true,
	}
}

// newFileLogger 打印日志文件到文件，适合在测试和生产环境使用
func newFileLogger() log.Writer {
	return &log.FileWriter{
		Filename: "./logs/main.log", // 指定日志文件的名字
		MaxSize:  16 * 1024 * 1024,  // 指定单个日志文件最大大小
		Cleaner: func(filename string, maxBackups int, matches []os.FileInfo) { // 指定过老日志清理策略
			// 清理一个月前的日志文件
			var dir = filepath.Dir(filename)
			t := time.Now().UTC().Add(-time.Hour * 24 * 30)
			for i := 0; i < len(matches); i++ {
				if matches[i].ModTime().Before(t) {
					_ = os.Remove(filepath.Join(dir, matches[i].Name()))
				}
			}
		},
	}
}
