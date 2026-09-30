//go:build !treesitter

package symbols

import scipingester "github.com/hollis-labs/stack-explorer/internal/symbols/scip"

func NewIngester(cfg Config) Ingester {
	return scipingester.NewIngester(cfg.Store)
}
