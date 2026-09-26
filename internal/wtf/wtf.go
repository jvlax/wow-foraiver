// Package wtf asserts declared values into a WoW Config.wtf-style file.
//
// The client rewrites its config on exit; whoever writes last wins, so this
// runs after the client lets go. The primitive: declarative desired-state
// for any key-value config a program insists on owning.
package wtf

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Spec declares the file to own and the values to assert.
type Spec struct {
	Config string            `json:"config"`
	Set    map[string]string `json:"set"`
}

var setLine = regexp.MustCompile(`^SET (\S+) "(.*)"$`)

// Apply rewrites matching SET lines in place, appends missing ones, and
// touches nothing else. Returns the number of values asserted.
func Apply(spec Spec) (int, error) {
	if spec.Config == "" {
		return 0, fmt.Errorf("spec has no config path")
	}
	want := make(map[string]string, len(spec.Set))
	for k, v := range spec.Set {
		want[k] = v
	}

	var lines []string
	if raw, err := os.ReadFile(spec.Config); err == nil {
		lines = strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	} else if !os.IsNotExist(err) {
		return 0, err
	}

	out := make([]string, 0, len(lines)+len(want))
	for _, line := range lines {
		if m := setLine.FindStringSubmatch(line); m != nil {
			if v, ok := want[m[1]]; ok {
				out = append(out, fmt.Sprintf("SET %s %q", m[1], v))
				delete(want, m[1])
				continue
			}
		}
		out = append(out, line)
	}
	// Deterministic append order for the leftovers.
	for _, k := range sortedKeys(want) {
		out = append(out, fmt.Sprintf("SET %s %q", k, want[k]))
	}

	if err := os.WriteFile(spec.Config, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		return 0, err
	}
	return len(spec.Set), nil
}

// ApplyFile loads a JSON spec from disk and applies it.
func ApplyFile(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var spec Spec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return Apply(spec)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
