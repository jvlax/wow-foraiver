package tools

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"sync"
)

//go:embed know_data.json
var knowRaw []byte

// Know is the client API truth: the surface dumped from a running client.
// When the game updates, this corpus updates; tool contracts do not.
type Know struct {
	Build struct {
		Build     string `json:"build"`
		Interface string `json:"interface"`
		Version   string `json:"version"`
	} `json:"build"`
	Locale     string              `json:"locale"`
	Globals    []string            `json:"globals"`
	Namespaces map[string][]string `json:"namespaces"`
}

var (
	knowOnce sync.Once
	know     Know
)

func corpus() *Know {
	knowOnce.Do(func() {
		if err := json.Unmarshal(knowRaw, &know); err != nil {
			panic("embedded know corpus is invalid: " + err.Error())
		}
	})
	return &know
}

// KnowResult is the answer to a corpus query.
type KnowResult struct {
	ClientVersion string   `json:"client_version"`
	Build         string   `json:"build"`
	Interface     string   `json:"interface"`
	Namespaces    []string `json:"namespaces,omitempty"`
	Globals       []string `json:"globals,omitempty"`
	Functions     []string `json:"functions,omitempty"`
}

// KnowQuery searches the corpus. With ns set, lists that namespace's
// functions; otherwise substring-matches globals and namespace names.
func KnowQuery(query, ns string, limit int) KnowResult {
	c := corpus()
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	res := KnowResult{
		ClientVersion: c.Build.Version,
		Build:         c.Build.Build,
		Interface:     c.Build.Interface,
	}

	if ns != "" {
		if fns, ok := c.Namespaces[ns]; ok {
			res.Namespaces = []string{ns}
			res.Functions = fns
		}
		return res
	}

	q := strings.ToLower(query)
	for name := range c.Namespaces {
		if strings.Contains(strings.ToLower(name), q) {
			res.Namespaces = append(res.Namespaces, name)
		}
	}
	sort.Strings(res.Namespaces)
	for _, g := range c.Globals {
		if strings.Contains(strings.ToLower(g), q) {
			res.Globals = append(res.Globals, g)
			if len(res.Globals) >= limit {
				break
			}
		}
	}
	return res
}

// InterfaceVersion is what generated TOC files declare.
func InterfaceVersion() string { return corpus().Build.Interface }
