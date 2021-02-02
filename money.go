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

func cacheConversion(from string, to string) {
	key := getConversionKey(from, to)
	conversion := GetConversionValue(from, to)
	currencyConverter[key] = conversion

	//and put reverse in too
	reverseKey := getConversionKey(to, from)
	currencyConverter[reverseKey] = NewFromInt(1).Div(conversion)
}

func (from Money) toCurrency(toCurrency string) Money {
	key := getConversionKey(from.Currency, toCurrency)

	if _, contains := currencyConverter[key]; !contains {
		cacheConversion(from.Currency, toCurrency)
	}

	conversion := currencyConverter[key]

	targetValue := from.Value.Mul(conversion)

	return Money{
		Currency: toCurrency,
		Value:    DecimalExt{targetValue},
	}
}

func (this Money) Add(other Money) Money {
	checkCurrency(this, other)

	newValue := this.Value.Add(other.Value.Decimal)
	return Money{
		Currency: this.Currency,
		Value:    DecimalExt{newValue},
	}
}

func checkCurrency(this Money, other Money) {
	if this.Currency != other.Currency { panic("Incompatible currencies" + this.Currency + " " + other.Currency)}
}

func getConversionKey(from string, to string) string {
	return from + ":" + to;
}

func getLastWorkingDay() time.Time {
	//doesn't work on weekends

	day := time.Now()

	for {
		dayOfWeek := day.Weekday()
		if dayOfWeek != time.Saturday && dayOfWeek != time.Sunday {
			return day
		}
		day = day.AddDate(0, 0, -1)
	}
}

func GetConversionValue(from string, to string) Decimal {

	weekdayStr := getLastWorkingDay().Format("2006-01-02")

	url := "https://api.exchangeratesapi.io/history?" +
		"start_at=" + weekdayStr +
		"&end_at=" + weekdayStr +
		"&base=" + from +
		"&symbols=" + to

	Log(url)

	response, err := http.Get(url)
	CheckError(err)

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	CheckError(err)

	responseString := string(responseData)
	Log(responseString)

	var retval map[string]interface{}
	err = json.Unmarshal(responseData, &retval)
	CheckError(err)

	conversion := retval["rates"].(map[string]interface{})[weekdayStr].(map[string]interface{})[to].(float64)
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

func (this Money) Sub(other Money) Money {
	checkCurrency(this, other)

	result := this.Value.Sub(other.Value.Decimal)

	return Money{
		Currency: this.Currency,
		Value:    DecimalExt{result},
	}
}

func (this Money) Div(other Money) Decimal {
	checkCurrency(this, other)

	return this.Value.Div(other.Value.Decimal)
}

func (m *Money) Mul(factor Decimal) Money {
	result := m.Value.Mul(factor)

	return Money{
		Currency: m.Currency,
		Value:    DecimalExt{result},
	}
}

func (m *Money) ToUnits() Money {
	result := m.Value.Decimal.Div(NewFromInt(100))
	return Money{
		Currency: m.Currency,
		Value:    DecimalExt{result},
	}
}

func (m *Money) String() string {
	return m.GetDesc()
}

func (m *Money) DebugString() string {
	return m.GetDesc()
}

func FromPounds(poundsStr string) Money {
	value, err := NewFromString(poundsStr)
	CheckError(err)
	return Money{
		Currency: CURRENCY_GBP,
		Value:    DecimalExt{value},
	}
}
