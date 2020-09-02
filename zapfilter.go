package zapfilter

import (
	"path"
	"strings"
	"sync"

	"go.uber.org/zap/zapcore"
)

// FilterFunc is used to check whether to filter the given entry and filters out.
type FilterFunc func(zapcore.Entry, []zapcore.Field) bool

type filteringCore struct {
	zapcore.Core
	filter FilterFunc
}

// NewFilteringCore returns a core middleware that uses the given filter function
// to decide whether to actually call Write on the next core in the chain.
func NewFilteringCore(next zapcore.Core, filter FilterFunc) zapcore.Core {
	return &filteringCore{next, filter}
}

func (core *filteringCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}
	return checked
}

func (core *filteringCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if !core.filter(entry, fields) {
		return nil
	}
	return core.Core.Write(entry, fields)
}

// ByNamespace takes a list of patterns to filter out logs based on their namespaces.
// Patterns are checked using path.Match.
func ByNamespaces(namespaces string) FilterFunc {
	var mutex sync.Mutex
	matchMap := map[string]bool{}
	patterns := strings.Split(namespaces, ",")

	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		mutex.Lock()
		defer mutex.Unlock()

		if _, found := matchMap[entry.LoggerName]; !found {
			matchMap[entry.LoggerName] = false
			for _, pattern := range patterns {
				if matched, _ := path.Match(pattern, entry.LoggerName); matched {
					matchMap[entry.LoggerName] = true
					break
				}
			}
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
		for _, filter := range filters {
			if !filter(entry, fields) {
				return false
			}
		}
		return true
	}
}
