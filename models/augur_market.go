package models

import "github.com/shopspring/decimal"

// IAugurMarketDao is an interface about how to fetch augur data from storage
type IAugurMarketDao interface {
	GetAll() []*AugurMarket
}

// augurMarketDao is default dao to fetch augur data from db.
type augurMarketDao struct{}

var AugurMarketDao = &augurMarketDao{}

type AugurMarket struct {
	ID          int64               `json:"id"          db:"id" primaryKey:"true" autoIncrement:"true"`
	Category    string              `json:"category"    db:"category"`
	Title       string              `json:"title"       db:"title"`
	Description string              `json:"description" db:"description"`
	Address     string              `json:"address"     db:"address"`
	Author      string              `json:"author"      db:"author"`
	Minimum     decimal.NullDecimal `json:"minimum"     db:"minimum"`
	Maximum     decimal.NullDecimal `json:"maximum"     db:"maximum"`
}

func (_ *augurMarketDao) GetAll() []*AugurMarket {
	markets := []*AugurMarket{}
	findAllBy(&markets, nil, nil, -1, -1)
	return markets
}
