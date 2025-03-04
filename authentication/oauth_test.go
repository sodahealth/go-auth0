package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/go-auth0/authentication/oauth"
)

func TestOAuthLoginWithPassword(t *testing.T) {
	t.Run("Should return tokens", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithPassword(context.Background(), oauth.LoginWithPasswordRequest{
			Username: "testuser",
			Password: "testuser123",
		}, oauth.IDTokenValidationOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})

	t.Run("Should support passing extra options", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithPassword(context.Background(), oauth.LoginWithPasswordRequest{
			Username: "testuser",
			Password: "testuser123",
			Realm:    "my-realm",
			Scope:    "extra-scope",
			ExtraParameters: map[string]string{
				"extra": "value",
			},
		}, oauth.IDTokenValidationOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})
}

func TestLoginWithAuthCode(t *testing.T) {
	t.Run("Should require client_secret", func(t *testing.T) {
		_, err := authAPI.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{})

		assert.Error(t, err, "client_secret is required but not provided")
	})

	t.Run("Should throw for an invalid code", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		_, err := authAPI.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			ClientAuthentication: oauth.ClientAuthentication{
				ClientSecret: clientSecret,
			},
			Code: "my-invalid-code",
		}, oauth.IDTokenValidationOptions{})

		assert.Error(t, err, "Invalid authorization code")
	})

	t.Run("Should return tokens", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			ClientAuthentication: oauth.ClientAuthentication{
				ClientSecret: clientSecret,
			},
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})

	t.Run("Should support setting a redirect uri", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			ClientAuthentication: oauth.ClientAuthentication{
				ClientSecret: clientSecret,
			},
			Code:        "test-code",
			RedirectURI: "http://localhost:3000",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})
}

