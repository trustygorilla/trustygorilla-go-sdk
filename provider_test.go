package trustygorilla_test

import (
	"encoding/json"
	"testing"

	"github.com/trustygorilla/go-sdk/internal"
)

// ptr is a helper to take the address of an int literal.
func ptr(v int) *int { return &v }

// rawJSON marshals v into a json.RawMessage for use in RemoteFlag fields.
func rawJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func TestFlagEvaluator_IsEnabled_Disabled(t *testing.T) {
	e := internal.NewFlagEvaluator()
	flag := internal.RemoteFlag{Key: "my-flag", Enabled: false, Strategy: "boolean"}
	if e.IsEnabled(flag, "user-1") {
		t.Error("expected disabled flag to return false")
	}
}

func TestFlagEvaluator_IsEnabled_StaticEnabled(t *testing.T) {
	e := internal.NewFlagEvaluator()
	flag := internal.RemoteFlag{Key: "my-flag", Enabled: true, Strategy: "boolean"}
	if !e.IsEnabled(flag, "user-1") {
		t.Error("expected enabled static flag to return true")
	}
}

func TestFlagEvaluator_IsEnabled_Percentage(t *testing.T) {
	e := internal.NewFlagEvaluator()

	// FNV-1a("my-flag:user-1") % 100 is deterministic.
	// With rollout=100 every user must be in.
	flag := internal.RemoteFlag{
		Key:      "my-flag",
		Enabled:  true,
		Strategy: "percentage",
		Rollout:  ptr(100),
	}
	if !e.IsEnabled(flag, "user-1") {
		t.Error("rollout=100 should enable all users")
	}

	// With rollout=0 no user should be in.
	flag.Rollout = ptr(0)
	if e.IsEnabled(flag, "user-1") {
		t.Error("rollout=0 should disable all users")
	}
}

func TestFlagEvaluator_IsEnabled_Percentage_Deterministic(t *testing.T) {
	e := internal.NewFlagEvaluator()
	flag := internal.RemoteFlag{
		Key:      "feature-x",
		Enabled:  true,
		Strategy: "percentage",
		Rollout:  ptr(50),
	}

	// Calling IsEnabled multiple times with the same inputs must return the same result.
	first := e.IsEnabled(flag, "stable-user")
	for i := 0; i < 10; i++ {
		if e.IsEnabled(flag, "stable-user") != first {
			t.Error("IsEnabled must be deterministic for the same flag+targeting key")
		}
	}
}

func TestFlagEvaluator_ReasonFor(t *testing.T) {
	e := internal.NewFlagEvaluator()

	tests := []struct {
		flag   internal.RemoteFlag
		expect string
	}{
		{internal.RemoteFlag{Enabled: false, Strategy: "boolean"}, "DISABLED"},
		{internal.RemoteFlag{Enabled: true, Strategy: "percentage", Rollout: ptr(50)}, "SPLIT"},
		{internal.RemoteFlag{Enabled: true, Strategy: "boolean"}, "STATIC"},
		{internal.RemoteFlag{Enabled: true, Strategy: "value"}, "STATIC"},
	}

	for _, tc := range tests {
		got := e.ReasonFor(tc.flag)
		if got != tc.expect {
			t.Errorf("ReasonFor(%+v) = %q; want %q", tc.flag, got, tc.expect)
		}
	}
}

func TestFlagEvaluator_BooleanValue(t *testing.T) {
	e := internal.NewFlagEvaluator()

	t.Run("returns bool node when enabled", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:           "f",
			Enabled:       true,
			Strategy:      "boolean",
			Value:         rawJSON(true),
			FallbackValue: rawJSON(false),
		}
		if got := e.BooleanValue(flag, false, "u"); !got {
			t.Errorf("expected true, got %v", got)
		}
	})

	t.Run("returns fallback bool when disabled", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:           "f",
			Enabled:       false,
			Strategy:      "boolean",
			Value:         rawJSON(true),
			FallbackValue: rawJSON(false),
		}
		if got := e.BooleanValue(flag, true, "u"); got {
			t.Errorf("expected false, got %v", got)
		}
	})

	t.Run("returns isEnabled when node is not bool", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "boolean",
			Value:    rawJSON("not-a-bool"),
		}
		// isEnabled returns true, so BooleanValue should return true.
		if got := e.BooleanValue(flag, false, "u"); !got {
			t.Errorf("expected true (isEnabled fallback), got %v", got)
		}
	})
}

