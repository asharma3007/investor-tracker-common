package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	. "github.com/shopspring/decimal"
)

const (
	TimeFormatRequest = "2006-01-02"
)

type RequestEndOfDay struct {
	RequestCommon
	DateFrom time.Time
	DateTo time.Time
	SortMode string
	Limit int
	Offset int
}

func (request *RequestEndOfDay) getUrl() string {
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
	Pagination Pagination `json:"pagination"`
	Data []EodMarketStack `json:"data"`
}

type Pagination struct {
	Limit int
	Offset int
	Count int
	Total int
}

func QueryEndOfDayMarketStack(client HttpSource, request RequestEndOfDay) ResponseMarketStack {

	url := request.getUrl()

	log := fmt.Sprintf("Getting price history for %v from URL: %v", request.Symbols, url)
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
	return GetPriceDesc(eod.PriceClosePounds)
}

type PriceHistory struct {
	Eods []EodMarketStack
}

type EodMarketStack struct {
	Date             timeMarketStack `json:"date"`
	PriceClosePounds Decimal         `json:"close"`
}

func (eod *EodMarketStack) GetPriceClosePence() Decimal {
	return eod.PriceClosePounds.Mul(NewFromInt(100))
}

func (eod *EodMarketStack) Dump() {
	LogDebug(eod.Date.Format("06 Jan 02") + " " + eod.PriceClosePounds.String())
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
