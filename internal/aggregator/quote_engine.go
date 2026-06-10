package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"time"

	"github.com/jjcc2000/swaprouter/internal/models"
	"github.com/redis/go-redis/v9"
)

type IAdapter interface {
	GetQuote(ctx context.Context, req models.QuoteRequest) (*models.Quote, error)
	GetSwapTx(ctx context.Context, quote *models.Quote, wallet string) (string, error)
	Name() string
}
type quoteResult struct {
	quote *models.Quote
	err   error
}
type QuoteEngine struct {
	adapters []IAdapter
	timeout  time.Duration
}

func (qe *QuoteEngine) GetCachedQuote(ctx context.Context, rdb *redis.Client, quoteID string) (*models.Quote, error) {
	fmt.Printf("[quote-engine] looking up quote: %s\n", quoteID)
	data, err := rdb.Get(ctx, "quote:"+quoteID).Bytes()
	if err != nil {
		return nil, err
	}

	var quote models.Quote
	json.Unmarshal(data, &quote)
	return &quote, nil

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

func (qe *QuoteEngine) GetBestQuote(ctx context.Context, req models.QuoteRequest, rdb *redis.Client) (*models.Quote, error) {
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

	best := bestQuote(quotes)
	data, err := json.Marshal(best)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[quote-engine] caching quote: %s\n", best.QuoteID)
	if err := rdb.Set(context.Background(), "quote:"+best.QuoteID, data, 5*time.Minute).Err(); err != nil {
		fmt.Printf("[quote-engine] redis cache error: %v\n", err)
	} else {
		fmt.Printf("[quote-engine] quote cached successfully: %s\n", best.QuoteID)
	}
	return best, nil
}

func (qe *QuoteEngine) ExecuteSwap(ctx context.Context, req models.SwapRequest, rdb *redis.Client) (*models.SwapResult, error) {

	var err error

	cachedQuoute, err := qe.GetCachedQuote(ctx, rdb, req.QuoteID)
	if err != nil {
		return nil, fmt.Errorf("Error in the reading of caching: %v", err.Error())
	}

	fmt.Printf("Cached Quoute: %v", cachedQuoute)

	for _, a := range qe.adapters {
		if a.Name() == cachedQuoute.Protocol {
			unsignedTx, err := a.GetSwapTx(ctx, cachedQuoute, req.Wallet)
			if err != nil {
				return nil, fmt.Errorf("swap tx error: %v", err)
			}
			return &models.SwapResult{
				UnsignedTx: unsignedTx,
				Protocol:   cachedQuoute.Protocol,
				Chain:      cachedQuoute.Chain,
				Wallet:     req.Wallet,
				FromToken:  cachedQuoute.FromToken, // add
				ToToken:    cachedQuoute.ToToken,   // add
				AmountIn:   cachedQuoute.AmountIn,  // add
				AmountOut:  cachedQuoute.AmountOut, // add
				Timestamp:  time.Now(),             // add
				Status:     "unsigned",
			}, nil
		}
	}
	return nil, fmt.Errorf("no adapter found for protocol: %s", cachedQuoute.Protocol)

}

func bestQuote(quotes []*models.Quote) *models.Quote {
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].NetValue > quotes[j].NetValue
	})
	return quotes[0]
}
