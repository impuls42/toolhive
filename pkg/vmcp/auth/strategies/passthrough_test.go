// SPDX-FileCopyrightText: Copyright 2025 Stacklok, Inc.
// SPDX-License-Identifier: Apache-2.0

package strategies

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/toolhive/pkg/auth"
	authtypes "github.com/stacklok/toolhive/pkg/vmcp/auth/types"
	"github.com/stacklok/toolhive/pkg/vmcp/health"
)

func TestPassthroughStrategy_Authenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupCtx      func() context.Context
		strategy      *authtypes.BackendAuthStrategy
		systemToken   string
		expectError   bool
		errorContains string
		checkHeader   func(t *testing.T, req *http.Request)
	}{
		{
			name: "successfully passes through token",
			setupCtx: func() context.Context {
				identity := &auth.Identity{
					Token:     "test-token-123",
					TokenType: "Bearer",
					Subject:   "user-1",
				}
				return auth.WithIdentity(context.Background(), identity)
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer test-token-123", req.Header.Get("Authorization"))
			},
		},
		{
			name: "uses custom header name",
			setupCtx: func() context.Context {
				identity := &auth.Identity{
					Token: "test-token-123",
				}
				return auth.WithIdentity(context.Background(), identity)
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
				Passthrough: &authtypes.PassthroughConfig{
					HeaderName: "X-Custom-Auth",
				},
			},
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer test-token-123", req.Header.Get("X-Custom-Auth"))
			},
		},
		{
			name: "skips if health check and no identity",
			setupCtx: func() context.Context {
				return health.WithHealthCheckMarker(context.Background())
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Empty(t, req.Header.Get("Authorization"))
			},
		},
		{
			name: "fails if no identity and not health check",
			setupCtx: func() context.Context {
				return context.Background()
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			expectError:   true,
			errorContains: "no identity found",
		},
		{
			name: "falls back to system token if no identity",
			setupCtx: func() context.Context {
				return context.Background()
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			systemToken: "system-token-123",
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer system-token-123", req.Header.Get("Authorization"))
			},
		},
		{
			name: "falls back to system token if identity has no token",
			setupCtx: func() context.Context {
				identity := &auth.Identity{
					Token: "",
				}
				return auth.WithIdentity(context.Background(), identity)
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			systemToken: "system-token-123",
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer system-token-123", req.Header.Get("Authorization"))
			},
		},
		{
			name: "user token takes precedence over system token",
			setupCtx: func() context.Context {
				identity := &auth.Identity{
					Token: "user-token-123",
				}
				return auth.WithIdentity(context.Background(), identity)
			},
			strategy: &authtypes.BackendAuthStrategy{
				Type: authtypes.StrategyTypePassthrough,
			},
			systemToken: "system-token-123",
			expectError: false,
			checkHeader: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer user-token-123", req.Header.Get("Authorization"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strategy := NewPassthroughStrategy(tt.systemToken)
			ctx := tt.setupCtx()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			err := strategy.Authenticate(ctx, req, tt.strategy)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			if tt.checkHeader != nil {
				tt.checkHeader(t, req)
			}
		})
	}
}
