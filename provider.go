package trustygorilla

import (
	"context"
	"time"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/trustygorilla/go-sdk/internal"
)

// TrustyGorillaProvider implements openfeature.FeatureProvider and
// openfeature.StateHandler backed by the TrustyGorilla feature-flag service.
type TrustyGorillaProvider struct {
	config    *TrustyGorillaConfig
	cache     *internal.FlagCache
	client    *internal.ApiClient
	poller    *internal.FlagPoller
	evaluator *internal.FlagEvaluator
	state     openfeature.State
}

// NewTrustyGorillaProvider creates a provider from the given configuration.
// Call Init (or register with the OpenFeature SDK, which calls it) before
// evaluating flags.
func NewTrustyGorillaProvider(cfg *TrustyGorillaConfig) *TrustyGorillaProvider {
	cache := internal.NewFlagCache()
	client := internal.NewApiClient(cfg.APIURL, cfg.APIKey)
	interval := time.Duration(cfg.PollIntervalSeconds) * time.Second
	poller := internal.NewFlagPoller(client, cache, interval)

	return &TrustyGorillaProvider{
		config:    cfg,
		cache:     cache,
		client:    client,
		poller:    poller,
		evaluator: internal.NewFlagEvaluator(),
		state:     openfeature.NotReadyState,
	}
}

// Metadata returns the provider's identifying metadata.
func (p *TrustyGorillaProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: "TrustyGorilla"}
}

// Hooks returns no lifecycle hooks.
func (p *TrustyGorillaProvider) Hooks() []openfeature.Hook {
	return nil
}

// Init starts the background flag poller. It is called automatically by the
// OpenFeature SDK when the provider is registered.
func (p *TrustyGorillaProvider) Init(_ openfeature.EvaluationContext) error {
	p.poller.Start()
	p.state = openfeature.ReadyState
	return nil
}

// Shutdown stops the background poller and releases resources.
func (p *TrustyGorillaProvider) Shutdown() {
	p.poller.Stop()
	p.state = openfeature.NotReadyState
}

// Status returns the current provider state for the StateHandler interface.
func (p *TrustyGorillaProvider) Status() openfeature.State {
	return p.state
}

// BooleanEvaluation resolves a boolean flag value.
func (p *TrustyGorillaProvider) BooleanEvaluation(
	_ context.Context,
	flagKey string,
	defaultValue bool,
	evalCtx openfeature.FlattenedContext,
) openfeature.BoolResolutionDetail {
	flag, resErr := p.resolve(flagKey)
	if resErr != nil {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: *resErr,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	targetingKey := extractTargetingKey(evalCtx)
	return openfeature.BoolResolutionDetail{
		Value: p.evaluator.BooleanValue(flag, defaultValue, targetingKey),
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.Reason(p.evaluator.ReasonFor(flag)),
		},
	}
}

// StringEvaluation resolves a string flag value.
func (p *TrustyGorillaProvider) StringEvaluation(
	_ context.Context,
	flagKey string,
	defaultValue string,
	evalCtx openfeature.FlattenedContext,
) openfeature.StringResolutionDetail {
	flag, resErr := p.resolve(flagKey)
	if resErr != nil {
		return openfeature.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: *resErr,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	targetingKey := extractTargetingKey(evalCtx)
	return openfeature.StringResolutionDetail{
		Value: p.evaluator.StringValue(flag, defaultValue, targetingKey),
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.Reason(p.evaluator.ReasonFor(flag)),
		},
	}
}

// IntEvaluation resolves an integer flag value.
func (p *TrustyGorillaProvider) IntEvaluation(
	_ context.Context,
	flagKey string,
	defaultValue int64,
	evalCtx openfeature.FlattenedContext,
) openfeature.IntResolutionDetail {
	flag, resErr := p.resolve(flagKey)
	if resErr != nil {
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: *resErr,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	targetingKey := extractTargetingKey(evalCtx)
	return openfeature.IntResolutionDetail{
		Value: p.evaluator.IntValue(flag, defaultValue, targetingKey),
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.Reason(p.evaluator.ReasonFor(flag)),
		},
	}
}

// FloatEvaluation resolves a float64 flag value.
func (p *TrustyGorillaProvider) FloatEvaluation(
	_ context.Context,
	flagKey string,
	defaultValue float64,
	evalCtx openfeature.FlattenedContext,
) openfeature.FloatResolutionDetail {
	flag, resErr := p.resolve(flagKey)
	if resErr != nil {
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: *resErr,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	targetingKey := extractTargetingKey(evalCtx)
	return openfeature.FloatResolutionDetail{
		Value: p.evaluator.FloatValue(flag, defaultValue, targetingKey),
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.Reason(p.evaluator.ReasonFor(flag)),
		},
	}
}

// ObjectEvaluation resolves an object flag value.
func (p *TrustyGorillaProvider) ObjectEvaluation(
	_ context.Context,
	flagKey string,
	defaultValue interface{},
	evalCtx openfeature.FlattenedContext,
) openfeature.InterfaceResolutionDetail {
	flag, resErr := p.resolve(flagKey)
	if resErr != nil {
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: *resErr,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	targetingKey := extractTargetingKey(evalCtx)
	return openfeature.InterfaceResolutionDetail{
		Value: p.evaluator.ObjectValue(flag, defaultValue, targetingKey),
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason: openfeature.Reason(p.evaluator.ReasonFor(flag)),
		},
	}
}

// GetAllFlags returns a snapshot of all flags currently held in the cache.
// This is a TrustyGorilla extension and is not part of the OpenFeature spec.
func (p *TrustyGorillaProvider) GetAllFlags() map[string]internal.RemoteFlag {
	return p.cache.All()
}

// resolve looks up a flag by key. Returns a non-nil ResolutionError when the
// flag is absent from the cache.
func (p *TrustyGorillaProvider) resolve(flagKey string) (internal.RemoteFlag, *openfeature.ResolutionError) {
	flag, ok := p.cache.Get(flagKey)
	if !ok {
		resErr := openfeature.NewFlagNotFoundResolutionError("flag not found: " + flagKey)
		return internal.RemoteFlag{}, &resErr
	}
	return flag, nil
}

// extractTargetingKey retrieves the targeting key from a FlattenedContext.
func extractTargetingKey(evalCtx openfeature.FlattenedContext) string {
	if v, ok := evalCtx[openfeature.TargetingKey]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
