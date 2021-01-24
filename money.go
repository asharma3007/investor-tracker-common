package common

import (
	. "github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CURRENCY_GBP = "GBP"
	CURRENCY_USD = "USD"
	CURRENCY_EUR = "EUR"
)

type Money struct {
	Currency string
	Value DecimalExt // always in units e.g pound, dollar not pence, cent
}

func (w Money) MarshalBSON() ([]byte, error) {
	intermediate := make(map[string]string)
	intermediate["value"] = w.Value.String()
	intermediate["currency"] = w.Currency

	return bson.Marshal(intermediate)
}

// SetBSON implements bson.Setter.
func (w *Money) UnmarshalBSON(raw []byte) error {

	var decoded map[string]string
	bsonErr := bson.Unmarshal(raw, &decoded)
	if bsonErr != nil {
		return bsonErr
	}

	pounds, err := NewFromString(decoded["value"])
	CheckError(err)
	currency := decoded["currency"]

	w.Currency = currency
	w.Value = DecimalExt{pounds}
	return nil
}

func FromCents(centsStr string) Money {
	cents, err := NewFromString(centsStr)
	CheckError(err)
	dollars := cents.Div(NewFromInt(100))
	return Money{
		Currency: CURRENCY_USD,
		Value:    DecimalExt{dollars},
	}
}

func FromPence(penceStr string) Money {
	pence, err := NewFromString(penceStr)
	CheckError(err)
	pounds := pence.Div(NewFromInt(100))
	return Money{
		Currency: CURRENCY_GBP,
		Value:    DecimalExt{pounds},
	}
}

func (m Money) ToSubunits() Decimal {
	return m.Value.Mul(NewFromInt(100))
}

func ConvertUsdToGbp(usd Decimal) Money {
	return Money{ "GBP", DecimalExt{NewFromInt(1)}}
}
