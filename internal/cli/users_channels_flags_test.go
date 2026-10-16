package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/openclaw/slacrawl/internal/config"
	"github.com/openclaw/slacrawl/internal/store"
)

func TestUsersAndChannelsHonorFlagsAfterTheQuery(t *testing.T) {
	t.Setenv("SLACRAWL_NO_UPDATE_CHECK", "1")
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	dbPath := filepath.Join(dir, "slacrawl.db")
	cfg := config.Default()
	cfg.DBPath = dbPath
	cfg.CacheDir = filepath.Join(dir, "cache")
	cfg.WorkspaceID = "T1"
	cfg.Share.Remote = ""
	cfg.Share.AutoUpdate = false
	cfg.Slack.Bot.Enabled = false
	cfg.Slack.App.Enabled = false
	cfg.Slack.User.Enabled = false
	cfg.Slack.Desktop.Enabled = false
	require.NoError(t, cfg.Save(configPath))

	ctx := context.Background()
	st, err := store.Open(dbPath)
	require.NoError(t, err)
	now := mustTime(t, "2026-03-08T18:20:43Z")
	require.NoError(t, st.UpsertWorkspace(ctx, store.Workspace{ID: "T1", Name: "one", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.UpsertWorkspace(ctx, store.Workspace{ID: "T2", Name: "two", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.UpsertChannel(ctx, store.Channel{ID: "C1", WorkspaceID: "T1", Name: "general", Kind: "public_channel", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.UpsertChannel(ctx, store.Channel{ID: "C2", WorkspaceID: "T2", Name: "general", Kind: "public_channel", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.UpsertUser(ctx, store.User{ID: "U1", WorkspaceID: "T1", Name: "alice", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.UpsertUser(ctx, store.User{ID: "U2", WorkspaceID: "T2", Name: "bob", RawJSON: "{}", UpdatedAt: now}))
	require.NoError(t, st.Close())

	app := &App{}
	aliceOnT2 := runJSONRows(t, app, "--config", configPath, "--json", "users", "alice", "--workspace", "T2")
	require.Empty(t, aliceOnT2)

	bob := runJSONRows(t, app, "--config", configPath, "--json", "users", "bob", "--workspace", "T2", "--limit", "5")
	require.Len(t, bob, 1)
	require.Equal(t, "T2", bob[0]["workspace_id"])
	require.Equal(t, "bob", bob[0]["name"])

	bobFirst := runJSONRows(t, app, "--config", configPath, "--json", "users", "--workspace", "T2", "bob")
	require.Equal(t, bob, bobFirst)

	generalOnT2 := runJSONRows(t, app, "--config", configPath, "--json", "channels", "general", "--workspace", "T2")
	require.Len(t, generalOnT2, 1)
	require.Equal(t, "T2", generalOnT2[0]["workspace_id"])
	require.Equal(t, "C2", generalOnT2[0]["id"])

	generalOnT1 := runJSONRows(t, app, "--config", configPath, "--json", "channels", "--workspace", "T1", "--kind", "public_channel", "general")
	require.Len(t, generalOnT1, 1)
	require.Equal(t, "T1", generalOnT1[0]["workspace_id"])
	require.Equal(t, "C1", generalOnT1[0]["id"])
}

func runJSONRows(t *testing.T, app *App, args ...string) []map[string]any {
	t.Helper()
	var stdout bytes.Buffer
	app.Stdout = &stdout
	app.Stderr = &stdout
	require.NoError(t, app.Run(context.Background(), args))
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	return rows
}
