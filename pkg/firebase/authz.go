package firebase

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/firebasehosting/v1beta1"
)

// AuthFile relative to source dir.
const AuthFile = "private/creds/blog-releaser.json"

type ServiceAccountCreds struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientID     string `json:"client_id"`
	AuthURI      string `json:"auth_uri"`
	TokenURI     string `json:"token_uri"`
}

// ReadServiceAccountCreds reads the credentials file for the Firebase service
// account.
func ReadServiceAccountCreds() (s ServiceAccountCreds, mErr error) {
	b, err := os.ReadFile(AuthFile)
	if err != nil {
		return s, fmt.Errorf("read service account creds: %w", err)
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("unmarshal service account creds: %w", err)
	}
	return s, nil
}

// NewTokenSource creates a token source for the service account credentials.
func NewTokenSource(ctx context.Context, accountCreds ServiceAccountCreds) oauth2.TokenSource {
	cfg := &jwt.Config{
		Email:      accountCreds.ClientEmail,
		PrivateKey: []byte(accountCreds.PrivateKey),
		Scopes:     []string{firebasehosting.FirebaseScope},
		TokenURL:   google.JWTTokenURL,
	}
	tokSource := cfg.TokenSource(ctx)
	return tokSource
}
