package zapfilter_test

import (
	"fmt"
	"math/rand"
	"strings"
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

func ExampleParseRules() {
	core := zap.NewExample().Core()
	// *=myns             => any level, myns namespace
	// info,warn:myns.*   => info or warn level, any namespace matching myns.*
	// error=*            => everything with error level
	logger := zap.New(zapfilter.NewFilteringCore(core, zapfilter.MustParseRules("*:myns info,warn:myns.* error:*")))
	defer logger.Sync()

	logger.Debug("top debug")                                 // no match
	logger.Named("myns").Debug("myns debug")                  // matches *:myns
	logger.Named("bar").Debug("bar debug")                    // no match
	logger.Named("myns").Named("foo").Debug("myns.foo debug") // no match

	logger.Info("top info")                                 // no match
	logger.Named("myns").Info("myns info")                  // matches *:myns
	logger.Named("bar").Info("bar info")                    // no match
	logger.Named("myns").Named("foo").Info("myns.foo info") // matches info,warn:myns.*

	logger.Warn("top warn")                                 // no match
	logger.Named("myns").Warn("myns warn")                  // matches *:myns
	logger.Named("bar").Warn("bar warn")                    // no match
	logger.Named("myns").Named("foo").Warn("myns.foo warn") // matches info,warn:myns.*

	logger.Error("top error")                                 // matches error:*
	logger.Named("myns").Error("myns error")                  // matches *:myns and error:*
	logger.Named("bar").Error("bar error")                    // matches error:*
	logger.Named("myns").Named("foo").Error("myns.foo error") // matches error:*

	// Output:
	// {"level":"debug","logger":"myns","msg":"myns debug"}
	// {"level":"info","logger":"myns","msg":"myns info"}
	// {"level":"info","logger":"myns.foo","msg":"myns.foo info"}
	// {"level":"warn","logger":"myns","msg":"myns warn"}
	// {"level":"warn","logger":"myns.foo","msg":"myns.foo warn"}
	// {"level":"error","msg":"top error"}
	// {"level":"error","logger":"myns","msg":"myns error"}
	// {"level":"error","logger":"bar","msg":"bar error"}
	// {"level":"error","logger":"myns.foo","msg":"myns.foo error"}
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

func TestParseRules(t *testing.T) {
	const (
		allDebug   = "aeimquy2"
		allInfo    = "bfjnrvz3"
		allWarn    = "cgkosw04"
		allError   = "dhlptx15"
		everything = "abcdefghijklmnopqrstuvwxyz012345"
	)

	cases := []struct {
		name          string
		input         string
		expectedLogs  string
		expectedError error
	}{
		{"empty", "", "", nil},
		{"everything", "*", everything, nil},
		{"debug+", "debug+:*", everything, nil},
		{"all-debug", "debug:*", allDebug, nil},
		{"all-info", "info:*", allInfo, nil},
		{"all-warn", "warn:*", allWarn, nil},
		{"all-error", "error:*", allError, nil},
		{"all-info-and-warn-1", "info,warn:*", "bcfgjknorsvwz034", nil},
		{"all-info-and-warn-2", "info:* warn:*", "bcfgjknorsvwz034", nil},
		{"warn+", "warn+:*", "cdghklopstwx0145", nil},
		{"redundant-1", "info,info:* info:*", allInfo, nil},
		{"redundant-2", "* *:* info:*", everything, nil},
		{"foo-ns", "foo", "efgh", nil},
		{"foo-ns-wildcard", "*:foo", "efgh", nil},
		{"foo-ns-debug,info", "debug,info:foo", "ef", nil},
		{"foo.star-ns", "foo.*", "qrstuvwx", nil},
		{"foo.star-ns-wildcard", "*:foo.*", "qrstuvwx", nil},
		{"foo.star-ns-debug,info", "debug,info:foo.*", "qruv", nil},
		{"all-in-one", "*:foo debug:foo.* info,warn:bar error:*", "defghjklpqtux15", nil},
		{"invalid-left", "invalid:*", "", fmt.Errorf(`unsupported keyword: "invalid"`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next, logs := observer.New(zapcore.DebugLevel)
			filter, err := zapfilter.ParseRules(tc.input)
			require.Equal(t, tc.expectedError, err)
			if err != nil {
				return
			}

			core := zapfilter.NewFilteringCore(next, filter)
			logger := zap.New(core)

			logger.Debug("a")
			logger.Info("b")
			logger.Warn("c")
			logger.Error("d")

			logger.Named("foo").Debug("e")
			logger.Named("foo").Info("f")
			logger.Named("foo").Warn("g")
			logger.Named("foo").Error("h")

			logger.Named("bar").Debug("i")
			logger.Named("bar").Info("j")
			logger.Named("bar").Warn("k")
			logger.Named("bar").Error("l")

			logger.Named("baz").Debug("m")
			logger.Named("baz").Info("n")
			logger.Named("baz").Warn("o")
			logger.Named("baz").Error("p")

			logger.Named("foo").Named("bar").Debug("q")
			logger.Named("foo").Named("bar").Info("r")
			logger.Named("foo").Named("bar").Warn("s")
			logger.Named("foo").Named("bar").Error("t")

			logger.Named("foo").Named("foo").Debug("u")
			logger.Named("foo").Named("foo").Info("v")
			logger.Named("foo").Named("foo").Warn("w")
			logger.Named("foo").Named("foo").Error("x")

			logger.Named("bar").Named("foo").Debug("y")
			logger.Named("bar").Named("foo").Info("z")
			logger.Named("bar").Named("foo").Warn("0")
			logger.Named("bar").Named("foo").Error("1")

			logger.Named("qux").Named("foo").Debug("2")
			logger.Named("qux").Named("foo").Info("3")
			logger.Named("qux").Named("foo").Warn("4")
			logger.Named("qux").Named("foo").Error("5")

			gotLogs := []string{}
			for _, log := range logs.All() {
				gotLogs = append(gotLogs, log.Message)
			}

			expectedLogs := strings.Split(tc.expectedLogs, "")
			require.Equal(t, expectedLogs, gotLogs)
		})
	}
}
