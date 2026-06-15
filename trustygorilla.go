// Package trustygorilla provides a TrustyGorilla feature-flag provider for
// the OpenFeature Go SDK.
package trustygorilla

import "fmt"

const (
	defaultAPIURL              = "https://api.trustygorilla.com"
	defaultPollIntervalSeconds = 30
)

// TrustyGorillaConfig holds the settings needed to connect to the
// TrustyGorilla feature-flag service.
type TrustyGorillaConfig struct {
	// APIURL is the base URL of the TrustyGorilla API.
	// Defaults to "https://api.trustygorilla.com".
	APIURL string

	// APIKey is the bearer token used to authenticate API requests.
	// Required; NewTrustyGorillaConfig panics if empty.
	APIKey string

	// PollIntervalSeconds controls how often flags are refreshed from the API.
	// Defaults to 30.
	PollIntervalSeconds int
}

// NewTrustyGorillaConfig constructs a TrustyGorillaConfig, applying defaults
// for optional fields. It returns an error when apiKey is empty.
func NewTrustyGorillaConfig(apiKey string) (*TrustyGorillaConfig, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("trustygorilla: apiKey must not be empty")
	}
	return &TrustyGorillaConfig{
		APIURL:              defaultAPIURL,
		APIKey:              apiKey,
		PollIntervalSeconds: defaultPollIntervalSeconds,
	}, nil
}
