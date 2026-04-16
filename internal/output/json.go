package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/procscope/procscope/internal/events"
)

// JSONWriter writes events as JSON or JSONL (newline-delimited JSON).
type JSONWriter struct {
	mu     sync.Mutex
	writer io.Writer
	file   *os.File // non-nil if we own the file
	count  uint64
}

// NewJSONWriter creates a JSONL writer to the given file path.
// Use "-" for stdout.
func NewJSONWriter(path string) (*JSONWriter, error) {
	if path == "-" || path == "" {
		return &JSONWriter{writer: os.Stdout}, nil
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSONL output file: %w", err)
	}

	return &JSONWriter{
		writer: f,
		file:   f,
	}, nil
}

// WriteEvent writes a single event as a JSONL line.
func (j *JSONWriter) WriteEvent(evt *events.Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	if _, err := j.writer.Write(data); err != nil {
		return err
	}
	if _, err := j.writer.Write([]byte("\n")); err != nil {
		return err
	}

	j.count++
	return nil
}

// Count returns the number of events written.
func (j *JSONWriter) Count() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.count
}

// Close closes the underlying file if we own it.
func (j *JSONWriter) Close() error {
	if j.file != nil {
		return j.file.Close()
	}
	return nil
}

// WriteEventsArray writes a slice of events as a JSON array to a writer.
// Used for evidence bundle output.
func WriteEventsArray(w io.Writer, evts []*events.Event) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(evts)
}
