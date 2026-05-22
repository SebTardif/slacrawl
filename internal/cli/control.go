package cli

import (
	"fmt"

	"github.com/openclaw/crawlkit/control"
	"github.com/vincentkoc/slacrawl/internal/config"
)

func controlManifest(configPath string) control.Manifest {
	cfg := config.Default()
	m := control.NewManifest("slacrawl", "Slack Crawl", "slacrawl")
	m.Description = "Local-first Slack archive crawler."
	m.Branding = control.Branding{SymbolName: "number", AccentColor: "#4a154b", BundleIdentifier: "com.tinyspeck.slackmacgap"}
	m.Paths = control.Paths{
		DefaultConfig:   configPath,
		DefaultDatabase: cfg.DBPath,
		DefaultCache:    cfg.CacheDir,
		DefaultLogs:     cfg.LogDir,
		DefaultShare:    cfg.Share.RepoPath,
	}
	m.Capabilities = []string{"metadata", "doctor", "status", "sync", "watch", "search", "git-share"}
	m.Privacy = control.Privacy{ContainsPrivateMessages: true, ExportsSecrets: false, LocalOnlyScopes: []string{"slack", "desktop-cache", "sqlite", "git-share"}}
	m.Commands = map[string]control.Command{
		"doctor": {Title: "Doctor", Argv: []string{"slacrawl", "--json", "doctor"}, JSON: true},
		"status": {Title: "Status", Argv: []string{"slacrawl", "--json", "status"}, JSON: true},
		"sync":   {Title: "Sync", Argv: []string{"slacrawl", "--json", "sync", "--source", "all", "--latest-only"}, JSON: true, Mutates: true},
		"search": {Title: "Search", Argv: []string{"slacrawl", "--json", "search"}, JSON: true},
	}
	return m
}

func metadataOutputFormat(args []string, fallback OutputFormat) (OutputFormat, error) {
	format := fallback
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			format = FormatJSON
		case "--format":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--format requires text, json, or log")
			}
			resolved, err := resolveOutputFormat(args[i+1], false)
			if err != nil {
				return "", err
			}
			format = resolved
			i++
		default:
			return "", fmt.Errorf("metadata takes flags only")
		}
	}
	return format, nil
}
