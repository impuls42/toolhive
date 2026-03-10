# ToolHive MCP Gateway Implementation Plan

This document outlines the planned Custom Resources (CRDs) and implementation steps required to configure ToolHive as the Main MCP Gateway, assuming the use of long-lived JWTs for all clients (React WebUI, AI Platform, 3rd Party) and postponed RBAC.

## 1. Architecture Simplifications (Zero Code Changes)

By issuing long-lived JWT tokens for the AI Assistant Platform and using an external Token Exchange Service, we can achieve the exact architecture you want **without editing the ToolHive codebase at all**:
*   **Incoming Auth:** All clients (WebUI, AI Platform, 3rd Party) use standard `Bearer <JWT>` tokens. ToolHive natively handles this via its `oidc` incoming auth configuration.
*   **Authorization (RBAC):** Since RBAC is postponed, we simply omit the Cedar policy configuration, allowing all authenticated requests to access discovered tools.
*   **Outgoing Auth (All Services):** We will use ToolHive's native `token_exchange` strategy for **all** backend services that require auth. We will point ToolHive to your external Token Exchange Service, but vary the `audience` parameter in the CRD:
    *   For **GitHub**, the Token Exchange Service will mint a GitHub App installation token.
    *   For **Grafana and Cluster Gateways**, the Token Exchange Service will simply validate the incoming JWT and return it back to ToolHive (acting as a passthrough), or mint a fresh internal standard JWT. 
    *   ToolHive will take whatever token the service returns, cache it, and inject it as `Authorization: Bearer <TOKEN>` to the downstream service.

---

## 2. Expected ToolHive Custom Resources (CRDs)

To achieve this setup, we will define the following custom resources in Kubernetes.

### A. MCPGroup (`main-gateway-backends`)
This resource groups the downstream backends the gateway routes to.

```yaml
apiVersion: toolhive.stacklok.com/v1alpha1
kind: MCPGroup
metadata:
  name: main-gateway-backends
spec:
  # Examples of expected backends:
  # - name: grafana
  #   url: http://grafana-mcp...
  # - name: cluster-a
  #   url: http://cluster-a-gw...
  # - name: github
  #   url: http://github-mcp...
```

### B. VirtualMCPServer (`main-gateway`)
This is the core gateway configuration mapping incoming auth, routing prefixes, and outgoing auth.

**Key Parameters:**

*   **`spec.groupRef`**: `main-gateway-backends`
*   **`spec.incomingAuth`**:
    *   `type: oidc` 
    *   `oidc.issuer`: The URL of your Hydra/OIDC provider.
    *   *(Note: Authz/Cedar policies are omitted since RBAC is postponed).*
*   **`spec.aggregation`**:
    *   `conflictResolution: manual` or `prefix`. Maps namespaces (e.g., `grafana:*`, `github:*`) to their respective backends.
*   **`spec.outgoingAuth`**:
    *   **For `grafana-mcp` and `cluster-gw`:** 
        *   `type: token_exchange`
        *   `tokenExchange.tokenUrl`: URL of your custom external Token Exchange Service.
        *   `tokenExchange.audience`: `internal-passthrough` (Signals your service to just return the JWT).
    *   **For `github-mcp`:** 
        *   `type: token_exchange`
        *   `tokenExchange.tokenUrl`: URL of your custom external Token Exchange Service.
        *   `tokenExchange.audience`: `github-mcp` (Signals your service to mint a GitHub App token).

---

## 3. Implementation Scope for Subagents

Since no code changes are needed in ToolHive, the scope is purely configuration and deployment.

1.  **[kubernetes-expert] Define CRDs and Manifests:**
    *   Author the `MCPGroup` and `VirtualMCPServer` CRDs based on the simplified parameters above (OIDC inbound, `token_exchange` outbound for all secured backends).
2.  **[mcp-protocol-expert] Configure Token Exchange:**
    *   Deploy a mock or baseline Token Exchange Service that handles the logic described above:
        *   If `audience=github-mcp`, mock the return of a GitHub App token.
        *   If `audience=internal-passthrough`, return the `subject_token` directly in the `access_token` response field.