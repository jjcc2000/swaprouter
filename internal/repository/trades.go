package repository

import (
	"database/sql"

	"github.com/jjcc2000/swaprouter/internal/models"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) Save(t models.Trade) error {

	_, err := r.db.Exec(`
        INSERT INTO trades (tx_hash, wallet, chain, protocol, from_token, to_token, amount_in, amount_out, gas_paid, status)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.TxHash, t.Wallet, t.Chain, t.Protocol,
		t.FromToken, t.ToToken, t.AmountIn, t.AmountOut, t.GasPaid, t.Status,
	)

	return err
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
		rows.Scan(&t.ID, &t.TxHash, &t.Wallet, &t.Chain, &t.Protocol,
			&t.FromToken, &t.ToToken, &t.AmountIn, &t.AmountOut, &t.GasPaid, &t.Status, &t.CreatedAt)

		trades = append(trades, t)
	}

	return trades, nil
}
