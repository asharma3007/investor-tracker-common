package common

import (
	"encoding/json"
	. "github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	CURRENCY_GBP = "GBP"
	CURRENCY_USD = "USD"
	CURRENCY_EUR = "EUR"
)

var currencyConverter = make(map[string]Decimal)

type Money struct {
	Currency string
	Value DecimalExt // always in units e.g pound, dollar not pence, cent
}

func (from Money) toCurrency(toCurrency string) Money {
	key := getConversionKey(from.Currency, toCurrency)

	if _, contains := currencyConverter[key]; !contains {
		conversion := GetConversionValue(from.Currency, toCurrency)
		currencyConverter[key] = conversion

		//and put reverse in too
		reverseKey := getConversionKey(toCurrency, from.Currency)
		currencyConverter[reverseKey] = NewFromInt(1).Div(conversion)
	}

	conversion := currencyConverter[key]

	targetValue := from.Value.Mul(conversion)

	return Money{
		Currency: toCurrency,
		Value:    DecimalExt{targetValue},
	}
}

func getConversionKey(from string, to string) string {
	return from + ":" + to;
}

func GetConversionValue(from string, to string) Decimal {

	todayStr := time.Now().Format("2006-01-02")

	response, err := http.Get("https://api.exchangeratesapi.io/history?" +
		"start_at=" + todayStr +
		"&end_at=" + todayStr +
		"&base=" + from +
		"&symbols=" + to)
	CheckError(err)

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	CheckError(err)

	responseString := string(responseData)
	Log(responseString)

	var retval map[string]interface{}
	err = json.Unmarshal(responseData, &retval)
	CheckError(err)

	conversion := retval["rates"].(map[string]interface{})[todayStr].(map[string]interface{})[to].(float64)
	return NewFromFloat(conversion)
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