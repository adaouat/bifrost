package tui

import (
	"encoding/json"
	"io"
)

// JSONEmitter writes newline-delimited JSON events to w.
type JSONEmitter struct {
	w io.Writer
}

// NewJSONEmitter returns a JSONEmitter writing to w.
func NewJSONEmitter(w io.Writer) *JSONEmitter {
	return &JSONEmitter{w: w}
}

// Emit serialises ev as a single JSON line.
func (e *JSONEmitter) Emit(ev map[string]any) {
	data, _ := json.Marshal(ev)
	_, _ = e.w.Write(append(data, '\n'))
}

// HookOutputWriter turns each Write into a "hook_output" JSON event via emit, so
// hook stdout/stderr never corrupts the newline-delimited event stream that the
// client parses. The deployer uses it as the hook writer in JSON mode.
type HookOutputWriter struct {
	emit *JSONEmitter
}

// NewHookOutputWriter returns a writer that emits hook output as JSON events.
func NewHookOutputWriter(emit *JSONEmitter) *HookOutputWriter {
	return &HookOutputWriter{emit: emit}
}

// Write emits p as a single hook_output event and reports all bytes consumed.
func (w *HookOutputWriter) Write(p []byte) (int, error) {
	w.emit.Emit(map[string]any{"event": "hook_output", "output": string(p)})
	return len(p), nil
}
