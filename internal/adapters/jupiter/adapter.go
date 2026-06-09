package jupiter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jjcc2000/swaprouter/internal/models"
)

type Adapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func New(baseURL string, apiKey string) *Adapter {
	return &Adapter{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
		apiKey:  apiKey,
	}
}

func (a *Adapter) Name() string { return "jupiter" }

func (a *Adapter) GetQuote(ctx context.Context, req models.QuoteRequest) (*models.Quote, error) {
	if req.Chain != "solana" {
		return nil, fmt.Errorf("jupiter only supports solana")
	}

	amountLamports := int64(req.AmountIn * 1e9)

	url := fmt.Sprintf("%s/quote?inputMint=%s&outputMint=%s&amount=%d&slippageBps=50",
		a.baseURL, req.FromToken, req.ToToken, amountLamports)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("jupiter API error: %d", resp.StatusCode)
	}

	var result struct {
		OutAmount            string `json:"outAmount"`
		OtherAmountThreshold string `json:"otherAmountThreshold"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	amountOut := float64(0)
	fmt.Sscanf(result.OutAmount, "%f", &amountOut)
	amountOut = amountOut / 1e6 // USDC has 6 decimals

	return &models.Quote{
		Protocol:  "jupiter",
		FromToken: req.FromToken,
		ToToken:   req.ToToken,
		AmountIn:  req.AmountIn,
		AmountOut: amountOut,
		NetValue:  amountOut,
		Chain:     req.Chain,
		ExpiresAt: time.Now().Add(30 * time.Second),
		QuoteID:   fmt.Sprintf("jupiter-%d", time.Now().UnixNano()),
	}, nil
}
