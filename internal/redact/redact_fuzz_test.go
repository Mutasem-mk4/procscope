package redact

import (
	"testing"
)

func FuzzRedactArg(f *testing.F) {
	config := DefaultConfig()
	f.Add("password=123")
	f.Add("/etc/shadow")
	f.Add("normal-argument")
	
	f.Fuzz(func(t *testing.T, data string) {
		res := config.Arg(data)
		if len(res) > config.MaxArgLen+3 && res != "[REDACTED]" {
			t.Errorf("Result length too long: %d", len(res))
		}
	})
}

func FuzzRedactPath(f *testing.F) {
	config := DefaultConfig()
	f.Add("/home/user/.ssh/id_rsa")
	f.Add("/var/log/auth.log")
	
	f.Fuzz(func(t *testing.T, data string) {
		res := config.Path(data)
		if len(res) > config.MaxPathLen+3 && res != "[REDACTED-PATH]" {
			t.Errorf("Path length too long: %d", len(res))
		}
	})
}
