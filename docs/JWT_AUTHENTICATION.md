# JWT Authentication Middleware

## Overview

Compass uses a Gin middleware for JWT authentication that validates Kubernetes bound service account tokens. The implementation uses the industry-standard `coreos/go-oidc` library with `k8s.io/client-go` for OIDC-compliant token verification.

## How It Works

### Middleware Initialization

On startup, the middleware:

1. **Loads Kubernetes configuration** using `rest.InClusterConfig()`
   - Reads TLS certificates and service account credentials
   - Configures connection to Kubernetes API server

2. **Sets up DNS bypass** (if `KUBERNETES_SERVICE_IP` or `KUBERNETES_SERVICE_HOST` is set)
   - Redirects all connections to the specified IP address
   - Maintains TLS security with certificate validation

3. **Initializes OIDC provider** from `https://kubernetes.default.svc`
   - Fetches OIDC discovery document
   - Loads JWKS (JSON Web Key Set) for signature verification
   - Retries 3 times with exponential backoff (1s, 2s, 4s) if failed

4. **Creates token verifier** with:
   - Expected audience (ClientID)
   - Issuer validation (skipped if DNS bypass is enabled)

### Request Validation

For each incoming request, the middleware:

1. Extracts the JWT from the `Authorization: Bearer <token>` header
2. Verifies the token using `go-oidc`
3. Validates the subject claim (if `AllowedSubjects` is configured)
4. Stores claims in the Gin context for downstream handlers

## Security: What Is Verified

The middleware performs the following security validations:

### 1. Cryptographic Signature
- Token signature verified using RSA public keys from Kubernetes JWKS
- Ensures token was issued by the Kubernetes API server
- Prevents token forgery

### 2. Audience Claim
- Validates `aud` claim matches `ExpectedAudience` configuration
- Ensures token is intended for this specific service
- Prevents token reuse across different services

### 3. Expiration
- Validates `exp` claim to reject expired tokens
- Tokens have 1-hour lifetime (configurable by Kubernetes)
- Kubernetes automatically rotates tokens before expiration

### 4. Issuer Claim *
- Validates `iss` claim matches the Kubernetes issuer
- ***Skipped when DNS bypass is enabled*** (`SkipIssuerCheck: true`)
- Safe to skip because signature, audience, and expiration are still validated

### 5. Subject Claim (optional)
- Validates `sub` claim against `AllowedSubjects` list
- Subject format: `system:serviceaccount:<namespace>:<serviceaccount-name>`
- Only specified service accounts can authenticate

### 6. TLS Security
- All connections use TLS with CA certificate validation
- Certificate loaded from `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`
- No insecure configurations (no `InsecureSkipVerify`)

## Configuration

### Basic Configuration

```go
middleware := JWTAuthMiddleware(JWTAuthConfig{
    ExpectedAudience: "compass-internal",
})
```

### With Subject Validation

```go
middleware := JWTAuthMiddleware(JWTAuthConfig{
    ExpectedAudience: "compass-internal",
    AllowedSubjects: []string{
        "system:serviceaccount:default:collector",
    },
})
```

### DNS Bypass Mode

Set environment variable to bypass DNS resolution issues:

```yaml
env:
  - name: KUBERNETES_SERVICE_IP
    value: "$IP"  # Kubernetes API ClusterIP
```

Find your cluster's IP:
```bash
kubectl get svc kubernetes -n default -o jsonpath='{.spec.clusterIP}'
```

## Client Token Configuration

Clients must include a bound service account token in requests:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: collector
spec:
  serviceAccountName: collector
  containers:
  - name: collector
    volumeMounts:
    - name: bound-token
      mountPath: /var/run/secrets/tokens
  volumes:
  - name: bound-token
    projected:
      sources:
      - serviceAccountToken:
          path: token
          expirationSeconds: 3600
          audience: compass-internal  # Must match ExpectedAudience
```

Client code:
```go
token, _ := os.ReadFile("/var/run/secrets/tokens/token")
req.Header.Set("Authorization", "Bearer "+string(token))
```

## Error Responses

- `401 Unauthorized` - Missing or invalid token
- `503 Service Unavailable` - OIDC provider initialization failed

## Troubleshooting

### OIDC Provider Initialization Failed

Check logs for:
```
level=ERROR msg="fatal: failed to create OIDC provider after retries"
```

**Solutions:**
- Verify network connectivity to Kubernetes API server
- Check if service account CA certificate is mounted correctly
- Enable DNS bypass mode if DNS resolution is failing

### Token Verification Failed

Check logs for:
```
level=WARN msg="token verification failed" error="..."
```

**Common causes:**
- Token expired (Kubernetes should auto-rotate)
- Audience mismatch (check both sides match)
- Invalid signature (token not from Kubernetes API)
- Subject not in AllowedSubjects list

### Issuer Validation Errors (DNS Bypass)

```
error="oidc: issuer URL ... did not match ..."
```
