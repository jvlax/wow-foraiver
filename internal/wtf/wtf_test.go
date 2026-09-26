package wtf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyRewritesAndAppends(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "Config.wtf")
	initial := "SET GxApi \"D3D12\"\nSET GamePadEnable \"0\"\nSET textLocale \"enUS\"\n"
	if err := os.WriteFile(cfg, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	n, err := Apply(Spec{Config: cfg, Set: map[string]string{
		"GamePadEnable":       "1",
		"GamePadEmulateShift": "PADLSHOULDER",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("asserted %d, want 2", n)
	}

	got, _ := os.ReadFile(cfg)
	s := string(got)
	for _, want := range []string{
		`SET GxApi "D3D12"`,                      // untouched line survives
		`SET textLocale "enUS"`,                  // untouched line survives
		`SET GamePadEnable "1"`,                  // rewritten in place
		`SET GamePadEmulateShift "PADLSHOULDER"`, // appended
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Contains(s, `SET GamePadEnable "0"`) {
		t.Error("old value survived rewrite")
	}
	if strings.Index(s, "GamePadEnable") != strings.LastIndex(s, "GamePadEnable") {
		t.Error("duplicate SET line for rewritten cvar")
	}
}

func TestApplyCreatesMissingFile(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "Config.wtf")
	if _, err := Apply(Spec{Config: cfg, Set: map[string]string{"A": "1", "B": "2"}}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(cfg)
	if string(got) != "SET A \"1\"\nSET B \"2\"\n" {
		t.Errorf("unexpected content:\n%s", got)
	}
}

func TestApplyIdempotent(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "Config.wtf")
	spec := Spec{Config: cfg, Set: map[string]string{"GamePadEnable": "1"}}
	if _, err := Apply(spec); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(cfg)
	if _, err := Apply(spec); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(cfg)
	if string(first) != string(second) {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}
