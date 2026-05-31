// Package auth provides authentication for Google APIs via omnitoken.
package auth

import (
	"context"
	"net/http"

	"github.com/grokify/goauth"
	"github.com/grokify/goauth/google"
	"github.com/plexusone/omnitoken"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/sheets/v4"
	slides "google.golang.org/api/slides/v1"
)

// Scopes returns the OAuth2 scopes required for read-only access to Google Slides, Docs, and Sheets.
func Scopes() []string {
	return []string{
		slides.PresentationsReadonlyScope,
		docs.DocumentsReadonlyScope,
		sheets.SpreadsheetsReadonlyScope,
		slides.DriveReadonlyScope,
	}
}

// NewClient creates an authenticated HTTP client using a Google service account credentials file.
// This uses the standard Google Cloud service account JSON format.
// Deprecated: Use NewTokenManager and TokenManager.GetClient instead.
func NewClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	return google.NewClientSvcAccountFromFile(ctx, credentialsFile, Scopes()...)
}

// NewClientFromCredentialsSet creates an authenticated HTTP client using a goauth CredentialsSet file.
// The CredentialsSet should contain a credential entry with the specified account key.
// The credential entry should be of type "gcpsa" with appropriate scopes configured.
// Deprecated: Use NewTokenManager and TokenManager.GetClient instead.
func NewClientFromCredentialsSet(ctx context.Context, credentialsFile, accountKey string) (*http.Client, error) {
	return goauth.NewClient(ctx, credentialsFile, accountKey)
}

// NewTokenManager creates an omnitoken.TokenManager configured for Google Slides access.
// It supports multiple credential sources:
//   - Google service account JSON file (serviceAccountFile)
//   - goauth CredentialsSet file with account key (goauthFile, goauthKey)
//   - Vault URI for vault-backed credentials (vaultURI, credentialsName)
//
// Only one credential source should be provided.
func NewTokenManager(ctx context.Context, opts TokenManagerOptions) (*omnitoken.TokenManager, error) {
	switch {
	case opts.VaultURI != "":
		// Vault-backed credentials
		mgr, err := omnitoken.NewFromVaultURI(opts.VaultURI)
		if err != nil {
			return nil, err
		}
		return mgr, nil

	case opts.GoauthFile != "":
		// goauth CredentialsSet file
		mgr, err := omnitoken.NewFromCredentialsFile(opts.GoauthFile)
		if err != nil {
			return nil, err
		}
		return mgr, nil

	case opts.ServiceAccountFile != "":
		// Google service account JSON file
		mgr, err := omnitoken.NewFromCredentials(opts.CredentialsName, nil)
		if err != nil {
			return nil, err
		}
		if err := mgr.LoadGoogleServiceAccount(ctx, opts.CredentialsName, opts.ServiceAccountFile, Scopes()); err != nil {
			return nil, err
		}
		return mgr, nil

	default:
		// Auto-detect from environment
		return omnitoken.NewAuto()
	}
}

// TokenManagerOptions configures how to create a TokenManager.
type TokenManagerOptions struct {
	// ServiceAccountFile is the path to a Google service account JSON file.
	ServiceAccountFile string

	// GoauthFile is the path to a goauth CredentialsSet JSON file.
	GoauthFile string

	// GoauthKey is the account key within the goauth CredentialsSet.
	GoauthKey string

	// VaultURI is the URI for vault-backed credentials (e.g., "op://vault", "file:///path").
	VaultURI string

	// CredentialsName is the name to use when storing/retrieving credentials.
	// Defaults to "google-slides" if not specified.
	CredentialsName string
}

// GetClient returns an authenticated HTTP client from the TokenManager.
// This is a convenience function that retrieves the client for the configured credentials.
func GetClient(ctx context.Context, mgr *omnitoken.TokenManager, credentialsName string) (*http.Client, error) {
	if credentialsName == "" {
		credentialsName = "google-slides"
	}
	return mgr.GetClient(ctx, credentialsName)
}
