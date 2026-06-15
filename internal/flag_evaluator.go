package internal

import "encoding/json"

// FlagEvaluator contains the evaluation logic for RemoteFlag values.
// It is stateless; all methods are pure functions of their inputs.
type FlagEvaluator struct{}

// NewFlagEvaluator returns a FlagEvaluator ready for use.
func NewFlagEvaluator() *FlagEvaluator {
	return &FlagEvaluator{}
}

// fnv1a32 computes the FNV-1a 32-bit hash of s, constrained to uint32 range.
func fnv1a32(s string) uint32 {
	const (
		offset32 uint32 = 0x811c9dc5
		prime32  uint32 = 0x01000193
	)
	h := offset32
	for _, b := range []byte(s) {
		h ^= uint32(b)
		h *= prime32
	}
	return h
}

// IsEnabled reports whether the flag should be considered active for the
// given targeting key, applying the flag's strategy rules.
func (e *FlagEvaluator) IsEnabled(flag RemoteFlag, targetingKey string) bool {
	if !flag.Enabled {
		return false
	}

	if flag.Strategy == "percentage" && flag.Rollout != nil {
		seed := flag.Key + ":" + targetingKey
		h := fnv1a32(seed)
		return int(h%100) < *flag.Rollout
	}

	return true
}

// ReasonFor returns the OpenFeature reason string for a flag evaluation.
func (e *FlagEvaluator) ReasonFor(flag RemoteFlag) string {
	if !flag.Enabled {
		return "DISABLED"
	}
	if flag.Strategy == "percentage" {
		return "SPLIT"
	}
	return "STATIC"
}

// BooleanValue evaluates the flag as a boolean.
// If the active JSON node is a bool it is returned directly; otherwise
// the result of IsEnabled is returned as the boolean value.
func (e *FlagEvaluator) BooleanValue(flag RemoteFlag, defaultValue bool, targetingKey string) bool {
	on := e.IsEnabled(flag, targetingKey)

	node := flag.FallbackValue
	if on {
		node = flag.Value
	}

	if len(node) > 0 {
		var b bool
		if err := json.Unmarshal(node, &b); err == nil {
			return b
		}
	}

	return on
}

// StringValue evaluates the flag as a string.
// If the active JSON node is null or missing the default is returned.
// If it is a JSON string the text is returned; otherwise the raw JSON
// representation is returned.
func (e *FlagEvaluator) StringValue(flag RemoteFlag, defaultValue string, targetingKey string) string {
	on := e.IsEnabled(flag, targetingKey)

	node := flag.FallbackValue
	if on {
		node = flag.Value
	}

	if len(node) == 0 || string(node) == "null" {
		return defaultValue
	}

	var s string
	if err := json.Unmarshal(node, &s); err == nil {
		return s
	}

	// Fallback: return the raw JSON as a string.
	return string(node)
}

// IntValue evaluates the flag as an integer.
// Returns defaultValue when the active node is null, missing, or not a number.
func (e *FlagEvaluator) IntValue(flag RemoteFlag, defaultValue int64, targetingKey string) int64 {
	on := e.IsEnabled(flag, targetingKey)

	node := flag.FallbackValue
	if on {
		node = flag.Value
	}

	if len(node) == 0 || string(node) == "null" {
		return defaultValue
	}

	var n json.Number
	if err := json.Unmarshal(node, &n); err != nil {
		return defaultValue
	}

	i, err := n.Int64()
	if err != nil {
		return defaultValue
	}

	return i
}

// FloatValue evaluates the flag as a float64.
// Returns defaultValue when the active node is null, missing, or not a number.
func (e *FlagEvaluator) FloatValue(flag RemoteFlag, defaultValue float64, targetingKey string) float64 {
	on := e.IsEnabled(flag, targetingKey)

	node := flag.FallbackValue
	if on {
		node = flag.Value
	}

	if len(node) == 0 || string(node) == "null" {
		return defaultValue
	}

	var n json.Number
	if err := json.Unmarshal(node, &n); err != nil {
		return defaultValue
	}

	f, err := n.Float64()
	if err != nil {
		return defaultValue
	}

	return f
}

// ObjectValue evaluates the flag as an arbitrary object (map or slice).
// Returns defaultValue when the active node is null or missing.
// Otherwise it returns the decoded Go value (map[string]interface{}, []interface{}, etc.).
func (e *FlagEvaluator) ObjectValue(flag RemoteFlag, defaultValue interface{}, targetingKey string) interface{} {
	on := e.IsEnabled(flag, targetingKey)

	node := flag.FallbackValue
	if on {
		node = flag.Value
	}

	if len(node) == 0 || string(node) == "null" {
		return defaultValue
	}

	var v interface{}
	if err := json.Unmarshal(node, &v); err != nil {
		return string(node)
	}

	return v
}
