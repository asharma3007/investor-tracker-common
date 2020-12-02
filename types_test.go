package common

import (
	"encoding/json"
	. "github.com/shopspring/decimal"
	"testing"
	"time"
)

const marketStackResponse = "{\"pagination\":{\"limit\":100,\"offset\":0,\"count\":6,\"total\":6},\"data\":[{\"open\":430.13,\"high\":434.5899,\"low\":426.4601,\"close\":434.0,\"volume\":28925656.0,\"adj_high\":434.5899,\"adj_low\":426.4601,\"adj_close\":434.0,\"adj_open\":430.13,\"adj_volume\":28925656.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-09T00:00:00+0000\"},{\"open\":438.44,\"high\":439.0,\"low\":425.3,\"close\":425.92,\"volume\":40421116.0,\"adj_high\":439.0,\"adj_low\":425.3,\"adj_close\":425.92,\"adj_open\":438.44,\"adj_volume\":40421116.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-08T00:00:00+0000\"},{\"open\":419.87,\"high\":429.9,\"low\":413.845,\"close\":425.3,\"volume\":43127709.0,\"adj_high\":429.9,\"adj_low\":413.845,\"adj_close\":425.3,\"adj_open\":419.87,\"adj_volume\":43127709.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-07T00:00:00+0000\"},{\"open\":423.79,\"high\":428.7799,\"low\":406.05,\"close\":413.98,\"volume\":49146259.0,\"adj_high\":428.7799,\"adj_low\":406.05,\"adj_close\":413.98,\"adj_open\":423.79,\"adj_volume\":49146259.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-06T00:00:00+0000\"},{\"open\":423.35,\"high\":433.64,\"low\":419.33,\"close\":425.68,\"volume\":44722786.0,\"adj_high\":433.64,\"adj_low\":419.33,\"adj_close\":425.68,\"adj_open\":423.35,\"adj_volume\":44722786.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-05T00:00:00+0000\"},{\"open\":421.39,\"high\":439.13,\"low\":415.0,\"close\":415.09,\"volume\":71430025.0,\"adj_high\":439.13,\"adj_low\":415.0,\"adj_close\":415.09,\"adj_open\":421.39,\"adj_volume\":71430025.0,\"symbol\":\"TSLA\",\"exchange\":\"XNAS\",\"date\":\"2020-10-02T00:00:00+0000\"}]}"

func getWatchDetailTesla() WatchDetail {
	exampleStock := Stock{
		Description: "Tesla, Inc.",
		Symbol: "TSLA",
	}

	var responseDays ResponseMarketStack
	err := json.Unmarshal([]byte(marketStackResponse), &responseDays)
	CheckError(err)

	var detail WatchDetail
	detail.History.Eods = responseDays.Data
	detail.Stock = exampleStock
	return detail
}

func TestParseMarketStackResponse(t *testing.T) {
	detail := getWatchDetailTesla()
	days := detail.History.Eods
	numberDays := len(days)
	if numberDays != 6 {
		t.Errorf("Unexpected number of days, expected %v actual %v", 6, numberDays)
	}

	day1 := days[0]
	day1ExpectedClose, _ := NewFromString("434.0")
	day1Close := day1.PriceClosePounds
	if !day1Close.Equals(day1ExpectedClose) {
		t.Errorf("Unexpected close price on 1 day, expected %v actual %v", day1ExpectedClose, day1Close)
	}

	day1DateExpected,_ := time.Parse("2006-01-02 03:04", "2020-10-09 00:00")
	day1Date := day1.Date
	if !day1DateExpected.Equal(day1Date.Time) {
		t.Errorf("Unexpected date on 1 day, expected %v actual %v", day1DateExpected, day1Date)
	}

	day6 := days[5]
	day6ExpectedClose, _ := NewFromString("415.09")
	day6Close := day6.PriceClosePounds
	if !day6Close.Equals(day6ExpectedClose) {
		t.Errorf("Unexpected close price on 6 day, expected %v actual %v", day6ExpectedClose, day6Close)
	}

	day6DateExpected,_ := time.Parse("2006-01-02 03:04", "2020-10-02 00:00")
	day6Date := day6.Date
	if !day6DateExpected.Equal(day6Date.Time) {
		t.Errorf("Unexpected date on 6 day, expected %v actual %v", day6DateExpected, day6Date)
	}

	changePercentDesc := detail.GetChangePercentDesc()
	day2CloseExpected, _ := NewFromString("425.92")

	day1ClosePriceActual := detail.GetPriceLastCloseDesc()
	day1ClosePriceExpected := "43400"
	if day1ClosePriceActual != day1ClosePriceExpected {
		t.Errorf("Unexpected last close price, expected %v actual %v", day1ClosePriceExpected, day1ClosePriceActual)
	}

	pricePreviousCloseDesc := detail.GetPricePreviousCloseDesc()
	pricePreviousCloseDescExp := "42592"
	if pricePreviousCloseDesc != pricePreviousCloseDescExp {
		t.Errorf("Unexpected last close price, expected %v actual %v", pricePreviousCloseDescExp, pricePreviousCloseDesc)
	}

	changePercentExpected := ((day2CloseExpected.Sub(day1ExpectedClose)).Div(day2CloseExpected)).Mul(NewFromInt(100))
	changePercentExpectedDesc := GetPercentDesc(changePercentExpected)
	if changePercentDesc != changePercentExpectedDesc {
		t.Errorf("Incorrect change percent, expected %v actual %v", changePercentExpected, changePercentDesc)
	}
}


func TestCalculatePercentageChangeFromReference(t *testing.T) {

	eod := EodMarketStack{
		PriceClosePounds: NewFromInt(42390),
	}

	wd := WatchDetail{
		Stock:   Stock{},
		Watch:   Watch{
			AddedPriceBuy: NewFromInt(41145),
		},
		History: PriceHistory{
			Eods: []EodMarketStack{eod},
		},
	}

	expected := "-3.026 %"
	result := wd.GetDeltaReferencePercentDesc()
	if result != expected {
		t.Errorf("Delta percent expected %v actual %v", expected, result)
	}
}

func TestGetDeltaReferencePercentDesc(t *testing.T) {
	detail := getWatchDetailTesla()

	//test growth
	detail.Watch.AddedPriceBuy, _ = NewFromString("42000") //was
	actual := detail.GetDeltaReferencePercentDesc() //is 434.0
	expected := "3.333 %"
	if actual != expected {
		t.Errorf("Expected %v Actual %v", expected, actual)
	}

	//test loss
	detail.Watch.AddedPriceBuy, _ = NewFromString("45000") //was
	actual = detail.GetDeltaReferencePercentDesc() //is 434.0
	expected = "-3.556 %"
	if actual != expected {
		t.Errorf("Expected %v Actual %v", expected, actual)
	}
}

func TestStock_PopulateCurrentPrice(t *testing.T) {

}
