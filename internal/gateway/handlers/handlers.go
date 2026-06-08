package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jjcc2000/swaprouter/internal/aggregator"
	"github.com/jjcc2000/swaprouter/internal/gateway/middleware"
	"github.com/jjcc2000/swaprouter/internal/models"
	"github.com/jjcc2000/swaprouter/internal/repository"
)

type QuoteHandler struct{ engine *aggregator.QuoteEngine }

type TradesHandler struct{ repo *repository.TradeRepository }

func NewQuoteHandler(e *aggregator.QuoteEngine) *QuoteHandler { return &QuoteHandler{engine: e} }

func (h *QuoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromToken := q.Get("fromToken")
	toToken := q.Get("toToken")
	chain := q.Get("chain")
	amountStr := q.Get("amount")

	if fromToken == "" || toToken == "" || chain == "" || amountStr == "" {
		writeError(w, 400, "MISSING_PARAMS", "fromToken, toToken, amount, chain are required")
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		writeError(w, 400, "INVALID_AMOUNT", "amount must be a positive number")
		return
	}

	req := models.QuoteRequest{
		FromToken: fromToken,
		ToToken:   toToken,
		AmountIn:  amount,
		Chain:     chain,
		Wallet:    middleware.WalletFromContext(r.Context()),
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.engine.Timeout())
	defer cancel()

	quote, err := h.engine.GetBestQuote(ctx, req)
	if err != nil {
		writeError(w, 502, "QUOTE_FAILED", err.Error())
		return
	}

	writeJSON(w, 200, quote)
}

type SwapHandler struct {
	engine *aggregator.QuoteEngine
	repo   *repository.TradeRepository
}

func NewSwapHandler(e *aggregator.QuoteEngine, repo *repository.TradeRepository) *SwapHandler {
	return &SwapHandler{engine: e, repo: repo}
}

func (h *SwapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req models.SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "INVALID_BODY", "invalid request body")
		return
	}
	if req.QuoteID == "" {
		writeError(w, 400, "MISSING_QUOTE_ID", "quoteId is required")
		return
	}
	req.Wallet = middleware.WalletFromContext(r.Context())

	result, err := h.engine.ExecuteSwap(r.Context(), req)
	if err != nil {
		writeError(w, 502, "SWAP_FAILED", err.Error())
		return
	}

	// save trade to database
	h.repo.Save(models.Trade{
		TxHash:    result.TxHash,
		Wallet:    result.Wallet,
		Chain:     result.Chain,
		Protocol:  result.Protocol,
		FromToken: result.FromToken,
		ToToken:   result.ToToken,
		AmountIn:  result.AmountIn,
		AmountOut: result.AmountOut,
		Status:    result.Status,
	})
	writeJSON(w, 200, result)
}
func NewTradesHandler(repo *repository.TradeRepository) *TradesHandler {
	return &TradesHandler{repo: repo}
}

func (h *TradesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wallet := middleware.WalletFromContext(r.Context())
	trades, err := h.repo.GetByWallet(wallet)
	if err != nil {
		writeError(w, 500, "DB_ERROR", "failed to fetch trades")
		return
	}

	writeJSON(w, 200, map[string]interface{}{"wallet": wallet, "trades": trades})
}

func TokensHandler(w http.ResponseWriter, r *http.Request) {
	chain := r.URL.Query().Get("chain")
	writeJSON(w, 200, map[string]interface{}{"tokens": supportedTokens(chain)})
}

func ChainsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{
		"chains": []models.Chain{
			{ID: "ethereum", Name: "Ethereum", ChainID: 1},
			{ID: "polygon", Name: "Polygon", ChainID: 137},
			{ID: "base", Name: "Base", ChainID: 8453},
			{ID: "solana", Name: "Solana"},
		},
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func supportedTokens(chain string) []models.Token {
	all := []models.Token{
		{Symbol: "USDC", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", Chain: "ethereum", Decimals: 6},
		{Symbol: "ETH", Address: "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE", Chain: "ethereum", Decimals: 18},
		{Symbol: "SOL", Address: "So11111111111111111111111111111111111111112", Chain: "solana", Decimals: 9},
		{Symbol: "USDC", Address: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", Chain: "solana", Decimals: 6},
	}
	if chain == "" {
		return all
	}
	var out []models.Token
	for _, t := range all {
		if t.Chain == chain {
			out = append(out, t)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, models.APIError{Code: code, Message: msg})
}
