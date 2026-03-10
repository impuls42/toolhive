// SPDX-FileCopyrightText: Copyright 2025 Stacklok, Inc.
// SPDX-License-Identifier: Apache-2.0

package strategies

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stacklok/toolhive/pkg/auth"
	authtypes "github.com/stacklok/toolhive/pkg/vmcp/auth/types"
	"github.com/stacklok/toolhive/pkg/vmcp/health"
)

// PassthroughStrategy forwards the incoming bearer token directly to the backend.
type PassthroughStrategy struct {
	systemToken string
}

// NewPassthroughStrategy creates a new PassthroughStrategy instance.
func NewPassthroughStrategy(systemToken string) *PassthroughStrategy {
	return &PassthroughStrategy{
		systemToken: systemToken,
	}
}

// Name returns the strategy identifier.
func (*PassthroughStrategy) Name() string {
	return authtypes.StrategyTypePassthrough
}

// Authenticate forwards the incoming token from the context to the request header.
func (s *PassthroughStrategy) Authenticate(
	ctx context.Context, req *http.Request, strategy *authtypes.BackendAuthStrategy,
) error {
	identity, ok := auth.IdentityFromContext(ctx)
	if !ok || identity.Token == "" {
		// Fallback to system token if configured
		if s.systemToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.systemToken))
			return nil
		}

		// Skip authentication for health checks if no identity is present.
		// If a system token is configured, the health monitor will provide an identity.
		if health.IsHealthCheck(ctx) {
			return nil
		}

		if !ok {
			return fmt.Errorf("no identity found in context")
		}
		return fmt.Errorf("identity has no token")
	}

	headerName := "Authorization"
	if strategy != nil && strategy.Passthrough != nil && strategy.Passthrough.HeaderName != "" {
		headerName = strategy.Passthrough.HeaderName
	}

	req.Header.Set(headerName, fmt.Sprintf("Bearer %s", identity.Token))
	return nil
}

// Validate checks if the strategy configuration is valid.
func (*PassthroughStrategy) Validate(strategy *authtypes.BackendAuthStrategy) error {
	if strategy == nil || strategy.Type != authtypes.StrategyTypePassthrough {
		return fmt.Errorf("invalid strategy configuration for passthrough")
	}
	return nil
}
