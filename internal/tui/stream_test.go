package tui_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// eventStream builds a newline-delimited JSON event stream from the given events.
func eventStream(events ...map[string]any) *strings.Reader {
	var buf strings.Builder
	for _, ev := range events {
		data, _ := json.Marshal(ev)
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return strings.NewReader(buf.String())
}

func minimalDeployEvents() []map[string]any {
	return []map[string]any{
		{"event": "start", "step": "extract", "artifact": "/tmp/app.tar.gz"},
		{"event": "done", "step": "extract", "duration_ms": float64(500)},
		{"event": "done", "step": "link", "dirs": []any{"var/log"}, "files": []any{".env"}, "duration_ms": float64(10)},
		{"event": "done", "step": "current_symlink", "path": "/var/www/releases/20260606-120000", "duration_ms": float64(5)},
		{"event": "done", "step": "purge", "purged": []any{"20260101-000000"}, "kept": float64(5), "duration_ms": float64(2)},
		{"event": "done", "step": "deploy", "release": "20260606-120000", "duration_ms": float64(517)},
	}
}

func TestForwardStream_JSONMode_AddsServerField(t *testing.T) {
	events := minimalDeployEvents()
	r := eventStream(events...)
	var buf bytes.Buffer

	release, err := tui.ForwardStream(r, "web-01", forgeui.JSON, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release != "20260606-120000" {
		t.Errorf("release: got %q, want 20260606-120000", release)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(events) {
		t.Fatalf("line count: got %d, want %d", len(lines), len(events))
	}
	for i, line := range lines {
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}
		if ev["server"] != "web-01" {
			t.Errorf("line %d: server field: got %q, want web-01", i, ev["server"])
		}
	}
}

func TestForwardStream_JSONMode_PreservesOriginalFields(t *testing.T) {
	r := eventStream(
		map[string]any{"event": "done", "step": "extract", "duration_ms": float64(500)},
		map[string]any{"event": "done", "step": "deploy", "release": "20260606-120000", "duration_ms": float64(500)},
	)
	var buf bytes.Buffer
	_, _ = tui.ForwardStream(r, "web-01", forgeui.JSON, &buf)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var ev map[string]any
	_ = json.Unmarshal([]byte(lines[0]), &ev)

	if ev["event"] != "done" {
		t.Errorf("event: got %q, want done", ev["event"])
	}
	if ev["step"] != "extract" {
		t.Errorf("step: got %q, want extract", ev["step"])
	}
}

func TestForwardStream_PlainMode_RendersSteps(t *testing.T) {
	r := eventStream(minimalDeployEvents()...)
	var buf bytes.Buffer

	_, err := tui.ForwardStream(r, "web-01", forgeui.Plain, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "extracted") {
		t.Errorf("expected 'extracted' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "symlink") {
		t.Errorf("expected 'symlink' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Deployed") {
		t.Errorf("expected 'Deployed' in output, got:\n%s", out)
	}
}

func TestForwardStream_PlainMode_SharedDetails(t *testing.T) {
	r := eventStream(minimalDeployEvents()...)
	var buf bytes.Buffer
	_, _ = tui.ForwardStream(r, "web-01", forgeui.Plain, &buf)

	out := buf.String()
	if !strings.Contains(out, "var/log") {
		t.Errorf("expected shared dir 'var/log' in output, got:\n%s", out)
	}
	if !strings.Contains(out, ".env") {
		t.Errorf("expected shared file '.env' in output, got:\n%s", out)
	}
}

func TestForwardStream_ErrorEvent_ReturnsError(t *testing.T) {
	r := eventStream(
		map[string]any{"event": "start", "step": "extract"},
		map[string]any{"event": "error", "step": "extract", "message": "archive is corrupt", "exit_code": 1},
	)
	var buf bytes.Buffer

	_, err := tui.ForwardStream(r, "web-01", forgeui.Plain, &buf)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "archive is corrupt") {
		t.Errorf("error message: got %q, want to contain 'archive is corrupt'", err.Error())
	}
}

func TestForwardStream_DeployDone_ReturnsRelease(t *testing.T) {
	r := eventStream(
		map[string]any{"event": "done", "step": "deploy", "release": "20260606-150000", "duration_ms": float64(1000)},
	)
	var buf bytes.Buffer

	release, err := tui.ForwardStream(r, "", forgeui.Plain, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release != "20260606-150000" {
		t.Errorf("release: got %q, want 20260606-150000", release)
	}
}

func TestForwardStream_PlainMode_RendersHookOutput(t *testing.T) {
	r := eventStream(
		map[string]any{"event": "hook_output", "output": "Migrating: create_users_table\nMigrated\n"},
		map[string]any{"event": "done", "step": "deploy", "release": "r1", "duration_ms": float64(10)},
	)
	var buf bytes.Buffer

	_, err := tui.ForwardStream(r, "web-01", forgeui.Plain, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out := buf.String(); !strings.Contains(out, "Migrating: create_users_table") {
		t.Errorf("expected hook output in rendered stream, got:\n%s", out)
	}
}

func TestForwardStream_HandlesLargeEventLine(t *testing.T) {
	big := strings.Repeat("x", 100*1024) // exceeds bufio.Scanner's 64 KB default token
	r := eventStream(
		map[string]any{"event": "hook_output", "output": big},
		map[string]any{"event": "done", "step": "deploy", "release": "r1", "duration_ms": float64(1)},
	)
	var buf bytes.Buffer

	release, err := tui.ForwardStream(r, "web-01", forgeui.Plain, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release != "r1" {
		t.Errorf("release: got %q, want r1 — the stream was truncated before the done event", release)
	}
	if !strings.Contains(buf.String(), big) {
		t.Error("large hook output was not rendered — the line was truncated")
	}
}

func TestForwardStream_EmptyStream(t *testing.T) {
	r := strings.NewReader("")
	var buf bytes.Buffer

	release, err := tui.ForwardStream(r, "", forgeui.Plain, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release != "" {
		t.Errorf("release: got %q, want empty", release)
	}
}
