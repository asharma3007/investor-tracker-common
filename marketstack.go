package common

import (
	"encoding/json"
	"fmt"
	. "github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
	"time"
)

const (
	TimeFormatRequest = "2006-01-02"
)

type RequestEndOfDay struct {
	RequestCommon
	DateFrom time.Time
	DateTo   time.Time
	SortMode string
	Limit    int
	Offset   int
}

func (request *RequestEndOfDay) GetUrl() string {
	symbols := strings.Join(request.Symbols, ",")
	dateFromStr := request.DateFrom.Format(TimeFormatRequest)
	dateToStr := request.DateTo.Format(TimeFormatRequest)

	token := GetSecret(EnvSecretTokenMarketStack)

	return fmt.Sprintf("http://api.marketstack.com/v1/eod?symbols=%v&access_key=%v&date_from=%v&date_to=%v&limit=%v",
		symbols,
		token,
		dateFromStr,
		dateToStr,
		request.Limit)
}

type RequestCommon struct {
	Symbols []string
}

type ResponseMarketStack struct {
	Pagination Pagination       `json:"pagination"`
	Data       []EodMarketStack `json:"data"`
}

func (resp *ResponseMarketStack) PopulateUsablePrice(stock *Stock) {
	for ix, _ := range resp.Data {
		resp.Data[ix].PopulateUsablePrice(stock)
	}
}

func (resp *ResponseMarketStack) GetExchange() string {
	if len(resp.Data) == 0 {
		Log("No EODs on ResponseMarketData")
		return ""
	}

	return resp.Data[0].Exchange
}

type Pagination struct {
	Limit  int
	Offset int
	Count  int
	Total  int
}

func QueryEndOfDayMarketStack(client HttpSource, request RequestEndOfDay) ResponseMarketStack {

	url := request.GetUrl()

	log := fmt.Sprintf("QueryEndOfDayMarketStack price history for %v from URL: %v", request.Symbols, url)
	Log(log)

	response, err := client.HttpGet(url)
	CheckError(err)

	defer response.Body.Close()

	//json.NewDecoder(response.Body).Decode(target)
	//var str string
	//decoder := json.NewDecoder(response.Body)
	//decoder.UseNumber()

	responseData, err := ioutil.ReadAll(response.Body)
	CheckError(err)

	responseString := string(responseData)
	Log(responseString)

	var retval ResponseMarketStack
	err = json.Unmarshal(responseData, &retval)
	CheckError(err)

	return retval
}

func (eod *EodMarketStack) GetPriceCloseDesc() string {
	return eod.PriceClosePounds.GetDesc()
}

type PriceHistory struct {
	Eods []EodMarketStack
}

type EodMarketStack struct {
	Date             timeMarketStack `json:"date"`
	PriceClose	Decimal	`json:"close"`
	Exchange string `json:"exchange"`
	PriceClosePounds Money         `json:"-"`
}

func (eod *EodMarketStack) GetPriceClosePence() Decimal {
	return eod.PriceClosePounds.Value.Mul(NewFromInt(100))
}

func (eod *EodMarketStack) Dump() {
	LogDebug(eod.Date.Format("06 Jan 02") + " " + eod.PriceClosePounds.GetDesc())
}

func (eod *EodMarketStack) PopulateUsablePrice(stock *Stock) {
	if strings.Contains(stock.Exchange, ExchangeUsa) {
		usd := Money{
			Currency: CURRENCY_USD,
			Value:    DecimalExt{ eod.PriceClose},
		}

		eod.PriceClosePounds = usd.toCurrency(CURRENCY_GBP)
	} else {
		eod.PriceClosePounds =  Money{
			Currency: CURRENCY_GBP,
			Value:    DecimalExt{eod.PriceClose.Mul(NewFromInt(1))},
		}
	}
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
