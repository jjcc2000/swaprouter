package aggregator


import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jjcc2000/swaprouter/internal/models"
)

type IAdapter interface {
	GetQuote(ctx context.Context, req models.QuoteRequest) (*models.Quote, error)
	Name() string
}

type QuoteEngine struct {
	adapters []IAdapter
	timeout  time.Duration
}

func NewQuoteEngine(adapters []IAdapter, timeoutMs int) *QuoteEngine {
	return &QuoteEngine{
		adapters: adapters,
		timeout:  time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (qe *QuoteEngine) Timeout() time.Duration {
	return qe.timeout
}

type quoteResult struct {
	quote *models.Quote
	err   error
}

func (qe *QuoteEngine) GetBestQuote(ctx context.Context, req models.QuoteRequest) (*models.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	results := make(chan quoteResult, len(qe.adapters))
	var wg sync.WaitGroup

	for _, adapter := range qe.adapters {
		wg.Add(1)
		go func(a IAdapter) {
			defer wg.Done()
			quote, err := a.GetQuote(ctx, req)
			results <- quoteResult{quote: quote, err: err}
		}(adapter)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var quotes []*models.Quote
	for r := range results {
		if r.err != nil {
			fmt.Printf("[quote-engine] adapter error: %v\n", r.err)
			continue
		}
		quotes = append(quotes, r.quote)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("all adapters failed for %s → %s on %s",
			req.FromToken, req.ToToken, req.Chain)
	}

	return bestQuote(quotes), nil
}

func (qe *QuoteEngine) ExecuteSwap(ctx context.Context, req models.SwapRequest) (*models.SwapResult, error) {
	// Execution engine comes in the next layer.
	// For now returns a pending stub so the API compiles and runs.
	return &models.SwapResult{
		TxHash:    "0xpending",
		Chain:     "ethereum",
		Wallet:    req.Wallet,
		Timestamp: time.Now(),
		Status:    "pending",
	}, nil
}

func bestQuote(quotes []*models.Quote) *models.Quote {
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].NetValue > quotes[j].NetValue
	})
	return quotes[0]
}