package tui_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
)

func TestHookOutputWriter_EmitsHookOutputEvent(t *testing.T) {
	var buf bytes.Buffer
	w := tui.NewHookOutputWriter(tui.NewJSONEmitter(&buf))

	const text = "running migrations\n"
	n, err := w.Write([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(text) {
		t.Errorf("n: got %d, want %d", n, len(text))
	}

	var ev map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &ev); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if ev["event"] != "hook_output" {
		t.Errorf("event: got %q, want hook_output", ev["event"])
	}
	if ev["output"] != text {
		t.Errorf("output: got %q, want %q", ev["output"], text)
	}
}
