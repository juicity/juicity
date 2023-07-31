package log

import (
	"testing"
	"time"
)

type testCase struct {
	condition string
	logger    *Logger
}

func TestLogger(t *testing.T) {
	// construct test cases
	cases := []testCase{
		{
			condition: "ConsoleWriter (Alone)",
			logger: NewLogger(&Options{
				TimeFormat: time.DateTime,
			}),
		},
		{
			condition: "JsonWriter (Alone)",
			logger: NewLogger(&Options{
				TimeFormat:    time.DateTime,
				JsonLogFormat: true,
			}),
		},
		{
			condition: "JsonWriter (Alone)",
			logger: NewLogger(&Options{
				TimeFormat:    time.DateTime,
				JsonLogFormat: true,
			}),
		},
		{
			condition: "ConsoleWriter + FileWriter (in Stdout format)",
			logger: NewLogger(&Options{
				TimeFormat: time.DateTime,
				LogFile:    "../../app_file_writer_stdout.log",
			}),
		},
		{
			condition: "ConsoleWriter + FileWriter (in Stdout format; disable Color Output)",
			logger: NewLogger(&Options{
				TimeFormat: time.DateTime,
				LogFile:    "../../app_file_writer_stdout.log",
				NoColor:    true,
			}),
		},
		{
			condition: "JsonWrtier + FileWriter (in JSON format)",
			logger: NewLogger(&Options{
				TimeFormat:    time.DateTime,
				LogFile:       "../../app_file_writer_json.log",
				JsonLogFormat: true,
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.condition, func(t *testing.T) {
			tc.logger.Info().Msg("Hello World!")
		})
	}
}
