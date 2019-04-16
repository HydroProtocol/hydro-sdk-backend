package launcher

import (
	"database/sql"
	"github.com/shopspring/decimal"
	"time"
)

type LaunchLog struct {
	ID       int64          `db:"id" auto:"true" primaryKey:"true" autoIncrement:"true"`
	ItemType string         `db:"item_type"`
	ItemID   int64          `db:"item_id"`
	Status   string         `db:"status"`
	Hash     sql.NullString `db:"transaction_hash"`

	BlockNumber sql.NullInt64 `db:"block_number"`

	From     string              `db:"t_from"`
	To       string              `db:"t_to"`
	Value    decimal.Decimal     `db:"value"`
	GasLimit int64               `db:"gas_limit"`
	GasUsed  sql.NullInt64       `db:"gas_used"`
	GasPrice decimal.NullDecimal `db:"gas_price"`
	Nonce    sql.NullInt64       `db:"nonce"`
	Data     string              `db:"data"`

	ExecutedAt time.Time `db:"executed_at"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
