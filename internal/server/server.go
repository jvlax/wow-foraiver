// Package server wires the tools into an MCP server, serving stdio for
// local agents and streamable HTTP for the hosted endpoint.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jvlax/wow-foraiver/internal/tools"
)

const instructions = `wow-foraiver: build WoW addons over blessed paths only.
Addon files in (read at /reload), SavedVariables out (flushed on exit),
Config.wtf between runs. Tools return file payloads — write them under
Interface/AddOns/ with your own file tools. If something needs a workaround,
stop: resistance means guidance is needed; friction is feedback.`

// New builds the MCP server with every published tool registered.
func New(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "wow-foraiver",
		Title:   "WoW forAIver",
		Version: version,
	}, &mcp.ServerOptions{Instructions: instructions})

	type scaffoldIn struct {
		Name  string `json:"name" jsonschema:"addon folder and TOC name, e.g. Compass"`
		Notes string `json:"notes,omitempty" jsonschema:"one-line description for the TOC"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name: "addon_scaffold",
		Description: "Generate a minimal working addon (TOC + entry file with a login event). " +
			"Returns file payloads relative to Interface/AddOns/ — write them to disk yourself.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in scaffoldIn) (*mcp.CallToolResult, tools.Bundle, error) {
		if in.Name == "" {
			return nil, tools.Bundle{}, fmt.Errorf("name is required")
		}
		return nil, tools.Scaffold(in.Name, in.Notes), nil
	})

	// gamepad_layout lived here until v0.2.0 — retired because the client's
	// native gamepad UI (modifiers included) turned out smoother than our
	// rigging. The doctrine working as intended: the native path won.

	type duplexIn struct {
		AddonName string `json:"addon_name,omitempty" jsonschema:"defaults to Duplex"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name: "duplex_kit",
		Description: "Generate the full-duplex primitive: payload.lua is the inbound request slot " +
			"(runs on every /reload), Report() writes answers to SavedVariables (outbound, lands on " +
			"flush). Pair with `wow-foraiver watch` on the SavedVariables file for the ack.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in duplexIn) (*mcp.CallToolResult, tools.Bundle, error) {
		return nil, tools.DuplexKit(in.AddonName), nil
	})

	type knowIn struct {
		Query     string `json:"query,omitempty" jsonschema:"substring to match against globals and namespace names"`
		Namespace string `json:"namespace,omitempty" jsonschema:"exact C_ namespace to list, e.g. C_GamePad"`
		Limit     int    `json:"limit,omitempty" jsonschema:"max globals returned, default 50"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name: "know_query",
		Description: "Query the client API truth: the real surface dumped from the running client " +
			"(5883 globals, 269 C_ namespaces). Check here before writing code — if the corpus " +
			"doesn't have it, the client doesn't either.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in knowIn) (*mcp.CallToolResult, tools.KnowResult, error) {
		if in.Query == "" && in.Namespace == "" {
			return nil, tools.KnowResult{}, fmt.Errorf("pass query or namespace")
		}
		return nil, tools.KnowQuery(in.Query, in.Namespace, in.Limit), nil
	})

	return s
}

// ServeStdio runs the server on stdin/stdout until the client hangs up.
func ServeStdio(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}

// ServeHTTP runs the streamable-HTTP endpoint (the hosted shape).
//
// Localhost DNS-rebinding protection is off here by design: the hosted
// endpoint sits behind an ingress that proxies via loopback, so every
// legitimate request carries a public Host header. Rebinding protection
// belongs to the stdio/local shape, which doesn't pass through here.
func ServeHTTP(addr, version string) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return New(version)
	}, &mcp.StreamableHTTPOptions{
		DisableLocalhostProtection: true,
		SessionTimeout:             10 * time.Minute,
	})
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return http.ListenAndServe(addr, mux)
}
