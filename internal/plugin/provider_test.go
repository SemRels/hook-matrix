// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatrixHookExecutePostsMessage(t *testing.T) {
	t.Parallel()

	var received struct {
		Auth string
		Body map[string]string
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.EscapedPath(), "/_matrix/client/v3/rooms/")
		received.Auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received.Body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"event_id":"$event-1"}`))
	}))
	defer server.Close()

	hook := NewMatrixHook(server.Client(), server.URL, "!room:example.org", "token-123")
	result, err := hook.Execute(context.Background(), &Release{Version: "1.2.3", Repository: "SemRels/semrel", Changelog: "- Added feature"})
	require.NoError(t, err)
	require.Equal(t, "Bearer token-123", received.Auth)
	require.Contains(t, received.Body["body"], "1.2.3")
	require.Equal(t, "$event-1", result.Outputs["event_id"])
}
