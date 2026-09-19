package tui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/adaouat/bifrost/internal/cmderr"
	forgeui "github.com/adaouat/forge/ui"
)

// ForwardStream reads a newline-delimited JSON event stream from r and renders
// or re-emits events to out. In json mode each event is re-emitted with the
// "server" field set to serverName. In human/plain mode events are rendered as
// TUI text. Returns the release name from the final deploy done event and any
// error, including errors reported by an error event in the stream.
func ForwardStream(r io.Reader, serverName string, mode forgeui.Mode, out io.Writer) (string, error) {
	// A hook's entire stdout is forwarded as a single line, which can be
	// arbitrarily large; bufio.Reader.ReadBytes has no fixed token-size cap
	// (unlike bufio.Scanner), so a verbose hook can never fail an otherwise
	// successful deploy here.
	reader := bufio.NewReaderSize(r, 64*1024)
	emit := NewJSONEmitter(out)
	sp := forgeui.NewSpinner(out, mode)

	var release string

	for {
		raw, readErr := reader.ReadBytes('\n')
		line := bytes.TrimRight(raw, "\n")

		var ev map[string]any
		if len(line) > 0 {
			if err := json.Unmarshal(line, &ev); err != nil {
				ev = nil
			}
		}

		if ev != nil {
			if mode == forgeui.JSON {
				if ev["event"] == "done" && ev["step"] == "deploy" {
					release, _ = ev["release"].(string)
				}
				ev["server"] = serverName
				emit.Emit(ev)
			} else {
				event, _ := ev["event"].(string)
				step, _ := ev["step"].(string)
				ms, _ := ev["duration_ms"].(float64)
				detail := fmt.Sprintf("(%.1fs)", ms/1000)

				switch event {
				case "error":
					msg, _ := ev["message"].(string)
					eventErr := fmt.Errorf("%s: %s", step, msg)
					if code, ok := ev["exit_code"].(float64); ok && code != 0 {
						eventErr = cmderr.Wrap(int(code), eventErr)
					}
					return "", eventErr

				case "done":
					switch step {
					case "extract":
						sp.Step("Artifact extracted", detail)

					case "link":
						dirs, _ := ev["dirs"].([]any)
						files, _ := ev["files"].([]any)
						sp.Step(fmt.Sprintf("Shared dirs linked (%d)", len(dirs)), "")
						for _, d := range dirs {
							PrintDetail(mode, out, fmt.Sprintf("%v", d))
						}
						sp.Step(fmt.Sprintf("Shared files linked (%d)", len(files)), "")
						for _, f := range files {
							PrintDetail(mode, out, fmt.Sprintf("%v", f))
						}

					case "current_symlink":
						path, _ := ev["path"].(string)
						sp.Step("Current symlink updated", "")
						if path != "" {
							PrintDetail(mode, out, path)
						}

					case "purge":
						kept, _ := ev["kept"].(float64)
						sp.Step(fmt.Sprintf("Old releases purged (kept %d)", int(kept)), "")
						purged, _ := ev["purged"].([]any)
						for _, p := range purged {
							PrintDetail(mode, out, fmt.Sprintf("%v deleted", p))
						}

					case "deploy":
						release, _ = ev["release"].(string)
						elapsed := time.Duration(ms) * time.Millisecond
						PrintSummary(mode, out, elapsed, release)
					}

				case "hook":
					cmd, _ := ev["cmd"].(string)
					exitCode, _ := ev["exit_code"].(float64)
					if exitCode != 0 {
						lifecycle, _ := ev["lifecycle"].(string)
						PrintDetail(mode, out, fmt.Sprintf("[FAILED] %s: %s", lifecycle, cmd))
					} else {
						PrintDetail(mode, out, cmd)
					}

				case "hook_output":
					if output, _ := ev["output"].(string); output != "" {
						_, _ = io.WriteString(out, output)
					}
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF { //nolint:errorlint // bufio.Reader.ReadBytes returns io.EOF verbatim, never wrapped
				break
			}
			return "", fmt.Errorf("reading event stream: %w", readErr)
		}
	}

	return release, nil
}
