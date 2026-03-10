// SPDX-FileCopyrightText: Copyright 2025 Stacklok, Inc.
// SPDX-License-Identifier: Apache-2.0

package converters

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	authtypes "github.com/stacklok/toolhive/pkg/vmcp/auth/types"
)

// PassthroughConverter converts JWT passthrough config to vMCP strategy.
type PassthroughConverter struct{}

// StrategyType returns the vMCP strategy type for passthrough.
func (*PassthroughConverter) StrategyType() string {
	return authtypes.StrategyTypePassthrough
}

// ConvertToStrategy converts config to a BackendAuthStrategy.
func (*PassthroughConverter) ConvertToStrategy(
	_ *mcpv1alpha1.MCPExternalAuthConfig,
) (*authtypes.BackendAuthStrategy, error) {
	// Passthrough has no config in MCPExternalAuthConfig yet,
	// but we can support it if we add a type to the CRD.
	return &authtypes.BackendAuthStrategy{
		Type: authtypes.StrategyTypePassthrough,
		Passthrough: &authtypes.PassthroughConfig{
			HeaderName: "Authorization",
		},
	}, nil
}

// ResolveSecrets is a no-op for passthrough.
func (*PassthroughConverter) ResolveSecrets(
	_ context.Context,
	_ *mcpv1alpha1.MCPExternalAuthConfig,
	_ client.Client,
	_ string,
	strategy *authtypes.BackendAuthStrategy,
) (*authtypes.BackendAuthStrategy, error) {
	return strategy, nil
}
