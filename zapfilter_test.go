package zapfilter_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"moul.io/zapfilter"
)

func Example() {

}

func ExampleNewFilteringCore_wrap() {
	filtered := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapfilter.NewFilteringCore(c, zapfilter.ByNamespaces("demo*"))
	})

	logger := zap.NewExample()
	defer logger.Sync()

	logger.WithOptions(filtered).Debug("hello world!")
	logger.WithOptions(filtered).Named("demo").Debug("hello earth!")
	logger.WithOptions(filtered).Named("other").Debug("hello universe!")

	// Output:
	// {"level":"debug","logger":"demo","msg":"hello earth!"}
}

func ExampleNewFilteringCore_newlogger() {
	c := zap.NewExample().Core()

	logger := zap.New(zapfilter.NewFilteringCore(c, zapfilter.ByNamespaces("demo*")))
	defer logger.Sync()

	logger.Debug("hello world!")
	logger.Named("demo").Debug("hello earth!")
	logger.Named("other").Debug("hello universe!")

	// Output:
	// {"level":"debug","logger":"demo","msg":"hello earth!"}
}

func ExampleFilterFunc_custom() {
	rand.Seed(42)

	core := zap.NewExample().Core()
	filterFunc := func(entry zapcore.Entry, fields []zapcore.Field) bool {
		return rand.Intn(2) == 1
	}
	logger := zap.New(zapfilter.NewFilteringCore(core, filterFunc))
	defer logger.Sync()

	logger.Debug("hello city!")
	logger.Debug("hello region!")
	logger.Debug("hello planet!")
	logger.Debug("hello solar system!")
	logger.Debug("hello universe!")
	logger.Debug("hello multiverse!")

	// Output:
	// {"level":"debug","msg":"hello city!"}
	// {"level":"debug","msg":"hello region!"}
	// {"level":"debug","msg":"hello universe!"}
	// {"level":"debug","msg":"hello multiverse!"}
}

func ExampleByNamespaces() {
	core := zap.NewExample().Core()
	logger := zap.New(zapfilter.NewFilteringCore(core, zapfilter.ByNamespaces("demo1.*,demo3.*")))
	defer logger.Sync()

	logger.Debug("hello city!")
	logger.Named("demo1.frontend").Debug("hello region!")
	logger.Named("demo2.frontend").Debug("hello planet!")
	logger.Named("demo3.frontend").Debug("hello solar system!")

	// Output:
	// {"level":"debug","logger":"demo1.frontend","msg":"hello region!"}
	// {"level":"debug","logger":"demo3.frontend","msg":"hello solar system!"}
}

func TestFilterFunc(t *testing.T) {
	cases := []struct {
		name         string
		filterFunc   zapfilter.FilterFunc
		expectedLogs []string
	}{
		{
			"allow-all",
			func(entry zapcore.Entry, fields []zapcore.Field) bool {
				return true
			},
			[]string{"a", "b", "c", "d"},
		}, {
			"disallow-all",
			func(entry zapcore.Entry, fields []zapcore.Field) bool {
				return false
			},
			[]string{},
		}, {
			"minimum-debug",
			zapfilter.MinimumLevel(zapcore.DebugLevel),
			[]string{"a", "b", "c", "d"},
		}, {
			"minimum-info",
			zapfilter.MinimumLevel(zapcore.InfoLevel),
			[]string{"b", "c", "d"},
		}, {
			"minimum-warn",
			zapfilter.MinimumLevel(zapcore.WarnLevel),
			[]string{"c", "d"},
		}, {
			"minimum-error",
			zapfilter.MinimumLevel(zapcore.ErrorLevel),
			[]string{"d"},
		}, {
			"exact-debug",
			zapfilter.ExactLevel(zapcore.DebugLevel),
			[]string{"a"},
		}, {
			"exact-info",
			zapfilter.ExactLevel(zapcore.InfoLevel),
			[]string{"b"},
		}, {
			"exact-warn",
			zapfilter.ExactLevel(zapcore.WarnLevel),
			[]string{"c"},
		}, {
			"exact-error",
			zapfilter.ExactLevel(zapcore.ErrorLevel),
			[]string{"d"},
		}, {
			"all-except-debug",
			zapfilter.Reverse(zapfilter.ExactLevel(zapcore.DebugLevel)),
			[]string{"b", "c", "d"},
		}, {
			"all-except-info",
			zapfilter.Reverse(zapfilter.ExactLevel(zapcore.InfoLevel)),
			[]string{"a", "c", "d"},
		}, {
			"all-except-warn",
			zapfilter.Reverse(zapfilter.ExactLevel(zapcore.WarnLevel)),
			[]string{"a", "b", "d"},
		}, {
			"all-except-error",
			zapfilter.Reverse(zapfilter.ExactLevel(zapcore.ErrorLevel)),
			[]string{"a", "b", "c"},
		}, {
			"any",
			zapfilter.Any(
				zapfilter.ExactLevel(zapcore.DebugLevel),
				zapfilter.ExactLevel(zapcore.WarnLevel),
			),
			[]string{"a", "c"},
		}, {
			"all-1",
			zapfilter.All(
				zapfilter.ExactLevel(zapcore.DebugLevel),
				zapfilter.ExactLevel(zapcore.WarnLevel),
			),
			[]string{},
		}, {
			"all-2",
			zapfilter.All(
				zapfilter.ExactLevel(zapcore.DebugLevel),
				zapfilter.ExactLevel(zapcore.DebugLevel),
			),
			[]string{"a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next, logs := observer.New(zapcore.DebugLevel)
			core := zapfilter.NewFilteringCore(next, tc.filterFunc)
			logger := zap.New(core)

			logger.Debug("a")
			logger.Info("b")
			logger.Warn("c")
			logger.Error("d")

			gotLogs := []string{}
			for _, log := range logs.All() {
				gotLogs = append(gotLogs, log.Message)
			}

			require.Equal(t, gotLogs, tc.expectedLogs)
		})
	}
}
