// Package auth handles interactive credential collection and server authentication.
//
// Security invariants:
//   - Password and Token are read with echo disabled (FR-AUTH-002).
//   - All failure cases (wrong credentials, wrong role, inactive account,
//     network error) return the same single message "Not authorized" (FR-AUTH-004).
//   - Credentials are never stored to disk (FR-AUTH-005).
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/client"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/secureinput"
)

// authRequest matches POST /download/auth JSON body.
type authRequest struct {
	UserKey  string `json:"user_key"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// authResponse matches POST /download/auth success JSON body.
type authResponse struct {
	Token string `json:"token"`
}

// errNotAuthorized is the single user-visible authentication failure message.
var errNotAuthorized = errors.New("Not authorized")

// Authenticate prompts interactively for User Key, Password, and Token,
// calls POST /download/auth, and returns the Bearer token on success.
// Any failure (wrong credentials, insufficient role, inactive user, network
// error) results in errNotAuthorized so the exact reason is never disclosed.
func Authenticate(ctx context.Context, c *client.Client) (string, error) {
	userKey, err := secureinput.ReadLine("User Key: ")
	if err != nil {
		return "", errNotAuthorized
	}
	if strings.TrimSpace(userKey) == "" {
		return "", errNotAuthorized
	}

	password, err := secureinput.ReadMasked("Password: ")
	if err != nil {
		return "", errNotAuthorized
	}
	if password == "" {
		return "", errNotAuthorized
	}

	token, err := secureinput.ReadMasked("Token: ")
	if err != nil {
		return "", errNotAuthorized
	}
	if token == "" {
		return "", errNotAuthorized
	}

	var resp authResponse
	err = c.PostJSON(ctx, "/download/auth", authRequest{
		UserKey:  userKey,
		Password: password,
		Token:    token,
	}, &resp)

	// Map every failure to the same message (SEC-007 / FR-AUTH-004).
	if err != nil {
		// Propagate context cancellation so the caller can handle Ctrl+C.
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		return "", errNotAuthorized
	}
	if resp.Token == "" {
		return "", errNotAuthorized
	}

	fmt.Println("Authenticated.")
	return resp.Token, nil
}
