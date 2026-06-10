package oneinch

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/jjcc2000/swaprouter/internal/models"
)

type Adapter struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func New(apiKey, baseURL string) *Adapter {
	return &Adapter{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (a *Adapter) Name() string { return "1inch" }
func (a *Adapter) GetSwapTx(ctx context.Context, quote *models.Quote, wallet string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (a *Adapter) GetQuote(ctx context.Context, req models.QuoteRequest) (*models.Quote, error) {
	chainId := chainID(req.Chain)
	if chainId == 0 {
		return nil, fmt.Errorf("unsuported chain: %s", req.Chain)
	}

	amountWei := toWei(req.AmountIn, 18)

	url := fmt.Sprintf("%s/%d/quote?src=%s&dst=%s&amount=%s",
		a.baseURL, chainId, req.FromToken, req.ToToken, amountWei)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("1inch API error: %d", resp.StatusCode)
	}
	var result struct {
		DstAmount string  `json:"dstAmount"`
		Gas       float64 `json:"gas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	amountOut := fromWei(result.DstAmount, 18)

	return &models.Quote{
		Protocol:    "1inch",
		FromToken:   req.FromToken,
		ToToken:     req.ToToken,
		AmountIn:    req.AmountIn,
		AmountOut:   amountOut,
		GasEstimate: result.Gas,
		NetValue:    amountOut - (result.Gas * 0.000000001),
		Chain:       req.Chain,
		ExpiresAt:   time.Now().Add(30 * time.Second),
		QuoteID:     fmt.Sprintf("1inch-%d", time.Now().UnixNano()),
	}, nil
}

func chainID(chain string) int {
	switch chain {
	case "ethereum":
		return 1
	case "polygon":
		return 137
	case "base":
		return 8453

	default:
		return 0
	}
}

func toWei(amount float64, decimals int) string {

	multiplier := new(big.Float).SetFloat64(amount)
	exp := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	wei, _ := new(big.Float).Mul(multiplier, exp).Int(nil)
	return wei.String()
}

func fromWei(weiString string, decimals int) float64 {
	wei := new(big.Int)
	wei.SetString(weiString, 10)
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	result, _ := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		new(big.Float).SetInt(exp),
	).Float64()
	return result

}
