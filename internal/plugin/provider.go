// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Release contains the SemRel release data consumed by this plugin.
type Release struct {
	Version         string
	PreviousVersion string
	TagName         string
	Repository      string
	Changelog       string
	CommitSHA       string
	DryRun          bool
	Metadata        map[string]string
	Commits         []string
}

// Result captures the outcome of a plugin execution.
type Result struct {
	Name       string
	Outputs    map[string]string
	Skipped    bool
	SkipReason string
}

// Provider is the contract exposed by this plugin implementation.
type Provider interface {
	Name() string
	HealthCheck(context.Context) error
	Validate(map[string]interface{}) error
	Execute(context.Context, *Release) (*Result, error)
	ReleaseContext() []string
}

// MatrixHook posts release announcements to a Matrix room.
type MatrixHook struct {
	Homeserver  string
	RoomID      string
	AccessToken string
	Client      *http.Client
}

// NewMatrixHook constructs a Matrix hook with explicit configuration.
func NewMatrixHook(client *http.Client, homeserver, roomID, accessToken string) *MatrixHook {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &MatrixHook{Homeserver: strings.TrimSpace(homeserver), RoomID: strings.TrimSpace(roomID), AccessToken: strings.TrimSpace(accessToken), Client: client}
}

// NewMatrixHookFromEnv constructs a Matrix hook from environment variables.
func NewMatrixHookFromEnv() *MatrixHook {
	return NewMatrixHook(nil, os.Getenv("MATRIX_HOMESERVER"), os.Getenv("MATRIX_ROOM_ID"), os.Getenv("MATRIX_ACCESS_TOKEN"))
}

func (m *MatrixHook) Name() string { return "hook-matrix" }

func (m *MatrixHook) HealthCheck(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *MatrixHook) Validate(map[string]interface{}) error {
	switch {
	case m.Homeserver == "":
		return fmt.Errorf("matrix: MATRIX_HOMESERVER is required")
	case m.RoomID == "":
		return fmt.Errorf("matrix: MATRIX_ROOM_ID is required")
	case m.AccessToken == "":
		return fmt.Errorf("matrix: MATRIX_ACCESS_TOKEN is required")
	}
	return nil
}

func (m *MatrixHook) ReleaseContext() []string {
	return []string{"version", "repository", "changelog"}
}

func (m *MatrixHook) Execute(ctx context.Context, rel *Release) (*Result, error) {
	if err := m.HealthCheck(ctx); err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, fmt.Errorf("matrix: release is required")
	}
	message := BuildMessage(rel)
	if rel.DryRun {
		return &Result{Name: m.Name(), Outputs: map[string]string{"room_id": m.RoomID, "message": message, "dry_run": "true"}}, nil
	}
	if err := m.Validate(nil); err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(normalizeBaseURL(m.Homeserver), "/") + "/_matrix/client/v3/rooms/" + url.PathEscape(m.RoomID) + "/send/m.room.message"
	body, err := json.Marshal(map[string]string{"msgtype": "m.text", "body": message})
	if err != nil {
		return nil, fmt.Errorf("matrix: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("matrix: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("matrix: send message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("matrix: unexpected status %s", resp.Status)
	}

	var response struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("matrix: decode response: %w", err)
	}

	return &Result{Name: m.Name(), Outputs: map[string]string{"room_id": m.RoomID, "event_id": response.EventID}}, nil
}

// BuildMessage renders the release announcement body.
func BuildMessage(rel *Release) string {
	version := strings.TrimSpace(rel.Version)
	if version == "" {
		version = "unknown"
	}
	var builder strings.Builder
	builder.WriteString("SemRel release ")
	builder.WriteString(version)
	if repository := strings.TrimSpace(rel.Repository); repository != "" {
		builder.WriteString(" for ")
		builder.WriteString(repository)
	}
	if notes := strings.TrimSpace(rel.Changelog); notes != "" {
		builder.WriteString("\n\n")
		builder.WriteString(notes)
	}
	return builder.String()
}

func normalizeBaseURL(homeserver string) string {
	homeserver = strings.TrimSpace(homeserver)
	if homeserver == "" {
		return ""
	}
	if strings.HasPrefix(homeserver, "http://") || strings.HasPrefix(homeserver, "https://") {
		return homeserver
	}
	return "https://" + homeserver
}
