package zerologr

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	t.Parallel()

	VerbosityFieldName = ""

	tests := []struct {
		name         string
		zerologLevel zerolog.Level
		logFunc      func(log logr.Logger)
		formatter    func(interface{}) interface{}
		reportCaller bool
		defaultName  []string
		assertions   map[string]string
	}{
		{
			name: "basic logging",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			name: "set name once",
			logFunc: func(log logr.Logger) {
				log.WithName("main").Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main",
			},
		},
		{
			name: "set name twice",
			logFunc: func(log logr.Logger) {
				log.WithName("main").WithName("subpackage").Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main/subpackage",
			},
		},
		{
			name: "set name and values and name again",
			logFunc: func(log logr.Logger) {
				log.
					WithName("main").
					WithValues("k1", "v1", "k2", "v2").
					WithName("subpackage").
					Info("hello, world", "k3", "v3")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main/subpackage",
				"k1":      "v1",
				"k2":      "v2",
				"k3":      "v3",
			},
		},
		{
			name: "V(0) logging with info level set is shown",
			logFunc: func(log logr.Logger) {
				log.V(0).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			name: "V(2) logging with info level set is not shown",
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
				log.V(2).Info("hello, world")
			},
			assertions: nil,
		},
		{
			name:         "V(1) logging with debug level set is shown",
			zerologLevel: zerolog.DebugLevel,
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "debug",
				"message": "hello, world",
			},
		},
		{
			name:         "V(2) logging with trace level set is shown",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(2).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "trace",
				"message": "hello, world",
			},
		},
		{
			name:         "negative V-logging truncates to info",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(-10).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			name:         "additive V-logging, negatives ignored",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(0).V(1).V(-20).V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "trace",
				"message": "hello, world",
			},
		},
		{
			name: "arguments are added while calling Info()",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"animal":  "walrus",
			},
		},
		{
			name: "arguments are added after WithValues()",
			logFunc: func(log logr.Logger) {
				log.WithValues("color", "green").Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"animal":  "walrus",
				"color":   "green",
			},
		},
		{
			name: "error logs have the appropriate information",
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level":   "error",
				"message": "error occurred",
				"error":   "this is error",
			},
		},
		{
			name: "error shown with low severity logger",
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level":   "error",
				"message": "error occurred",
				"error":   "this is error",
			},
		},
		{
			name: "bad number of arguments discards all",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus", "foo")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"-animal": "walrus",
			},
		},
		{
			name: "with default name",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			defaultName: []string{"some", "name"},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "some/name",
			},
		},
		{
			name: "without report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"-caller": "no-caller",
			},
		},
		{
			name: "with report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"caller":  `~zerologr_test.go:\d+`,
			},
		},
		{
			name: "with report caller and depth",
			logFunc: func(log logr.Logger) {
				log.WithCallDepth(2).Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"caller":  `~testing.go:\d+`,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use a buffer for our output.
			logWriter := &bytes.Buffer{}

			zerologLogger := zerolog.New(logWriter)

			if tc.zerologLevel != zerolog.PanicLevel {
				zerologLogger.Level(tc.zerologLevel)
			}

			// Send the created logger to the test case to invoke desired
			// logging.
			if tc.reportCaller {
				zerologLogger = zerologLogger.With().Caller().Logger()
			}

			if tc.assertions == nil {
				assert.Equal(t, logWriter.Len(), 0)
				return
			}

			logger := New(&zerologLogger)

			if tc.defaultName != nil {
				logger = logger.WithName(strings.Join(tc.defaultName, NameSeparator))
			}

			tc.logFunc(logger)

			var loggedLine map[string]string
			b := logWriter.Bytes()
			err := json.Unmarshal(b, &loggedLine)

			require.NoError(t, err)

			for k, v := range tc.assertions {
				field, ok := loggedLine[k]

				// Annotate negative tests with a minus. To ensure `key` is
				// *not* in the output, name the assertion `-key`.
				if strings.HasPrefix(k, "-") {
					assert.False(t, ok)
					assert.Empty(t, field)

					continue
				}

				// Annotate regexp matches with the value starting with a tilde
				// (~). The tilde will be dropped and used to compile a regexp to
				// match the field.
				if strings.HasPrefix(v, "~") {
					assert.Regexp(t, regexp.MustCompile(v[1:]), field)
					continue
				}

				assert.True(t, ok)
				assert.NotEmpty(t, field)
				assert.Equal(t, v, field)
			}
		})
	}
}

func TestLogSink_Enabled(t *testing.T) {
	t.Parallel()

	type fields struct {
		verbosity zerolog.Level
	}

	type args struct {
		level int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "output level 0 is enabled with verbosity info",
			fields: fields{verbosity: zerolog.InfoLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is not enabled with verbosity info",
			fields: fields{verbosity: zerolog.InfoLevel},
			args:   args{level: 1},
			want:   false,
		},
		{
			name:   "output level 0 is enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 1},
			want:   true,
		},
		{
			name:   "output level 2 is not enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 2},
			want:   false,
		},
		{
			name:   "output level 0 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 1},
			want:   true,
		},
		{
			name:   "output level 2 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 2},
			want:   true,
		},
		{
			name:   "output level 3 is not enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 3},
			want:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.New(nil).Level(tc.fields.verbosity)

			ls := &LogSink{
				l: &logger,
			}

			assert.Equalf(t, tc.want, ls.Enabled(tc.args.level), "Enabled(%v)", tc.args.level)
		})
	}
}
