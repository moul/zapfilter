package zapfilter

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FilterFunc is used to check whether to filter the given entry and filters out.
type FilterFunc func(zapcore.Entry, []zapcore.Field) bool

// NewFilteringCore returns a core middleware that uses the given filter function to
// determine whether to actually call Write on the next core in the chain.
func NewFilteringCore(next zapcore.Core, filter FilterFunc) zapcore.Core {
	if filter == nil {
		filter = alwaysFalseFilter
	}
	return &filteringCore{next, filter}
}

// CheckAnyLevel determines whether at least one log level isn't filtered-out by the logger.
func CheckAnyLevel(logger *zap.Logger) bool {
	for _, level := range allLevels {
		if level >= zapcore.PanicLevel {
			continue // panic and fatal cannot be skipped
		}
		if logger.Check(level, "") != nil {
			return true
		}
	}
	return false
}

type filteringCore struct {
	next   zapcore.Core
	filter FilterFunc
}

// Check determines whether the supplied zapcore.Entry should be logged.
// If the entry should be logged, the filteringCore adds itself to the zapcore.CheckedEntry
// and returns the results.
func (core *filteringCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// FIXME: consider calling downstream core.Check too, but need to document how to
	// properly set logging level.
	if core.filter(entry, nil) {
		ce = ce.AddCore(entry, core)
	}
	return ce
}

// Write determines whether the supplied zapcore.Entry with provided []zapcore.Field should
// be logged, then calls the wrapped zapcore.Write.
func (core *filteringCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if !core.filter(entry, fields) {
		return nil
	}
	return core.next.Write(entry, fields)
}

// With adds structured context to the wrapped zapcore.Core.
func (core *filteringCore) With(fields []zapcore.Field) zapcore.Core {
	return &filteringCore{
		next:   core.next.With(fields),
		filter: core.filter,
	}
}

// Enabled asks the wrapped zapcore.Core to decide whether a given logging level is enabled
// when logging a message.
func (core *filteringCore) Enabled(level zapcore.Level) bool {
	// FIXME: Maybe it's better to always return true and only rely on the Check() func?
	//        Another way to consider it is to keep the smaller log level configured on
	//        zapfilter.
	return core.next.Enabled(level)
}

// Sync flushed buffered logs (if any).
func (core *filteringCore) Sync() error {
	return core.next.Sync()
}

// ByNamespaces takes a list of patterns to filter out logs based on their namespaces.
// Patterns are checked using path.Match.
func ByNamespaces(input string) FilterFunc {
	if input == "" {
		return alwaysFalseFilter
	}
	patterns := strings.Split(input, ",")

	// edge case optimization (always true)
	{
		hasIncludeWildcard := false
		hasExclude := false
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			if pattern == "*" {
				hasIncludeWildcard = true
			}
			if pattern[0] == '-' {
				hasExclude = true
			}
		}
		if hasIncludeWildcard && !hasExclude {
			return alwaysTrueFilter
		}
	}

	var mutex sync.Mutex
	matchMap := map[string]bool{}
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		mutex.Lock()
		defer mutex.Unlock()

		if _, found := matchMap[entry.LoggerName]; !found {
			matchMap[entry.LoggerName] = false
			matchInclude := false
			matchExclude := false
			for _, pattern := range patterns {
				switch {
				case pattern[0] == '-' && !matchExclude:
					if matched, _ := path.Match(pattern[1:], entry.LoggerName); matched {
						matchExclude = true
					}
				case pattern[0] != '-' && !matchInclude:
					if matched, _ := path.Match(pattern, entry.LoggerName); matched {
						matchInclude = true
					}
				}
			}
			matchMap[entry.LoggerName] = matchInclude && !matchExclude
		}
		return matchMap[entry.LoggerName]
	}
}

// ExactLevel filters out entries with an invalid level.
func ExactLevel(level zapcore.Level) FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		return entry.Level == level
	}
}

// MinimumLevel filters out entries with a too low level.
func MinimumLevel(level zapcore.Level) FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		return entry.Level >= level
	}
}

// Any checks if any filter returns true.
func Any(filters ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		for _, filter := range filters {
			if filter == nil {
				continue
			}
			if filter(entry, fields) {
				return true
			}
		}
		return false
	}
}