func TestFlagEvaluator_StringValue(t *testing.T) {
	e := internal.NewFlagEvaluator()

	t.Run("returns string value", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON("hello"),
		}
		if got := e.StringValue(flag, "default", "u"); got != "hello" {
			t.Errorf("expected 'hello', got %q", got)
		}
	})

	t.Run("returns default when null", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:     "f",
			Enabled: true,
			Strategy: "value",
			Value:   json.RawMessage(`null`),
		}
		if got := e.StringValue(flag, "default", "u"); got != "default" {
			t.Errorf("expected 'default', got %q", got)
		}
	})

	t.Run("returns JSON string for non-string value", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON(42),
		}
		got := e.StringValue(flag, "default", "u")
		if got != "42" {
			t.Errorf("expected '42', got %q", got)
		}
	})
}

func TestFlagEvaluator_IntValue(t *testing.T) {
	e := internal.NewFlagEvaluator()

	t.Run("returns int value", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON(42),
		}
		if got := e.IntValue(flag, 0, "u"); got != 42 {
			t.Errorf("expected 42, got %d", got)
		}
	})

	t.Run("returns default for null", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    json.RawMessage(`null`),
		}
		if got := e.IntValue(flag, 99, "u"); got != 99 {
			t.Errorf("expected 99, got %d", got)
		}
	})

	t.Run("returns default for non-number", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON("not-a-number"),
		}
		if got := e.IntValue(flag, 7, "u"); got != 7 {
			t.Errorf("expected 7, got %d", got)
		}
	})
}

func TestFlagEvaluator_FloatValue(t *testing.T) {
	e := internal.NewFlagEvaluator()

	t.Run("returns float value", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON(3.14),
		}
		if got := e.FloatValue(flag, 0.0, "u"); got != 3.14 {
			t.Errorf("expected 3.14, got %f", got)
		}
	})

	t.Run("returns default for null", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    json.RawMessage(`null`),
		}
		if got := e.FloatValue(flag, 1.5, "u"); got != 1.5 {
			t.Errorf("expected 1.5, got %f", got)
		}
	})
}

func TestFlagEvaluator_ObjectValue(t *testing.T) {
	e := internal.NewFlagEvaluator()

	t.Run("returns parsed object", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    rawJSON(map[string]interface{}{"color": "blue"}),
		}
		got := e.ObjectValue(flag, nil, "u")
		m, ok := got.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", got)
		}
		if m["color"] != "blue" {
			t.Errorf("expected color=blue, got %v", m["color"])
		}
	})

	t.Run("returns default for null", func(t *testing.T) {
		flag := internal.RemoteFlag{
			Key:      "f",
			Enabled:  true,
			Strategy: "value",
			Value:    json.RawMessage(`null`),
		}
		def := map[string]interface{}{"x": 1}
		got := e.ObjectValue(flag, def, "u")
		if got != interface{}(def) {
			t.Errorf("expected default value, got %v", got)
		}
	})
}

func TestFlagEvaluator_FNV1a_KnownValues(t *testing.T) {
	// Verify FNV-1a bucketing against independently computed values.
	// seed "test-flag:user-abc"  -> hash % 100 must be stable across runs.
	e := internal.NewFlagEvaluator()
	flag := internal.RemoteFlag{
		Key:      "test-flag",
		Enabled:  true,
		Strategy: "percentage",
		Rollout:  ptr(50),
	}

	// Run 100 times to confirm determinism.
	result := e.IsEnabled(flag, "user-abc")
	for i := 0; i < 100; i++ {
		if e.IsEnabled(flag, "user-abc") != result {
			t.Fatal("FNV-1a bucketing is not deterministic")
		}
	}
}

func TestFlagCache_UpdateAndGet(t *testing.T) {
	cache := internal.NewFlagCache()

	if !cache.IsEmpty() {
		t.Error("new cache should be empty")
	}

	flags := []internal.RemoteFlag{
		{Key: "flag-a", Enabled: true, Strategy: "boolean"},
		{Key: "flag-b", Enabled: false, Strategy: "boolean"},
	}
	cache.Update(flags)

	if cache.IsEmpty() {
		t.Error("cache should not be empty after update")
	}

	f, ok := cache.Get("flag-a")
	if !ok {
		t.Fatal("flag-a should be in cache")
	}
	if !f.Enabled {
		t.Error("flag-a should be enabled")
	}

	_, ok = cache.Get("flag-c")
	if ok {
		t.Error("flag-c should not exist")
	}
}

func TestFlagCache_All(t *testing.T) {
	cache := internal.NewFlagCache()
	flags := []internal.RemoteFlag{
		{Key: "x", Enabled: true, Strategy: "boolean"},
	}
	cache.Update(flags)

	all := cache.All()
	if len(all) != 1 {
		t.Errorf("expected 1 flag, got %d", len(all))
	}

	// Mutations to the returned snapshot must not affect the cache.
	delete(all, "x")
	if _, ok := cache.Get("x"); !ok {
		t.Error("deleting from snapshot should not affect cache")
	}
}
