package output

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Mutasem-mk4/procscope/internal/events"
)

func TestBundleEventWriteFailureIsReturned(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux /dev/full to reproduce an out-of-space write")
	}
	dir := t.TempDir()
	if err := os.Symlink("/dev/full", filepath.Join(dir, "events.jsonl")); err != nil {
		t.Fatal(err)
	}
	bundle := Bundle{Dir: dir, Events: []*events.Event{{PID: 123}}}
	if err := bundle.writeEventsJSONL(); err == nil {
		t.Fatal("out-of-space write was reported as successful")
	}
}
