package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"k8s.io/client-go/rest"
)

// JWTAuthConfig holds the configuration for JWT authentication
type JWTAuthConfig struct {
	// ExpectedAudience is the expected audience claim in the token
	ExpectedAudience string
	// AllowedSubjects is an optional list of allowed subject claims
	// If empty, any subject is allowed
	AllowedSubjects []string
}

// JWTAuthMiddleware creates a Gin middleware that validates bound service account tokens
// using standard Kubernetes OIDC verification with go-oidc library
func JWTAuthMiddleware(config JWTAuthConfig) gin.HandlerFunc {
	// Get in-cluster Kubernetes config
	// This handles TLS cert loading, service account tokens, etc.
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("failed to get in-cluster config", "error", err)
		// Fallback: try to continue without client-go configuration
		k8sConfig = &rest.Config{
			Host: "https://kubernetes.default.svc",
		}
	}

	// Create HTTP client using Kubernetes configuration
	httpClient, err := rest.HTTPClientFor(k8sConfig)
	if err != nil {
		slog.Error("failed to create HTTP client", "error", err)
		httpClient = http.DefaultClient
	}

	// Check if DNS bypass is enabled
	kubernetesServiceIP := os.Getenv("KUBERNETES_SERVICE_IP")
	dnsBypassEnabled := kubernetesServiceIP != ""

	// Apply DNS bypass if KUBERNETES_SERVICE_IP is set
	// This is the only custom part - override DialContext to use direct IP
	if dnsBypassEnabled {
		slog.Info("DNS bypass enabled - using direct Kubernetes API IP", "kubernetes_ip", kubernetesServiceIP)

		// Get the base transport from the client
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			// Clone the transport to avoid modifying the original
			customTransport := transport.Clone()

			// Override DialContext to replace DNS lookup with direct IP
			customTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Replace kubernetes.default.svc hostname with direct IP
				if strings.Contains(addr, "kubernetes.default.svc") {
					addr = strings.Replace(addr, "kubernetes.default.svc", kubernetesServiceIP, 1)
					slog.Debug("DNS bypass: connecting directly to Kubernetes API", "addr", addr)
				}

				// Use standard dialer
				dialer := &net.Dialer{
					Timeout:   15 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				return dialer.DialContext(ctx, network, addr)
			}

			httpClient.Transport = customTransport
		}
	}

	// Always use the standard Kubernetes hostname for OIDC issuer
	// This ensures consistency with what the OIDC discovery endpoint returns
	issuerURL := "https://kubernetes.default.svc"

	// Create OIDC provider with Kubernetes client
	ctx := oidc.ClientContext(context.Background(), httpClient)
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		slog.Error("failed to create OIDC provider", "error", err)
		// Return middleware that always fails
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "JWT authentication not available",
			})
		}
	}

	// Create token verifier configuration
	verifierConfig := &oidc.Config{
		ClientID: config.ExpectedAudience,
	}

	// When DNS bypass is enabled, skip issuer validation
	// This is safe because:
	// 1. We're in a trusted environment (inside the cluster)
	// 2. TLS certificate validation is still performed
	// 3. JWKS signature verification is still performed
	if dnsBypassEnabled {
		verifierConfig.SkipIssuerCheck = true
		slog.Info("OIDC issuer validation disabled due to DNS bypass")
	}

	// Create token verifier
	verifier := provider.Verifier(verifierConfig)

	slog.Info("JWT authentication middleware initialized",
		"issuer", issuerURL,
		"audience", config.ExpectedAudience,
		"dns_bypass", dnsBypassEnabled)

	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("missing authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			slog.Warn("invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		rawToken := parts[1]

		// Verify token using go-oidc
		// This handles:
		// - JWKS fetching and caching
		// - Signature verification
		// - Audience validation
		// - Expiration validation
		idToken, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			slog.Warn("token verification failed", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
			return
		}

		slog.Info("jwt signature and standard claims verified successfully")

		// Extract claims for additional validation
		var claims jwt.MapClaims
		if err := idToken.Claims(&claims); err != nil {
			slog.Warn("failed to extract claims", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "failed to extract token claims",
			})
			return
		}

		// Validate subject (service account) if configured
		if len(config.AllowedSubjects) > 0 {
			subject, ok := claims["sub"].(string)
			if !ok {
				slog.Warn("subject claim missing or invalid")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "subject claim missing",
				})
				return
			}

			if err := validateSubject(subject, config.AllowedSubjects); err != nil {
				slog.Warn("subject validation failed", "error", err, "subject", subject)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": fmt.Sprintf("subject validation failed: %v", err),
				})
				return
			}
			slog.Info("jwt subject validation successful")
		}

		// Store claims in context for downstream handlers
		c.Set("jwt_claims", claims)

		slog.Info("jwt authentication successful")
		c.Next()
	}
}

// validateSubject checks if the subject is in the allowed list
func validateSubject(subject string, allowedSubjects []string) error {
	for _, allowed := range allowedSubjects {
		if subject == allowed {
			return nil
		}
	}
	return fmt.Errorf("subject %q not in allowed list", subject)
}