// Reverse checks is the passed filter returns false.
func Reverse(filter FilterFunc) FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		return !filter(entry, fields)
	}
}

// All checks if all filters return true.
func All(filters ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		var atLeastOneSuccessful bool
		for _, filter := range filters {
			if filter == nil {
				continue
			}
			if !filter(entry, fields) {
				return false
			}
			atLeastOneSuccessful = true
		}
		return atLeastOneSuccessful
	}
}

// ParseRules takes a CLI-friendly set of rules to construct a filter.
func ParseRules(input string) (FilterFunc, error) {
	var topFilter FilterFunc

	// rules are separated by spaces, tabs or \n
	for _, rule := range strings.Fields(input) {
		// split rule into parts (separated by ':')
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, ":", 2)
		var left, right string
		switch len(parts) {
		case 1:
			// if no separator, left stays empty
			right = parts[0]
		case 2:
			if parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("bad syntax")
			}
			left = parts[0]
			right = parts[1]
		default:
			return nil, fmt.Errorf("bad syntax")
		}

		// parse left part
		var (
			enabledLevels = make(map[zapcore.Level]bool)
		)
		for _, leftPart := range strings.Split(left, ",") {
			switch strings.ToLower(leftPart) {
			case "", "*", "debug+":
				enabledLevels[zapcore.DebugLevel] = true
				enabledLevels[zapcore.InfoLevel] = true
				enabledLevels[zapcore.WarnLevel] = true
				enabledLevels[zapcore.ErrorLevel] = true
				enabledLevels[zapcore.DPanicLevel] = true
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "debug":
				enabledLevels[zapcore.DebugLevel] = true
			case "info":
				enabledLevels[zapcore.InfoLevel] = true
			case "info+":
				enabledLevels[zapcore.InfoLevel] = true
				enabledLevels[zapcore.WarnLevel] = true
				enabledLevels[zapcore.ErrorLevel] = true
				enabledLevels[zapcore.DPanicLevel] = true
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "warn":
				enabledLevels[zapcore.WarnLevel] = true
			case "warn+":
				enabledLevels[zapcore.WarnLevel] = true
				enabledLevels[zapcore.ErrorLevel] = true
				enabledLevels[zapcore.DPanicLevel] = true
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "error":
				enabledLevels[zapcore.ErrorLevel] = true
			case "error+":
				enabledLevels[zapcore.ErrorLevel] = true
				enabledLevels[zapcore.DPanicLevel] = true
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "dpanic":
				enabledLevels[zapcore.DPanicLevel] = true
			case "dpanic+":
				enabledLevels[zapcore.DPanicLevel] = true
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "panic":
				enabledLevels[zapcore.PanicLevel] = true
			case "panic+":
				enabledLevels[zapcore.PanicLevel] = true
				enabledLevels[zapcore.FatalLevel] = true
			case "fatal", "fatal+":
				enabledLevels[zapcore.FatalLevel] = true
			default:
				return nil, fmt.Errorf("unsupported keyword: %q", left)
			}
		}

		// create rule's filter
		switch len(enabledLevels) {
		case 7:
			topFilter = Any(topFilter, ByNamespaces(right))
		default:
			var levelFilter FilterFunc
			for level := range enabledLevels {
				levelFilter = Any(ExactLevel(level), levelFilter)
			}
			topFilter = Any(topFilter, All(levelFilter, ByNamespaces(right)))
		}
	}

	return topFilter, nil
}

func MustParseRules(input string) FilterFunc {
	filter, err := ParseRules(input)
	if err != nil {
		panic(err)
	}
	return filter
}

func alwaysFalseFilter(_ zapcore.Entry, _ []zapcore.Field) bool {
	return false
}

func alwaysTrueFilter(_ zapcore.Entry, _ []zapcore.Field) bool {
	return true
}

var allLevels = []zapcore.Level{
	zapcore.DebugLevel,
	zapcore.InfoLevel,
	zapcore.WarnLevel,
	zapcore.ErrorLevel,
	zapcore.DPanicLevel,
	zapcore.PanicLevel,
	zapcore.FatalLevel,
}
