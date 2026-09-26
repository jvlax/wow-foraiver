# wow-foraiver

WoW addon tooling for AI-assisted building, published as MCP tools. Anyone
who can code with Claude should be able to build their entire UI from a
vibed prompt — this is the toolbox that makes that true.

## The doctrine

Everything here rides **blessed paths only** — the doors the game opens for
everyone:

- **Inbound**: addon files, read at load and `/reload`. The only way bytes
  on disk become behavior in-game.
- **Outbound**: SavedVariables, written on flush. The only way in-game state
  becomes bytes on disk.
- **Sideband**: `Config.wtf` — the client owns it while running, you own it
  between runs.

No injection, no memory patching, nothing that fights the client or breaks
on the next build. If a tool here ever needs a workaround, the tool is
wrong, not the game. **Resistance means guidance is needed. Friction is
feedback.**

## The tools

| MCP tool | What it does |
|---|---|
| `addon_scaffold` | A minimal working addon — TOC + entry file, ready to grow |
| `duplex_kit` | The full-duplex primitive: `payload.lua` in via `/reload`, `Report()` out via SavedVariables |
| `know_query` | The client API truth — 5883 globals, 269 `C_` namespaces, dumped from the running client |

Tools return **file payloads**; your agent writes them under
`Interface/AddOns/`. The truth layer (`know_query`) is versioned with the
game: when the client updates, the corpus updates behind the same tool
contract — your prompts and clients don't change.

## Quickstart

Grab a binary from [releases](../../releases) (built by GoReleaser, GitHub
OIDC build provenance attached — `gh attestation verify` works), then:

```sh
claude mcp add wow-foraiver -- wow-foraiver serve
```

Ask Claude for a UI. That's the product.

### The companion commands

```sh
wow-foraiver wtf-apply spec.json   # assert cvars into Config.wtf after the client exits
wow-foraiver watch --file <SavedVariables/Addon.lua> --exec ./on-answer
wow-foraiver serve --http :8080   # the hosted endpoint shape (streamable HTTP at /mcp)
```

`wtf-apply` takes `{"config": "<path>", "set": {"CVar": "value"}}` — the
declarative counter to "the client rewrites its config on exit": you simply
write last.

## The duplex loop

```
edit payload.lua ──/reload──▶ runs in-game ──Report()──▶ SavedVariables
      ▲                                                      │
      └────────────── wow-foraiver watch ◀───flush───────────┘
```

An ordered, durable request/response cycle built entirely from sanctioned
persistence. Same shape as a commit queue: requests ride in on the game's
own schedule, answers ride out on its own flush.

## House rules

- CI is Go (`go run ./ci`) — no bash in workflows.
- Releases are git tags → GoReleaser → attested binaries. Nothing manual.
- License: **AGPL-3.0** — run a modified endpoint, publish your source.
