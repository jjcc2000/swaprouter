package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jjcc2000/swaprouter/internal/repository"
)

type Poller struct {
	repo      *repository.TradeRepository
	solanaRPC string
	interval  time.Duration
	client    *http.Client
}

func New(repo *repository.TradeRepository, solanaRPC string, intervalSec int) *Poller {
	return &Poller{
		repo:      repo,
		solanaRPC: solanaRPC,
		interval:  time.Duration(intervalSec) * time.Second,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.poll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	trades, err := p.repo.GetPendingTrades()
	if err != nil || len(trades) == 0 {
		return
	}
	for _, trade := range trades {
		if trade.Chain != "solana" || trade.TxHash == "" {
			continue
		}
		status, err := p.checkSolana(ctx, trade.TxHash)
		if err != nil || status == "" {
			continue
		}
		p.repo.UpdateStatus(trade.ID, trade.TxHash, trade.Wallet, status)
	}
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type sigStatus struct {
	Err                interface{} `json:"err"`
	ConfirmationStatus string      `json:"confirmationStatus"`
}

type rpcResponse struct {
	Result struct {
		Value []*sigStatus `json:"value"`
	} `json:"result"`
}

func (p *Poller) checkSolana(ctx context.Context, signature string) (string, error) {
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getSignatureStatuses",
		Params: []interface{}{
			[]string{signature},
			map[string]bool{"searchTransactionHistory": true},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.solanaRPC, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Result.Value) == 0 || result.Result.Value[0] == nil {
		return "", nil // not on-chain yet
	}

	val := result.Result.Value[0]
	if val.Err != nil {
		return "failed", nil
	}
	if val.ConfirmationStatus == "confirmed" || val.ConfirmationStatus == "finalized" {
		return "confirmed", nil
	}
	return "", nil
}
