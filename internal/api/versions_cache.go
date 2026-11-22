package api

import (
	"context"
	"time"
)

// One shared, deadline-bounded refresh per daemon. A forced refresh may bypass
// the normal TTL, but not the cooldown or an already-running fetch.
func cachedVersionsCatalog(ctx context.Context, force bool) (*MinecraftVersionsCatalog, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		versionsMu.Lock()
		now := time.Now()
		if versionsCached != nil && ((!force && now.Before(versionsExpiresAt)) || now.Before(versionsRefreshAfter)) {
			catalog := versionsCached
			versionsMu.Unlock()
			return catalog, nil
		}
		done := versionsLoading
		if done == nil {
			done = make(chan struct{})
			versionsLoading = done
			go func() {
				// buildVersionsCatalog has its own six-second deadline. One caller
				// disconnecting must not cancel a fetch shared by other callers.
				catalog := buildVersionsCatalog(context.Background())
				versionsMu.Lock()
				versionsCached = catalog
				versionsExpiresAt = time.Now().Add(versionsCacheTTL)
				versionsRefreshAfter = time.Now().Add(time.Minute)
				versionsLoading = nil
				close(done)
				versionsMu.Unlock()
			}()
		}
		versionsMu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
	}
}
