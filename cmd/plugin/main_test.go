package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	plugin "github.com/SemRels/hook-matrix/internal/plugin"
)

type fakeNotifier struct {
	version string
	repo    string
	err     error
}

func (f *fakeNotifier) Notify(_ context.Context, version, _ string, repository string) error {
	f.version = version
	f.repo = repository
	return f.err
}

func env(kv map[string]string) func(string) string {
	return func(key string) string { return kv[key] }
}

func TestRunSuccess(t *testing.T) {

	fake := &fakeNotifier{}
	old := newNotifier
	newNotifier = func(cfg plugin.MatrixConfig) notifier {
		if cfg.RoomID != "!room:example.com" {
			t.Fatalf("unexpected room id: %s", cfg.RoomID)
		}
		return fake
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_HOMESERVER_URL": "https://matrix.example.com",
		"SEMREL_PLUGIN_TOKEN":          "token",
		"SEMREL_PLUGIN_ROOM_ID":        "!room:example.com",
		"SEMREL_VERSION":               "v1.2.3",
		"SEMREL_PLUGIN_REPOSITORY":     "SemRels/semrel",
	}), &stderr)
	if code != 0 || stderr.String() != "plugin_schema_version=1\n" {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
	if fake.version != "v1.2.3" || fake.repo != "SemRels/semrel" {
		t.Fatalf("unexpected notify args: %+v", fake)
	}
}

func TestRunDryRun(t *testing.T) {

	called := false
	old := newNotifier
	newNotifier = func(plugin.MatrixConfig) notifier {
		called = true
		return &fakeNotifier{}
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_HOMESERVER_URL": "https://matrix.example.com",
		"SEMREL_PLUGIN_TOKEN":          "token",
		"SEMREL_PLUGIN_ROOM_ID":        "!room:example.com",
		"SEMREL_VERSION":               "v1.2.3",
		"SEMREL_DRY_RUN":               "true",
	}), &stderr)
	if code != 0 || called {
		t.Fatalf("unexpected result: code=%d called=%v", code, called)
	}
}

func TestRunValidationError(t *testing.T) {

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunNotifyError(t *testing.T) {

	old := newNotifier
	newNotifier = func(plugin.MatrixConfig) notifier {
		return &fakeNotifier{err: errors.New("boom")}
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_HOMESERVER_URL": "https://matrix.example.com",
		"SEMREL_PLUGIN_TOKEN":          "token",
		"SEMREL_PLUGIN_ROOM_ID":        "!room:example.com",
		"SEMREL_VERSION":               "v1.2.3",
	}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunMissingVersion(t *testing.T) {
	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_HOMESERVER_URL": "https://matrix.example.com",
		"SEMREL_PLUGIN_TOKEN":          "token",
		"SEMREL_PLUGIN_ROOM_ID":        "!room:example.com",
	}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "v1.2.3", "v1.2.4"); got != "v1.2.3" {
		t.Fatalf("unexpected value: %s", got)
	}
}
