package common

import (
	"encoding/json"
	. "github.com/shopspring/decimal"
	"io/ioutil"
	"testing"
	"time"
)

func getWatchDetailUsd() WatchDetail {
	exampleStock := Stock{
		Description: "Tesla, Inc.",
		Symbol:      "TSLA",
	}

	file, _ := ioutil.ReadFile("examples/tsla.json")

	var responseDays ResponseMarketStack
	err := json.Unmarshal([]byte(file), &responseDays)
	CheckError(err)

	return CreateWatchDetailFromMarketStackResponse(&responseDays, &exampleStock)
}

func getWatchDetailUk() WatchDetail {
	exampleStock := Stock{
		Description: "International Airlines Group",
		Symbol:      "IAG",
	}

	file, _ := ioutil.ReadFile("examples/iag.json")

	var responseDays ResponseMarketStack
	err := json.Unmarshal([]byte(file), &responseDays)
	CheckError(err)

	return CreateWatchDetailFromMarketStackResponse(&responseDays, &exampleStock)
}

func TestParseMarketStackResponseUsd(t *testing.T) {
	key := getConversionKey(CURRENCY_USD, CURRENCY_GBP)
	currencyConverter[key], _ = NewFromString("0.5")

	wd := getWatchDetailUsd();
	testParseMarketStackResponse(t, wd, 6,
		"434.0",  //434.0
		"217",
		"2020-10-09 00:00",
		"217 GBP", //434.0 USD
		"425.92",
		"2020-10-02 00:00",
		"207.545", //"415.09"
		"212.96 GBP")  //"425.92 USD"
}

func TestParseMarketStackResponseGbp(t *testing.T) {
	wd := getWatchDetailUk();
	testParseMarketStackResponse(t, wd, 232,
		"96.44",
		"0.9644",
		"2020-10-30 00:00",
		"0.964 GBP",
		"91.08",
		"2020-10-23 00:00",
		"1.090",
		"0.911 GBP")
}

func testParseMarketStackResponse(t *testing.T,
	detail WatchDetail,
	expectedDays int,
	day1expectedCloseOriginalStr string,
	day1ExpectedCloseConvertedStr string,
	day1DateStr string,
	day1ClosePriceExpected string,
	day2ClosePriceStr string,
	day6DateStr string,
	day6ExpectedCloseStr string,
	pricePreviousCloseDescExp string) {

	days := detail.History.Eods
	numberDays := len(days)
	if numberDays != expectedDays {
		t.Errorf("Unexpected number of days, expected %v actual %v", expectedDays, numberDays)
	}

	day1 := days[0]
	day1ExpectedCloseOriginal, _ := NewFromString(day1expectedCloseOriginalStr)
	day1Close := day1.PriceClose
	if !day1Close.Equal(day1ExpectedCloseOriginal) {
		t.Errorf("Unexpected close price on 1 day, expected %v actual %v", day1ExpectedCloseOriginal, day1Close)
	}

	day1ExpectedCloseConverted, _ :=  NewFromString(day1ExpectedCloseConvertedStr)
	day1CloseConverted := day1.PriceClosePounds
	if !day1CloseConverted.Value.Equal(day1ExpectedCloseConverted) {
		t.Errorf("Unexpected close price converted on 1 day, expected %v actual %v", day1ExpectedCloseConverted, day1CloseConverted)
	}

	day1DateExpected, _ := time.Parse("2006-01-02 03:04", day1DateStr)
	day1Date := day1.Date
	if !day1DateExpected.Equal(day1Date.Time) {
		t.Errorf("Unexpected date on 1 day, expected %v actual %v", day1DateExpected, day1Date)
	}

	day6 := days[5]
	day6ExpectedClose, _ := NewFromString(day6ExpectedCloseStr)
	day6Close := day6.PriceClosePounds
	if !day6Close.Value.Equals(day6ExpectedClose) {
		t.Errorf("Unexpected close price on 6 day, expected %v actual %v", day6ExpectedClose, day6Close)
	}

	day6DateExpected, _ := time.Parse("2006-01-02 03:04", day6DateStr)
	day6Date := day6.Date
	if !day6DateExpected.Equal(day6Date.Time) {
		t.Errorf("Unexpected date on 6 day, expected %v actual %v", day6DateExpected, day6Date)
	}

	changePercentDesc := detail.GetChangePercentDesc()
	day2CloseExpected, _ := NewFromString(day2ClosePriceStr)

	day1ClosePriceActual := detail.GetPriceLastClosePoundsDesc()
	if day1ClosePriceActual != day1ClosePriceExpected {
		t.Errorf("Unexpected last close price, expected %v actual %v", day1ClosePriceExpected, day1ClosePriceActual)
	}

	pricePreviousCloseDesc := detail.GetPricePreviousCloseDesc()
	if pricePreviousCloseDesc != pricePreviousCloseDescExp {
		t.Errorf("Unexpected previous close price, expected %v actual %v", pricePreviousCloseDescExp, pricePreviousCloseDesc)
	}

	changePercentExpected := ((day1ExpectedCloseOriginal.Sub(day2CloseExpected)).Div(day2CloseExpected)).Mul(NewFromInt(100))
	changePercentExpectedDesc := GetPercentDesc(changePercentExpected)
	if changePercentDesc != changePercentExpectedDesc {
		t.Errorf("Incorrect change percent, expected %v actual %v", changePercentExpected, changePercentDesc)
	}
}

