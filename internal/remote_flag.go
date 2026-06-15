package internal

import "encoding/json"

// RemoteFlag represents a feature flag fetched from the TrustyGorilla API.
// The "key" field is populated from the map key after JSON parsing, not from
// the JSON body itself.
type RemoteFlag struct {
	Key             string          `json:"-"`
	Enabled         bool            `json:"enabled"`
	Strategy        string          `json:"strategy"`
	Description     string          `json:"description"`
	Tags            []string        `json:"tags"`
	Rollout         *int            `json:"rollout"`
	Value           json.RawMessage `json:"value"`
	FallbackValue   json.RawMessage `json:"fallback_value"`
	TargetingRules  json.RawMessage `json:"targeting_rules"`
	Variants        json.RawMessage `json:"variants"`
}
