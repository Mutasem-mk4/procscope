// Package redact provides safe-default redaction controls for procscope output.
//
// By default, procscope:
//   - Does NOT dump environment variables
//   - Does NOT capture payload/body content
//   - Bounds argument/path lengths
//   - Allows opt-in to verbose modes
package redact

import (
	"strings"
)

// Config controls what data is redacted or truncated in output.
type Config struct {
	// MaxArgLen is the maximum length of a single command-line argument.
	// Arguments longer than this are truncated with "...".
	// Default: 1024
	MaxArgLen int

	// MaxPathLen is the maximum length of a file path.
	// Default: 4096
	MaxPathLen int

	// MaxArgs is the maximum number of argv elements to retain.
	// Default: 64
	MaxArgs int

	// ShowEnv enables environment variable output (disabled by default).
	ShowEnv bool

	// SensitivePatterns are substrings that trigger redaction in paths and args.
	// When matched, the value is replaced with "[REDACTED]".
	SensitivePatterns []string
}

// DefaultConfig returns the default redaction configuration.
// Safe defaults: no env, bounded lengths, common sensitive patterns.
func DefaultConfig() *Config {
	return &Config{
		MaxArgLen:  1024,
		MaxPathLen: 4096,
		MaxArgs:    64,
		ShowEnv:    false,
		SensitivePatterns: []string{
			"password",
			"passwd",
			"secret",
			"token",
			"api_key",
			"apikey",
			"api-key",
			"authorization",
			"credential",
			"private_key",
			"private-key",
		},
	}
}

// Path applies redaction rules to a file path.
func (c *Config) Path(path string) string {
	if len(path) > c.MaxPathLen {
		path = path[:c.MaxPathLen] + "..."
	}
	if c.containsSensitive(path) {
		return "[REDACTED-PATH]"
	}
	return path
}

// Arg applies redaction rules to a single command-line argument.
func (c *Config) Arg(arg string) string {
	if len(arg) > c.MaxArgLen {
		arg = arg[:c.MaxArgLen] + "..."
	}
	if c.containsSensitive(arg) {
		return "[REDACTED]"
	}
	return arg
}

// Args applies redaction rules to a slice of arguments.
func (c *Config) Args(args []string) []string {
	if len(args) > c.MaxArgs {
		args = args[:c.MaxArgs]
	}
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = c.Arg(arg)
	}
	return result
}

// containsSensitive checks if a string contains any sensitive patterns.
func (c *Config) containsSensitive(s string) bool {
	lower := strings.ToLower(s)
	for _, pattern := range c.SensitivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
