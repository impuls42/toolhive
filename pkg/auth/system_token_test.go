// SPDX-FileCopyrightText: Copyright 2025 Stacklok, Inc.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemTokenAuthentication(t *testing.T) {
	// Save current env var and restore after test
	originalToken := os.Getenv(SystemTokenEnvVar)
	defer func() {
		if originalToken != "" {
			os.Setenv(SystemTokenEnvVar, originalToken)
		} else {
			os.Unsetenv(SystemTokenEnvVar)
		}
	}()

	// Set system token
	testSystemToken := "test-system-token-123"
	os.Setenv(SystemTokenEnvVar, testSystemToken)

	ctx := context.Background()

	// Get middleware (no OIDC config, so it falls back to local user + system token wrapper)
	middleware, _, err := GetAuthenticationMiddleware(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, middleware)

	// Create a test handler that verifies the identity
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFromContext(r.Context())
		require.True(t, ok, "Identity should be present")

		// Check if it's the system identity
		if identity.Subject == "toolhive-system" {
			assert.Equal(t, testSystemToken, identity.Token)
			assert.Equal(t, bearerTokenType, identity.TokenType)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("system-authenticated"))
		} else {
			// It must be the local user identity
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("local-authenticated"))
		}
	})

	wrappedHandler := middleware(testHandler)

	t.Run("valid_system_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+testSystemToken)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "system-authenticated", w.Body.String())
	})

	t.Run("invalid_system_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		// Should fall back to local auth
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "local-authenticated", w.Body.String())
	})

	t.Run("no_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		// Should fall back to local auth
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "local-authenticated", w.Body.String())
	})
}
