package investor_tracker_common

import (
	"encoding/json"
	. "github.com/shopspring/decimal"
)

const (
	EnvDatabaseUrl      = "DATABASE_URL"
	EnvDatabasePassword = "DATABASE_PASSWORD"
	port                = "3306"
	user                = "root"
	dbname              = "tracker"

	EnvTokenMarketStack = "TOKEN_MARKETSTACK"
)

type Stock struct {
	Description string
	Symbol string
	HlName    string
	HlUrlOverride string
	Url       string
	PriceBuy  Decimal
	PriceSell Decimal
}

type MessageSendEmail struct {
	Html string
	PlainText string
	SenderName string
	Subject string
}

type Holding struct {
	StockId      int
	Lots         []Lot
	Transactions []Transaction
}

type Lot struct {
	StockId     int
	PriceBought Decimal
	Units       Decimal
	Transaction Transaction
}

type Transaction struct {
	TransactionId int
	StockId int
	DtTrade string
	UnitPrice Decimal
	Units Decimal
	ValueQuoted Decimal
}

func (stock Stock) ToString() string {
	desc, _ := json.Marshal(stock)
	return string(desc)
}

func (stock *Stock) IsSourceHl() bool {
	return len(stock.HlName) > 0
}

type Alert struct {
	Instruction MonitorInstruction
	Message string
}

func TestAnkit(test string) string {
	return test
}

type MonitorInstruction struct {
	StockId            int
	PriceTypeToMonitor int
	MarkerPrice        Decimal
	Message            string
	Holding            Holding
}
