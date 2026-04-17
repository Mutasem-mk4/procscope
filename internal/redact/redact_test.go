package redact

import (
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.MaxArgLen != 1024 {
		t.Errorf("MaxArgLen = %d, want 1024", c.MaxArgLen)
	}
	if c.MaxPathLen != 4096 {
		t.Errorf("MaxPathLen = %d, want 4096", c.MaxPathLen)
	}
	if c.ShowEnv {
		t.Error("ShowEnv should default to false")
	}
}

func TestRedactPath(t *testing.T) {
	c := DefaultConfig()

	// Normal path — no redaction
	got := c.Path("/usr/bin/test")
	if got != "/usr/bin/test" {
		t.Errorf("Path = %s, want /usr/bin/test", got)
	}

	// Path containing sensitive pattern
	got = c.Path("/etc/shadow-password-backup")
	if got != "[REDACTED-PATH]" {
		t.Errorf("Path with sensitive pattern = %s, want [REDACTED-PATH]", got)
	}

	// Long path truncation
	longPath := strings.Repeat("a", 5000)
	got = c.Path(longPath)
	if len(got) > c.MaxPathLen+3 { // +3 for "..."
		t.Errorf("long path not truncated: len=%d", len(got))
	}
}

func TestRedactArg(t *testing.T) {
	c := DefaultConfig()

	// Normal arg
	if got := c.Arg("--verbose"); got != "--verbose" {
		t.Errorf("Arg = %s, want --verbose", got)
	}

	// Sensitive arg
	if got := c.Arg("--api_key=abc123"); got != "[REDACTED]" {
		t.Errorf("sensitive Arg = %s, want [REDACTED]", got)
	}

	// Case insensitive
	if got := c.Arg("--API_KEY=abc123"); got != "[REDACTED]" {
		t.Errorf("case-insensitive sensitive Arg = %s, want [REDACTED]", got)
	}
}

func TestRedactArgs(t *testing.T) {
	c := DefaultConfig()
	c.MaxArgs = 3

	args := []string{"cmd", "--flag", "value", "extra1", "extra2"}
	got := c.Args(args)

	if len(got) != 3 {
		t.Errorf("Args count = %d, want 3 (maxArgs)", len(got))
	}
}

func TestRedactArgSensitivePatterns(t *testing.T) {
	c := DefaultConfig()

	patterns := []string{
		"password=secret123",
		"my_secret_value",
		"Authorization: Bearer xyz",
		"token=abc",
		"private_key_content",
		"credential_data",
	}

	for _, p := range patterns {
		got := c.Arg(p)
		if got != "[REDACTED]" {
			t.Errorf("Arg(%q) = %q, want [REDACTED]", p, got)
		}
	}
}

func TestRedactArgSafeValues(t *testing.T) {
	c := DefaultConfig()

	safe := []string{
		"--verbose",
		"/usr/bin/test",
		"hello world",
		"--port=8080",
		"--output=/tmp/result.json",
	}

	for _, s := range safe {
		got := c.Arg(s)
		if got != s {
			t.Errorf("Arg(%q) = %q, should not be redacted", s, got)
		}
	}
}
