package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	plugin "github.com/SemRels/hook-matrix/internal/plugin"
)

type notifier interface {
	Notify(context.Context, string, string, string) error
}

var newNotifier = func(cfg plugin.MatrixConfig) notifier {
	return plugin.NewMatrixNotifier(cfg)
}

func run(ctx context.Context, getenv func(string) string, stderr io.Writer) int {
	homeserverURL := getenv("SEMREL_PLUGIN_HOMESERVER_URL")
	token := getenv("SEMREL_PLUGIN_TOKEN")
	roomID := getenv("SEMREL_PLUGIN_ROOM_ID")
	version := firstNonEmpty(getenv("SEMREL_VERSION"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_NEXT_VERSION"))

	if homeserverURL == "" || token == "" || roomID == "" {
		fmt.Fprintln(stderr, "hook-matrix: SEMREL_PLUGIN_HOMESERVER_URL, SEMREL_PLUGIN_TOKEN, and SEMREL_PLUGIN_ROOM_ID are required")
		return 1
	}
	if version == "" {
		fmt.Fprintln(stderr, "hook-matrix: SEMREL_VERSION, SEMREL_TAG_NAME, or SEMREL_NEXT_VERSION is required")
		return 1
	}
	if getenv("SEMREL_DRY_RUN") == "true" {
		return 0
	}

	cfg := plugin.MatrixConfig{
		HomeserverURL: homeserverURL,
		AccessToken:   token,
		RoomID:        roomID,
	}
	if err := newNotifier(cfg).Notify(ctx, version, getenv("SEMREL_CHANGELOG"), getenv("SEMREL_PLUGIN_REPOSITORY")); err != nil {
		fmt.Fprintln(stderr, "hook-matrix:", err)
		return 1
	}
	return 0
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	os.Exit(run(ctx, os.Getenv, os.Stderr))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
