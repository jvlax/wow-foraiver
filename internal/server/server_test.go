package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect wires a real client to the real server over in-memory transports —
// the full protocol, no network.
func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ct, st := mcp.NewInMemoryTransports()
	srv := New("test")
	if _, err := srv.Connect(context.Background(), st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(context.Background(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func TestToolsPublished(t *testing.T) {
	cs := connect(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %s has no description — the MCP is where we publish, descriptions are the docs", tool.Name)
		}
	}
	for _, want := range []string{"addon_scaffold", "gamepad_layout", "duplex_kit", "know_query"} {
		if !got[want] {
			t.Errorf("tool %s not published (have %v)", want, got)
		}
	}
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s returned tool error: %v", name, res.Content)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("%s: structured content: %v", name, err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return out
}

func TestGamepadLayoutRoundTrip(t *testing.T) {
	cs := connect(t)
	out := callTool(t, cs, "gamepad_layout", map[string]any{})
	files, _ := out["files"].([]any)
	if len(files) != 2 {
		t.Fatalf("want 2 files, got %d", len(files))
	}
	lua := files[1].(map[string]any)["content"].(string)
	if !strings.Contains(lua, `["SHIFT-PADDRIGHT"] = "ACTIONBUTTON12"`) {
		t.Error("base layout incomplete over the wire")
	}
}

func TestKnowQueryRoundTrip(t *testing.T) {
	cs := connect(t)
	out := callTool(t, cs, "know_query", map[string]any{"namespace": "C_GamePad"})
	if out["interface"] != "16001" {
		t.Errorf("interface = %v, want 16001", out["interface"])
	}
	fns, _ := out["functions"].([]any)
	if len(fns) < 20 {
		t.Errorf("C_GamePad returned %d functions, want 25ish", len(fns))
	}
}

func TestDuplexKitRoundTrip(t *testing.T) {
	cs := connect(t)
	out := callTool(t, cs, "duplex_kit", map[string]any{})
	files, _ := out["files"].([]any)
	if len(files) != 3 {
		t.Fatalf("want 3 files, got %d", len(files))
	}
}

func TestScaffoldValidation(t *testing.T) {
	cs := connect(t)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "addon_scaffold", Arguments: map[string]any{},
	})
	if err == nil && !res.IsError {
		t.Error("scaffold without a name should fail")
	}
}