func TestLoginWithAuthCodeWithPKCE(t *testing.T) {
	t.Run("Should throw for an invalid code", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		_, err := authAPI.OAuth.LoginWithAuthCodeWithPKCE(context.Background(), oauth.LoginWithAuthCodeWithPKCERequest{
			Code:         "test-invalid-code",
			CodeVerifier: "test-code-verifier",
		}, oauth.IDTokenValidationOptions{})

		assert.Error(t, err, "Invalid authorization code")
	})

	t.Run("Should throw for an invalid code verifier", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		_, err := authAPI.OAuth.LoginWithAuthCodeWithPKCE(context.Background(), oauth.LoginWithAuthCodeWithPKCERequest{
			Code:         "test-code",
			CodeVerifier: "test-invalid-code-verifier",
		}, oauth.IDTokenValidationOptions{})

		assert.Error(t, err, "Failed to verify code verifier")
	})

	t.Run("Should return tokens", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithAuthCodeWithPKCE(context.Background(), oauth.LoginWithAuthCodeWithPKCERequest{
			Code:         "test-code",
			CodeVerifier: "test-code-verifier",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})

	t.Run("Should support setting a redirect uri", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithAuthCodeWithPKCE(context.Background(), oauth.LoginWithAuthCodeWithPKCERequest{
			Code:         "test-code",
			CodeVerifier: "test-code-verifier",
			RedirectURI:  "http://localhost:3000",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})
}

func TestLoginWithClientCredentials(t *testing.T) {
	t.Run("Should require client_secret", func(t *testing.T) {
		_, err := authAPI.OAuth.LoginWithClientCredentials(context.Background(), oauth.LoginWithClientCredentialsRequest{
			Audience: "test-audience",
		}, oauth.IDTokenValidationOptions{})

		assert.Error(t, err, "client_secret is required but not provided")
	})

	t.Run("Should return tokens", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithClientCredentials(context.Background(), oauth.LoginWithClientCredentialsRequest{
			ClientAuthentication: oauth.ClientAuthentication{
				ClientSecret: clientSecret,
			},
			Audience: "test-audience",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})

	t.Run("Should allow overriding clientid", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.LoginWithClientCredentials(context.Background(), oauth.LoginWithClientCredentialsRequest{
			ClientAuthentication: oauth.ClientAuthentication{
				ClientSecret: clientSecret,
				ClientID:     "test-other-clientid",
			},
			Audience: "test-audience",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("Should return tokens", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.RefreshToken(context.Background(), oauth.RefreshTokenRequest{
			RefreshToken: "test-refresh-token",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.NotEmpty(t, tokenSet.RefreshToken)
		assert.NotEmpty(t, tokenSet.Scope)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})

	t.Run("Should return tokens with reduced scopes", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		tokenSet, err := authAPI.OAuth.RefreshToken(context.Background(), oauth.RefreshTokenRequest{
			RefreshToken: "test-refresh-token",
			Scope:        "openid profile offline_access",
		}, oauth.IDTokenValidationOptions{})

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenSet.AccessToken)
		assert.NotEmpty(t, tokenSet.RefreshToken)
		assert.Equal(t, "openid profile offline_access", tokenSet.Scope)
		assert.Equal(t, "Bearer", tokenSet.TokenType)
	})
}

func TestRevokeRefreshToken(t *testing.T) {
	t.Run("Should revoke token", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		err := authAPI.OAuth.RevokeRefreshToken(context.Background(), oauth.RevokeRefreshTokenRequest{
			Token: "test-refresh-token",
		})

		assert.NoError(t, err)
	})

	t.Run("Should support passing a ClientID and ClientSecret", func(t *testing.T) {
		configureHTTPTestRecordings(t)

		auth, err := New(
			context.Background(),
			domain,
			WithClientID(clientID),
			WithClientSecret(clientSecret),
			WithIDTokenSigningAlg("HS256"),
		)
		assert.NoError(t, err)

		err = auth.OAuth.RevokeRefreshToken(context.Background(), oauth.RevokeRefreshTokenRequest{
			Token: "test-refresh-token",
		})
		assert.NoError(t, err)
	})
}

func TestWithIDTokenVerification(t *testing.T) {
	t.Run("error for an invalid organization when using org_id", func(t *testing.T) {
		extras := map[string]interface{}{
			"org_id": "org_123",
		}
		api, err := withIDToken(t, extras)
		assert.NoError(t, err)

		_, err = api.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{Organization: "org_456"})

		assert.ErrorContains(t, err, "org_id claim value mismatch in the ID token")
	})

	t.Run("error for an invalid organization when using org_name", func(t *testing.T) {
		extras := map[string]interface{}{
			"org_name": "wrong-org",
		}
		api, err := withIDToken(t, extras)
		assert.NoError(t, err)

		_, err = api.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{Organization: "right-org"})

		assert.ErrorContains(t, err, "org_name claim value mismatch in the ID token")
	})

	t.Run("error for an invalid nonce", func(t *testing.T) {
		extras := map[string]interface{}{
			"nonce": "wrong-nonce",
		}
		api, err := withIDToken(t, extras)
		assert.NoError(t, err)

		_, err = api.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{Nonce: "test-nonce"})

		assert.ErrorContains(t, err, "nonce claim value mismatch in the ID token; expected")
	})

	t.Run("error for an invalid maxage", func(t *testing.T) {
		extras := map[string]interface{}{
			"auth_time": time.Now().Add(-500 * time.Second).Unix(),
		}
		api, err := withIDToken(t, extras)
		assert.NoError(t, err)

		_, err = api.OAuth.LoginWithAuthCode(context.Background(), oauth.LoginWithAuthCodeRequest{
			Code: "my-code",
		}, oauth.IDTokenValidationOptions{MaxAge: 100 * time.Second})

		assert.ErrorContains(t, err, "auth_time claim in the ID token indicates that too much time has passed")
	})
}

func withIDToken(t *testing.T, extras map[string]interface{}) (*Authentication, error) {
	t.Helper()

	idTokenClientSecret := "test-client-secret"
	idTokenClientid := "test-client-id"

	var idToken string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenSet := &oauth.TokenSet{
			AccessToken: "test-access-token",
			ExpiresIn:   86400,
			IDToken:     idToken,
			TokenType:   "Bearer",
		}

		b, err := json.Marshal(tokenSet)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(b))
	})
	s := httptest.NewTLSServer(h)
	t.Cleanup(func() {
		s.Close()
	})

	URL, err := url.Parse(s.URL)
	assert.NoError(t, err)

	api, err := New(
		context.Background(),
		URL.Host,
		WithClient(s.Client()),
		WithClientID(idTokenClientid),
		WithClientSecret(idTokenClientSecret),
		WithIDTokenSigningAlg("HS256"),
	)
	assert.NoError(t, err)

	builder := jwt.NewBuilder().
		Issuer(s.URL + "/").
		Subject("me").
		Audience([]string{idTokenClientid}).
		Expiration(time.Now().Add(time.Hour)).
		IssuedAt(time.Now().Add(-time.Hour))

	for claim, value := range extras {
		builder.Claim(claim, value)
	}

	token, err := builder.Build()

	if err != nil {
		return nil, err
	}

	b, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, []byte(idTokenClientSecret)))
	if err != nil {
		return nil, err
	}

	idToken = string(b)

	return api, nil
}
