package log

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/juicity/juicity/common"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger = zerolog.Logger

type Options struct {
	Output     string
	TimeFormat string
	FileFormat string
	NoColor    bool
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func NewLogger(opt *Options) *Logger {
	var writerInputs []io.Writer
	if opt.Output == "" {
		opt.Output = "console"
	}
	outputs := strings.Split(opt.Output, ",")
	for i := range outputs {
		outputs[i] = strings.TrimSpace(outputs[i])
	}
	outputs = common.Deduplicate(outputs)
	for _, output := range outputs {
		switch output {
		case "file":
			// File writer.
			fw := &lumberjack.Logger{
				Filename:   opt.File,       // path
				MaxSize:    opt.MaxSize,    // megabytes
				MaxBackups: opt.MaxBackups, // copies
				MaxAge:     opt.MaxAge,     // days
				Compress:   opt.Compress,   // enable by default
			}

			if opt.File != "" {
				switch opt.FileFormat {
				case "json":
					// Json format.
					writerInputs = append(writerInputs, fw)
				case "raw":
					fallthrough
				default:
					// Raw format.
					writerInputs = append(writerInputs, zerolog.ConsoleWriter{
						Out:        fw,
						TimeFormat: opt.TimeFormat,
						NoColor:    opt.NoColor,
					})
				}
			}
		case "console":
			// Write to os.Stdout directly.
			writerInputs = append(writerInputs, &zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: opt.TimeFormat,
				NoColor:    opt.NoColor,
			})
		}
	}

	logger := zerolog.New(zerolog.MultiLevelWriter(writerInputs...))

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	logger = logger.With().Caller().Logger()
	if opt.TimeFormat != "" {
		logger = logger.With().Timestamp().Logger()
	}
	logger.Level(zerolog.DebugLevel)

	return &logger
}
