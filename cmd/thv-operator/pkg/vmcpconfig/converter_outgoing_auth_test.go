// SPDX-FileCopyrightText: Copyright 2025 Stacklok, Inc.
// SPDX-License-Identifier: Apache-2.0

package vmcpconfig

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	authtypes "github.com/stacklok/toolhive/pkg/vmcp/auth/types"
	vmcpconfig "github.com/stacklok/toolhive/pkg/vmcp/config"
)

func TestConverter_OutgoingAuthDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		outgoingAuth         *mcpv1alpha1.OutgoingAuthConfig
		expectedSource       string
		expectedDefaultType  string
		expectedBackendTypes map[string]string
	}{
		{
			name:                "nil outgoingAuth defaults to discovered with passthrough",
			outgoingAuth:        nil,
			expectedSource:      "discovered",
			expectedDefaultType: authtypes.StrategyTypePassthrough,
		},
		{
			name: "explicit discovered type maps to passthrough",
			outgoingAuth: &mcpv1alpha1.OutgoingAuthConfig{
				Source: "discovered",
				Backends: map[string]mcpv1alpha1.BackendAuthConfig{
					"backend1": {
						Type: mcpv1alpha1.BackendAuthTypeDiscovered,
					},
				},
			},
			expectedSource: "discovered",
			expectedBackendTypes: map[string]string{
				"backend1": authtypes.StrategyTypePassthrough,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			vmcpServer := &mcpv1alpha1.VirtualMCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmcp",
					Namespace: "default",
				},
				Spec: mcpv1alpha1.VirtualMCPServerSpec{
					Config:       vmcpconfig.Config{Group: "test-group"},
					IncomingAuth: &mcpv1alpha1.IncomingAuthConfig{Type: "anonymous"},
					OutgoingAuth: tt.outgoingAuth,
				},
			}

			converter := newTestConverter(t, newNoOpMockResolver(t))
			ctx := log.IntoContext(context.Background(), logr.Discard())
			config, err := converter.Convert(ctx, vmcpServer)

			require.NoError(t, err)
			require.NotNil(t, config)
			require.NotNil(t, config.OutgoingAuth)

			assert.Equal(t, tt.expectedSource, config.OutgoingAuth.Source)

			if tt.expectedDefaultType != "" {
				require.NotNil(t, config.OutgoingAuth.Default)
				assert.Equal(t, tt.expectedDefaultType, config.OutgoingAuth.Default.Type)
			}

			for backend, expectedType := range tt.expectedBackendTypes {
				strategy, ok := config.OutgoingAuth.Backends[backend]
				require.True(t, ok, "backend %s should exist", backend)
				assert.Equal(t, expectedType, strategy.Type)
			}
		})
	}
}