func TestCalculatePercentageChangeFromReference(t *testing.T) {

	priceClosePounds := Money {
		Currency: CURRENCY_GBP,
		Value: DecimalExt{NewFromFloat(423.90)},
	}

	eod := EodMarketStack{
		PriceClosePounds: priceClosePounds,
	}

	addedPriceBuy := Money {
		Currency: CURRENCY_GBP,
		Value: DecimalExt{NewFromFloat(411.45)},
	}

	wd := WatchDetail{
		Stock: &Stock{},
		Watch: Watch{
			AddedPriceBuy: addedPriceBuy,
		},
		History: PriceHistory{
			Eods: []EodMarketStack{eod},
		},
	}

	expected := "3.026 %"
	result := wd.GetDeltaReferencePercentDesc()
	if result != expected {
		t.Errorf("Delta percent expected %v actual %v", expected, result)
	}
}

func TestGetDeltaReferencePercentDesc(t *testing.T) {
	key := getConversionKey(CURRENCY_USD, CURRENCY_GBP)
	currencyConverter[key], _ = NewFromString("1")

	detail := getWatchDetailUsd()

	addedPriceBuy := Money{
		Currency: CURRENCY_GBP,
		Value:    DecimalExt{NewFromFloat(420.00)},
	}
	//test growth
	detail.Watch.AddedPriceBuy = addedPriceBuy //was 420
	actual := detail.GetDeltaReferencePercentDesc()        //is 434.0
	expected := "3.333 %"
	if actual != expected {
		t.Errorf("Expected %v Actual %v", expected, actual)
	}

	addedPriceBuy2 := Money{
		Currency: CURRENCY_GBP,
		Value:    DecimalExt{NewFromFloat(450.00)},
	}

	//test loss
	detail.Watch.AddedPriceBuy = addedPriceBuy2 //was 450
	actual = detail.GetDeltaReferencePercentDesc()         //is 434.0
	expected = "-3.556 %"
	if actual != expected {
		t.Errorf("Expected %v Actual %v", expected, actual)
	}
}

func TestGetCurrencyConversion(t *testing.T) {
	conversionValue := GetConversionValue(CURRENCY_GBP, CURRENCY_USD)
	Log(conversionValue.String())
}

func TestCurrencyConversion(t *testing.T) {
	from := Money{
		Currency: CURRENCY_GBP,
		Value:    DecimalExt{NewFromInt(1)},
	}

	toCurrency := from.toCurrency(CURRENCY_USD)
	Log(toCurrency.GetDesc())
}
