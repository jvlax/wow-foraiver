package tools

import (
	"strings"
	"testing"
)

func TestKnowCorpusLoads(t *testing.T) {
	res := KnowQuery("GamePad", "", 50)
	if res.Interface != "16001" {
		t.Errorf("interface = %q, want 16001", res.Interface)
	}
	found := false
	for _, ns := range res.Namespaces {
		if ns == "C_GamePad" {
			found = true
		}
	}
	if !found {
		t.Errorf("C_GamePad missing from namespace matches: %v", res.Namespaces)
	}
}

func TestKnowNamespaceListing(t *testing.T) {
	res := KnowQuery("", "C_GamePad", 0)
	if len(res.Functions) == 0 {
		t.Fatal("C_GamePad listed no functions")
	}
	has := func(name string) bool {
		for _, f := range res.Functions {
			if f == name {
				return true
			}
		}
		return false
	}
	if !has("ApplyConfigs") {
		t.Errorf("C_GamePad missing ApplyConfigs: %v", res.Functions)
	}
}

func TestScaffold(t *testing.T) {
	b := Scaffold("Compass", "")
	if len(b.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(b.Files))
	}
	tocFile := b.Files[0]
	if tocFile.Path != "Compass/Compass.toc" {
		t.Errorf("toc path = %q", tocFile.Path)
	}
	for _, want := range []string{"## Interface: 16001", "## Title: Compass", "Compass.lua"} {
		if !strings.Contains(tocFile.Content, want) {
			t.Errorf("toc missing %q:\n%s", want, tocFile.Content)
		}
	}
	if !strings.Contains(b.Files[1].Content, "PLAYER_LOGIN") {
		t.Error("scaffold lua has no login event")
	}
}

func TestGamepadLayoutDefaults(t *testing.T) {
	b := GamepadLayout(GamepadSpec{})
	if len(b.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(b.Files))
	}
	lua := b.Files[1].Content
	// The full base layout must be present.
	for _, want := range []string{
		`["PADA"] = "JUMP"`,
		`["SHIFT-PADX"] = "ACTIONBUTTON8"`,
		`["SHIFT-PADDRIGHT"] = "ACTIONBUTTON12"`,
		`["PADRSHOULDER"] = "TARGETNEARESTENEMY"`,
		`["PADRTRIGGER"] = "TOGGLEAUTORUN"`,
		`["GamePadEmulateShift"] = "PADLSHOULDER"`,
		`SaveBindings(GetCurrentBindingSet())`,
		"local VERSION = 1",
		"SLASH_TILLER1",
	} {
		if !strings.Contains(lua, want) {
			t.Errorf("layout lua missing %q", want)
		}
	}
	if !strings.Contains(b.Files[0].Content, "## SavedVariables: TillerDB") {
		t.Error("toc missing SavedVariables")
	}
}

func TestGamepadLayoutCustom(t *testing.T) {
	b := GamepadLayout(GamepadSpec{
		AddonName: "Helm",
		Version:   3,
		Bindings:  map[string]string{"PADA": "TARGETSELF"},
	})
	lua := b.Files[1].Content
	if !strings.Contains(lua, "local VERSION = 3") {
		t.Error("custom version not honored")
	}
	if !strings.Contains(lua, `["PADA"] = "TARGETSELF"`) {
		t.Error("custom binding not honored")
	}
	if strings.Contains(lua, "ACTIONBUTTON1") {
		t.Error("custom bindings should replace, not merge, the base layout")
	}
	if !strings.Contains(lua, "SLASH_HELM1") {
		t.Error("slash command not derived from addon name")
	}
}

func TestDuplexKit(t *testing.T) {
	b := DuplexKit("")
	if len(b.Files) != 3 {
		t.Fatalf("want 3 files, got %d", len(b.Files))
	}
	var tocContent, payload string
	for _, f := range b.Files {
		if strings.HasSuffix(f.Path, ".toc") {
			tocContent = f.Content
		}
		if strings.HasSuffix(f.Path, "payload.lua") {
			payload = f.Content
		}
	}
	if !strings.Contains(tocContent, "payload.lua") {
		t.Error("toc does not load payload.lua")
	}
	if !strings.Contains(tocContent, "## SavedVariables: DuplexDB") {
		t.Error("toc missing SavedVariables — the outbound path")
	}
	if !strings.Contains(payload, "Report(") {
		t.Error("payload example does not demonstrate Report")
	}
}
