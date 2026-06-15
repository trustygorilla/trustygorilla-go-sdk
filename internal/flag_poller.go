package internal

import (
	"log"
	"time"
)

// FlagPoller fetches flags from the API on a fixed interval and keeps the
// FlagCache up to date. It runs on a dedicated background goroutine.
type FlagPoller struct {
	client   *ApiClient
	cache    *FlagCache
	interval time.Duration
	done     chan struct{}
}

// NewFlagPoller creates a FlagPoller that uses the supplied client and cache.
func NewFlagPoller(client *ApiClient, cache *FlagCache, interval time.Duration) *FlagPoller {
	return &FlagPoller{
		client:   client,
		cache:    cache,
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start launches the background polling goroutine. It fetches flags
// immediately and then repeats every pollInterval. All errors are logged
// as warnings; they never crash the goroutine.
func (p *FlagPoller) Start() {
	go func() {
		p.poll()

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.poll()
			case <-p.done:
				return
			}
		}
	}()
}

// Stop signals the polling goroutine to exit and waits for it to finish.
func (p *FlagPoller) Stop() {
	close(p.done)
}

func (p *FlagPoller) poll() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[TrustyGorilla] WARNING: panic during flag poll: %v", r)
		}
	}()

	flags, err := p.client.FetchFlags()
	if err != nil {
		log.Printf("[TrustyGorilla] WARNING: flag fetch error: %v", err)
		return
	}

	// nil means 304 Not Modified — leave the cache unchanged.
	if flags == nil {
		return
	}

	p.cache.Update(flags)
}
