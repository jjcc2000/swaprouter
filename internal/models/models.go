package models

import "time"

type QuoteRequest struct {
	FromToken string  `json:"fromToken"`
	ToToken   string  `json:"toToken"`
	AmountIn  float64 `json:"amount"`
	Chain     string  `json:"chain"`
	Wallet    string  `json:"wallet"`
}

type Quote struct {
	Protocol    string    `json:"protocol"`
	FromToken   string    `json:"fromToken"`
	ToToken     string    `json:"toToken"`
	AmountIn    float64   `json:"amountIn"`
	AmountOut   float64   `json:"amountOut"`
	GasEstimate float64   `json:"gasEstimate"`
	NetValue    float64   `json:"netValue"`
	Chain       string    `json:"chain"`
	ExpiresAt   time.Time `json:"expiresAt"`
	QuoteID     string    `json:"quoteId"`
}

type SwapRequest struct {
	QuoteID     string `json:"quoteId"`
	SlippageBps int    `json:"slippageBps"`
	Wallet      string `json:"wallet"`
}

type SwapResult struct {
	TxHash    string    `json:"txHash"`
	Chain     string    `json:"chain"`
	Protocol  string    `json:"protocol"`
	AmountIn  float64   `json:"amountIn"`
	AmountOut float64   `json:"amountOut"`
	Wallet    string    `json:"wallet"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type Trade struct {
	ID        string    `json:"id"`
	TxHash    string    `json:"txHash"`
	Wallet    string    `json:"wallet"`
	Chain     string    `json:"chain"`
	Protocol  string    `json:"protocol"`
	FromToken string    `json:"fromToken"`
	ToToken   string    `json:"toToken"`
	AmountIn  float64   `json:"amountIn"`
	AmountOut float64   `json:"amountOut"`
	GasPaid   float64   `json:"gasPaid"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Token struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Chain    string `json:"chain"`
	Decimals int    `json:"decimals"`
}

type Chain struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	ChainID int    `json:"chainId,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Message }
