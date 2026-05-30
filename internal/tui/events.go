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
