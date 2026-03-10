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
	vmcpconfig "github.com/stacklok/toolhive/pkg/vmcp/config"
)

func TestConverter_SystemTokenEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outgoingAuth *mcpv1alpha1.OutgoingAuthConfig
		expectedEnv  string
		description  string
	}{
		{
			name:         "nil outgoingAuth",
			outgoingAuth: nil,
			expectedEnv:  "",
			description:  "Should not set SystemTokenEnv when outgoingAuth is nil",
		},
		{
			name: "outgoingAuth without system token ref",
			outgoingAuth: &mcpv1alpha1.OutgoingAuthConfig{
				Source: "discovered",
			},
			expectedEnv: "",
			description: "Should not set SystemTokenEnv when SystemTokenRef is nil",
		},
		{
			name: "outgoingAuth with system token ref",
			outgoingAuth: &mcpv1alpha1.OutgoingAuthConfig{
				Source: "discovered",
				SystemTokenRef: &mcpv1alpha1.SecretKeyRef{
					Name: "my-secret",
					Key:  "my-key",
				},
			},
			expectedEnv: "VMCP_SYSTEM_TOKEN",
			description: "Should set SystemTokenEnv to VMCP_SYSTEM_TOKEN when SystemTokenRef is present",
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

			if tt.outgoingAuth != nil {
				require.NotNil(t, config.OutgoingAuth, "OutgoingAuth should be set")
				assert.Equal(t, tt.expectedEnv, config.OutgoingAuth.SystemTokenEnv, tt.description)
			} else {
				// When nil in spec, it gets a default "discovered" config
				require.NotNil(t, config.OutgoingAuth)
				assert.Equal(t, "discovered", config.OutgoingAuth.Source)
				assert.Empty(t, config.OutgoingAuth.SystemTokenEnv)
			}
		})
	}
}

// Reuse test helpers from converter_test.go since they are in the same package (vmcpconfig)
// - newTestConverter
// - newNoOpMockResolver
// - newTestK8sClient
