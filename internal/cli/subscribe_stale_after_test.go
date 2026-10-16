package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/openclaw/slacrawl/internal/config"
)

func TestSubscribeRejectsStaleAfterBeforeSaving(t *testing.T) {
	t.Setenv("SLACRAWL_NO_UPDATE_CHECK", "1")
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	cfg := config.Default()
	cfg.DBPath = filepath.Join(dir, "slacrawl.db")
	cfg.CacheDir = filepath.Join(dir, "cache")
	cfg.WorkspaceID = "TKEEP"
	cfg.Share.StaleAfter = "15m"
	cfg.Slack.Desktop.Enabled = true
	cfg.Slack.Desktop.Path = filepath.Join(dir, "Slack")
	require.NoError(t, cfg.Save(configPath))
	before, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var output bytes.Buffer
	app := &App{Stdout: &output, Stderr: &output}
	err = app.Run(context.Background(), []string{
		"--config", configPath, "subscribe",
		"--stale-after", "90d",
		"--repo", filepath.Join(dir, "share"),
		filepath.Join(dir, "missing.git"),
	})
	require.ErrorContains(t, err, "invalid share.stale_after")
	after, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, string(before), string(after))

	loaded, err := config.Load(configPath)
	require.NoError(t, err)
	require.Equal(t, "15m", loaded.Share.StaleAfter)
	require.Equal(t, filepath.Join(dir, "Slack"), loaded.Slack.Desktop.Path)
}
