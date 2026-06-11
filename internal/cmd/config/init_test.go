package config

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const schemaURL = "https://raw.githubusercontent.com/adaouat/bifrost/main/bifrost.schema.json"

func TestScaffold_IncludesSchemaModeline(t *testing.T) {
	modeline := "# yaml-language-server: $schema=" + schemaURL
	// The modeline must be the first line so the YAML language server reliably
	// associates the file with the schema.
	if !strings.HasPrefix(scaffold, modeline+"\n") {
		first, _, _ := strings.Cut(scaffold, "\n")
		t.Errorf("scaffold first line must be the schema modeline\n got: %s\nwant: %s", first, modeline)
	}
}

func TestBifrostSchema_IsValidJSONSchema(t *testing.T) {
	data, err := os.ReadFile("../../../bifrost.schema.json")
	if err != nil {
		t.Fatalf("reading schema file: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if _, ok := doc["$schema"]; !ok {
		t.Error("schema missing $schema draft declaration")
	}

	props, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema missing top-level properties object")
	}
	for _, key := range []string{"strategy", "servers", "paths", "settings", "variables", "hooks", "environments"} {
		if _, ok := props[key]; !ok {
			t.Errorf("schema properties missing %q", key)
		}
	}
}
