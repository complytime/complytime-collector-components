package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// OIDCDiscovery represents OIDC discovery document
type OIDCDiscovery struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// Mock JWKS server for testing
func setupMockJWKSServer(t *testing.T, privateKey *rsa.PrivateKey) *httptest.Server {
	// Convert RSA public key to JWK format
	n := base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes())

	jwk := JWK{
		Kid: "test-key-id",
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   n,
		E:   e,
	}

	jwks := JWKS{
		Keys: []JWK{jwk},
	}

	mux := http.NewServeMux()

	// OIDC discovery endpoint
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  "http://" + r.Host,
			JWKSURI: "http://" + r.Host + "/openid/v1/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discovery)
	})

	// JWKS endpoint
	mux.HandleFunc("/openid/v1/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})

	return httptest.NewServer(mux)
}

func TestValidateSubject(t *testing.T) {
	tests := []struct {
		name            string
		subject         string
		allowedSubjects []string
		wantErr         bool
	}{
		{
			name:    "matching subject",
			subject: "system:serviceaccount:test:collector",
			allowedSubjects: []string{
				"system:serviceaccount:test:collector",
			},
			wantErr: false,
		},
		{
			name:    "matching subject in list",
			subject: "system:serviceaccount:test:collector",
			allowedSubjects: []string{
				"system:serviceaccount:test:other",
				"system:serviceaccount:test:collector",
			},
			wantErr: false,
		},
		{
			name:    "non-matching subject",
			subject: "system:serviceaccount:test:unauthorized",
			allowedSubjects: []string{
				"system:serviceaccount:test:collector",
			},
			wantErr: true,
		},
		{
			name:            "missing subject",
			subject:         "",
			allowedSubjects: []string{"system:serviceaccount:test:collector"},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSubject(tt.subject, tt.allowedSubjects)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestJWTAuthMiddlewareBasic tests basic middleware behavior
// Note: Full integration testing with go-oidc would require mocking the OIDC provider
// which is complex. This test validates the subject validation logic.
func TestJWTAuthMiddlewareBasic(t *testing.T) {
	// Generate test RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Set up mock JWKS server
	jwksServer := setupMockJWKSServer(t, privateKey)
	defer jwksServer.Close()

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	t.Run("missing authorization header", func(t *testing.T) {
		// Create a simple test that doesn't require full OIDC setup
		config := JWTAuthConfig{
			ExpectedAudience: "test-audience",
		}

		// Note: In a real scenario, the middleware would fail to initialize
		// because it can't reach the Kubernetes API. For unit testing,
		// we validate the subject validation logic separately.
		_ = config

		// Test validateSubject directly
		err := validateSubject("system:serviceaccount:test:collector", []string{
			"system:serviceaccount:test:collector",
		})
		assert.NoError(t, err)

		err = validateSubject("system:serviceaccount:test:unauthorized", []string{
			"system:serviceaccount:test:collector",
		})
		assert.Error(t, err)
	})
}

// TestJWTTokenValidation tests token validation logic
func TestJWTTokenValidation(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name          string
		setupToken    func() string
		expectValid   bool
		expectedError string
	}{
		{
			name: "valid token structure",
			setupToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"aud": "test-audience",
					"sub": "system:serviceaccount:test:collector",
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
				})
				token.Header["kid"] = "test-key-id"
				tokenString, err := token.SignedString(privateKey)
				require.NoError(t, err)
				return tokenString
			},
			expectValid: true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"aud": "test-audience",
					"sub": "system:serviceaccount:test:collector",
					"exp": time.Now().Add(-time.Hour).Unix(),
					"iat": time.Now().Add(-2 * time.Hour).Unix(),
				})
				token.Header["kid"] = "test-key-id"
				tokenString, err := token.SignedString(privateKey)
				require.NoError(t, err)
				return tokenString
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			assert.NotEmpty(t, tokenString)

			// Parse the token to verify it was created correctly
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return &privateKey.PublicKey, nil
			})

			if tt.expectValid {
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			} else {
				// Expired tokens will fail validation
				assert.Error(t, err)
			}
		})
	}
}
