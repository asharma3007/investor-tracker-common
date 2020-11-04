package common

import (
	"testing"
	. "github.com/shopspring/decimal"
)

func TestCalculatePercentageChangeFromReference(t *testing.T) {

	eod := EodMarketStack{
		PriceClose: NewFromInt(42390),
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

