// wow-foraiver — WoW addon tooling over blessed paths, published as MCP tools.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jvlax/wow-foraiver/internal/server"
	"github.com/jvlax/wow-foraiver/internal/wtf"
)

// version is stamped by goreleaser via ldflags.
var version = "dev"

const usage = `wow-foraiver %s — WoW addon tooling over blessed paths.

usage:
  wow-foraiver serve [--http ADDR]     MCP server (stdio default; --http for the endpoint)
  wow-foraiver wtf-apply SPEC.json     assert cvars into a Config.wtf (run after client exit)
  wow-foraiver watch --file PATH [--exec CMD]
                                       fire when a SavedVariables file flushes
  wow-foraiver version
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, usage, version)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		if len(os.Args) >= 4 && os.Args[2] == "--http" {
			fail(server.ServeHTTP(os.Args[3], version))
		}
		fail(server.ServeStdio(context.Background(), version))
	case "wtf-apply":
		if len(os.Args) != 3 {
			fatal("wtf-apply needs exactly one spec path")
		}
		n, err := wtf.ApplyFile(os.Args[2])
		fail(err)
		fmt.Printf("wtf-apply: %d values asserted\n", n)
	case "watch":
		fail(watch(os.Args[2:]))
	case "version":
		fmt.Println(version)
	default:
		fmt.Fprintf(os.Stderr, usage, version)
		os.Exit(2)
	}
}

// watch polls a file's mtime; on change (a SavedVariables flush) it prints
// and optionally execs a hook. The outbound half of the duplex loop.
func watch(args []string) error {
	var file, execCmd string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			i++
			file = args[i]
		case "--exec":
			i++
			execCmd = args[i]
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if file == "" {
		return fmt.Errorf("watch needs --file")
	}
	var last time.Time
	if st, err := os.Stat(file); err == nil {
		last = st.ModTime()
	}
	fmt.Printf("watch: %s (ctrl-c to stop)\n", file)
	for {
		time.Sleep(2 * time.Second)
		st, err := os.Stat(file)
		if err != nil {
			continue
		}
		if st.ModTime().After(last) {
			last = st.ModTime()
			fmt.Printf("flush: %s at %s\n", file, last.Format(time.RFC3339))
			if execCmd != "" {
				runHook(execCmd, file)
			}
		}
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "wow-foraiver:", msg)
	os.Exit(1)
}

func fail(err error) {
	if err != nil {
		fatal(err.Error())
	}
}
