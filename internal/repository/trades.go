package repository

import (
	"database/sql"
	"fmt"

	"github.com/jjcc2000/swaprouter/internal/models"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) UpdateStatus(tradeID, txHash, wallet, status string) error {
	result, err := r.db.Exec(
		`UPDATE trades SET status=$1, tx_hash=$2 WHERE id=$3 AND wallet=$4`,
		status, txHash, tradeID, wallet,
	)

	if err != nil {
		return err
	}

	rows, _ :=result.RowsAffected()
	if rows ==0{
		return fmt.Errorf("trade not found or wallet mismatch")

	}
	return nil
}

func (r *TradeRepository) Save(t models.Trade) (string, error) {

	var id string
	err := r.db.QueryRow(`
        INSERT INTO trades (tx_hash, wallet, chain, protocol, from_token, to_token, amount_in, amount_out, gas_paid, status)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING id`,
		t.TxHash, t.Wallet, t.Chain, t.Protocol,
		t.FromToken, t.ToToken, t.AmountIn, t.AmountOut, t.GasPaid, t.Status,
	).Scan(&id)
	return id, err
}

func (r *TradeRepository) GetByWallet(wallet string) ([]models.Trade, error) {
	rows, err := r.db.Query(`SELECT * FROM trades WHERE wallet=$1 ORDER BY created_at DESC`, wallet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {

		var t models.Trade
		if err := rows.Scan(&t.ID, &t.TxHash, &t.Wallet, &t.Chain, &t.Protocol,
			&t.FromToken, &t.ToToken, &t.AmountIn, &t.AmountOut, &t.GasPaid, &t.Status, &t.CreatedAt); err != nil {
			continue
		}

		trades = append(trades, t)
	}

	return trades, nil
}

func (r *TradeRepository) GetPendingTrades() ([]models.Trade, error) {
	rows, err := r.db.Query(
		`SELECT id, tx_hash, wallet, chain FROM trades WHERE status = 'pending' AND tx_hash != ''`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		rows.Scan(&t.ID, &t.TxHash, &t.Wallet, &t.Chain)
		trades = append(trades, t)
	}
	return trades, nil
}
