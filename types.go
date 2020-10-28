package investor_tracker_common

import (
	"encoding/json"
	. "github.com/shopspring/decimal"
	"strings"
	"time"
)

const (
	EnvDatabaseUrl      = "DATABASE_URL"
	EnvDatabasePassword = "DATABASE_PASSWORD"
	port                = "3306"
	user                = "root"
	dbname              = "tracker"

	EnvTokenMarketStack = "TOKEN_MARKETSTACK"

	PriceTypeSell = 0
	PriceTypeBuy = 1
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

func (stock *Stock) GetDisplayName() string {
	if len(stock.HlName) > 0 {
		return stock.HlName
	} else {
		return stock.Description
	}
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

func (instruction MonitorInstruction) getDesc() string {
	if instruction.isSell() {
		return "Sell"
	}
	return "Buy"
}

func (instruction MonitorInstruction) isBuy() bool {
	return instruction.PriceTypeToMonitor == PriceTypeBuy
}

func (instruction MonitorInstruction) isSell() bool {
	return instruction.PriceTypeToMonitor == PriceTypeSell
}


func (holding Holding) getUnitsTotal() Decimal {
	retVal := NewFromInt(0)
	for _, lot := range holding.Lots {
		retVal = retVal.Add(lot.Units)
	}
	return retVal
}

func (lot Lot) getValueTotal() Decimal {
	return lot.PriceBought.Mul(lot.Units)
}

func (holding Holding) getPriceAverage() Decimal {
	totalValue := NewFromInt(0)
	for _, lot := range holding.Lots {
		lotValue := lot.getValueTotal()
		totalValue = totalValue.Add(lotValue)
	}

	totalUnits := holding.getUnitsTotal()
	return totalValue.Div(totalUnits)
}

type watchDetail struct {
	Stock Stock
	Watch watch
	History priceHistory
	ChangePercent Decimal
}

type watch struct {
	WatchId        int
	StockId        int
	DtReference    string
	AddedPriceBuy  Decimal
	AddedPriceSell Decimal
	AlertThreshold Decimal
	Notes		   string
}

type priceHistory struct {
	Eods []eodMarketStack
}

type eodMarketStack struct {
	Date timeMarketStack `json:"date"`
	PriceClose Decimal `json:"close"`
}

type timeMarketStack struct {
	time.Time
}

const TimeFormatMarketStack = "2006-01-02T03:04:05+0000"

func (t *timeMarketStack) UnmarshalJSON(buf []byte) error {
	tt, err := time.Parse(TimeFormatMarketStack, strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

func (transaction Transaction) IsBuy() bool {
	return transaction.ValueQuoted.IsNegative()
}

func (transaction Transaction) IsSell() bool {
	return transaction.ValueQuoted.IsPositive()
}
