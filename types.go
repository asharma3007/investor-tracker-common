package common

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	. "github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	EnvDatabaseUrl      = "DATABASE_URL"
	EnvDatabasePort     = "DATABASE_PORT"
	EnvDatabaseName     = "DATABASE_NAME"
	EnvSecretDbUser     = "SECRET_DATABASE_USER"
	EnvSecretDbPassword = "SECRET_DATABASE_PASSWORD"

	EnvTokenMarketStack = "TOKEN_MARKETSTACK"
	EnvTokenIex         = "TOKEN_IEX"

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

type MonitorInstruction struct {
	StockId            int
	PriceTypeToMonitor int
	MarkerPrice        Decimal
	Message            string
	Holding            Holding
}

func (instruction MonitorInstruction) GetDesc() string {
	if instruction.IsSell() {
		return "Sell"
	}
	return "Buy"
}

func (instruction MonitorInstruction) IsBuy() bool {
	return instruction.PriceTypeToMonitor == PriceTypeBuy
}

func (instruction MonitorInstruction) IsSell() bool {
	return instruction.PriceTypeToMonitor == PriceTypeSell
}


func (holding Holding) GetUnitsTotal() Decimal {
	retVal := NewFromInt(0)
	for _, lot := range holding.Lots {
		retVal = retVal.Add(lot.Units)
	}
	return retVal
}

func (stock Stock) GetRelevantPrice(instruction MonitorInstruction) Decimal {
	if instruction.PriceTypeToMonitor == PriceTypeBuy {
		return stock.PriceBuy
	} else if instruction.PriceTypeToMonitor == PriceTypeSell {
		return stock.PriceSell
	} else {
		panic(fmt.Sprint("Unknown price type ", instruction.PriceTypeToMonitor))
	}
}

func (lot Lot) GetValueTotalBought() Decimal {
	return lot.PriceBought.Mul(lot.Units)
}

func (holding Holding) GetPriceAverageBought() Decimal {
	totalValue := holding.GetValueTotalBought()

	totalUnits := holding.GetUnitsTotal()
	return totalValue.Div(totalUnits)
}

func (holding *Holding) GetValueTotalBought() Decimal {
	totalValue := NewFromInt(0)
	for _, lot := range holding.Lots {
		lotValue := lot.GetValueTotalBought()
		totalValue = totalValue.Add(lotValue)
	}
	return totalValue
}

type WatchDetail struct {
	Stock Stock
	Watch Watch
	History PriceHistory
	ChangePercent Decimal
}

type Watch struct {
	WatchId        int
	StockId        int
	DtReference    string
	AddedPriceBuy  Decimal
	AddedPriceSell Decimal
	AlertThreshold Decimal
	Notes		   string
}

func (wd *WatchDetail) GetPriceLastCloseDesc() string {
	priceLastClose := wd.GetPriceLastClose()
	if priceLastClose.IsNegative() { return "No price history"}

	return GetPriceDesc(priceLastClose)
}

func (w *Watch) GetAlertThresholdDesc() string {
	return GetPercentDesc(w.AlertThreshold)
}

func (w *Watch) GetPriceBuyDesc() string {
	rounded := w.AddedPriceBuy.Round(3)
	return fmt.Sprintf("%vp", rounded)
}

type watchDetailIex struct {
	Stock Stock
	Watch Watch
	PriceOpen Decimal `json:"open"`
	PriceClose Decimal `json:"iexClose"`
	PriceHigh Decimal `json:"high"`
	ChangePercent Decimal `json:"changePercent"`
	Volume Decimal `json:"volume"`
	AverageVolume Decimal `json:"avgTotalVolume"`
	PriceBid Decimal `json:"iexBidPrice"`
	PriceAsk Decimal `json:"iexAskPrice"`
	PricePreviousClose Decimal `json:"previousClose"`
	History PriceHistory
}

func (eod *EodMarketStack) GetPriceCloseDesc() string {
	return GetPriceDesc(eod.PriceClose)
}

func GetPriceDesc(price Decimal) string {
	return fmt.Sprintf("%v", convertPenceToPounds(price))
}

func (wd *WatchDetail) GetPricePreviousCloseDesc() string {
	if len(wd.History.Eods) <= 1 {
		return "No previous"
	}

	previousDay := wd.History.Eods[1]
	previousClose := previousDay.PriceClose

	return fmt.Sprintf("%v", convertPenceToPounds(previousClose))
}

func convertPenceToPounds(price Decimal) Decimal {
	return price.Mul(NewFromInt(100))
}

func (wd *WatchDetail) GetChangePercentDesc() string {
	if !wd.ChangePercent.IsZero() {
		return GetPercentDesc(wd.ChangePercent)
	}

	if len(wd.History.Eods) > 1 {
		lastDay := wd.History.Eods[0]
		previousDay := wd.History.Eods[1]
		lastClose := lastDay.PriceClose
		previousClose := previousDay.PriceClose
		changePercent := getPercent(previousClose, lastClose)
		return GetPercentDesc(changePercent)
	}

	Log(fmt.Sprintf("Could not get percent change from watchdetail %v", wd))
	return ""
}

func getPercent(one Decimal, two Decimal) Decimal {
	change := one.Sub(two)
	ratio := change.Div(one)
	return ratio.Mul(NewFromInt(100))
}

func GetPercentDesc(percent Decimal) string {
	return fmt.Sprintf("%v %%", percent.Round(3))
}

const TimeFormatMySql = "2006-01-02 15:04:05"
const TimeFormatPostGres = time.RFC3339Nano

func (wd *WatchDetail) GetDtReferenceDesc() string {
	//postgres
	//parse, err := time.Parse(TimeFormatPostGres, wd.Watch.DtReference)
	//mysql
	parse, err := time.Parse(TimeFormatMySql, wd.Watch.DtReference)
	CheckError(err)
	return parse.Format(time.RFC822)
}

func (wd *WatchDetail) GetDeltaReferencePercentDesc() string {
	priceStartWatch := wd.Watch.AddedPriceBuy
	priceChange := priceStartWatch.Sub(wd.GetPriceLastClose())
	percent := getPercent(priceStartWatch, priceChange)
	return GetPercentDesc(percent)
}

type PriceHistory struct {
	Eods []EodMarketStack
}

type EodMarketStack struct {
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


func (wd *WatchDetail) GetPriceLastClose() Decimal {
	if len(wd.History.Eods) == 0 { return NewFromInt(-1) }

	lastEod := wd.History.Eods[0]
	return lastEod.PriceClose
}

type ResponseMarketStack struct {
	Pagination Pagination // `json:"pagination"`
	Data []EodMarketStack `json:"data"`
}

type Pagination struct {
	Limit int
	Offset int
	Count int
	Total int
}


func (stock *Stock) GetPriceUrl() string {
	if stock.IsSourceHl() {
		return stock.getHlUrl()
	} else {
		return stock.getMarketStackUrl()
	}
}

func (stock *Stock) getHlUrl() string {
	var urlToUse string
	// need to override the full URL for some of them
	if strings.HasPrefix(stock.HlUrlOverride, "http") {
		urlToUse = stock.HlUrlOverride
	} else {

		var urlSuffix string
		if len(stock.HlUrlOverride) != 0 {
			urlSuffix = stock.HlUrlOverride
		} else {
			urlSuffix = getHlUrlFromName(stock.HlName)
		}
		urlToUse = "https://www.hl.co.uk/funds/fund-discounts,-prices--and--factsheets/search-results/" + urlSuffix
	}

	return urlToUse
}

func (stock *Stock) getMarketStackUrl() string {
	today := time.Now()
	weekAgo := today.Add(-time.Hour * 24 * 7)

	todayStr := today.Format("2006-01-02")
	weekAgoStr := weekAgo.Format("2006-01-02")

	return fmt.Sprintf("http://api.marketstack.com/v1/eod?symbols=%v&access_key=%v&date_from=%v&date_to=%v",
		stock.Symbol,
		os.Getenv(EnvTokenMarketStack),
		weekAgoStr,
		todayStr)
}

func getHlUrlFromName(hlName string) string {
	retVal := strings.ReplaceAll(hlName, " ", "-")
	retVal = strings.ReplaceAll(retVal, "---", "-")
	retVal = strings.ReplaceAll(retVal, "%", "")
	retVal = strings.ReplaceAll(retVal, "&", "and")
	retVal = strings.ToLower(retVal)

	retVal = retVal[:1] + "/" + retVal

	return retVal
}

func (stock Stock) PopulateCurrentPrice() Stock {
	if stock.IsSourceHl() {
		stock.populateFromHl()
	} else {
		stock.populateFromMarketStack()
	}

	return stock
}

//go:generate mockgen -destination=mocks/mock_httpsource.go -package=cloudfunction . HttpSource
type HttpSource interface {
	HttpGet(url string) (*http.Response, error)
}

func (client *DefaultHttp) HttpGet(url string) (*http.Response, error) {
	return http.Get(url)
}

type DefaultHttp struct {

}

func BuildWatchDetail(client *DefaultHttp, stock Stock) WatchDetail {
	if stock.IsSourceHl() {
		return buildWatchDetailHl(stock)
		//return watchDetail{}
	} else {
		return BuildWatchDetailMarketStack(client, stock)
		//return watchDetail{}
	}
}

func buildWatchDetailHl(stock Stock) WatchDetail {
	Log(fmt.Sprintf("Getting history from HL for %v from %v", stock.ToString(), stock.Url))

	stockPage, err := http.Get(stock.Url)
	CheckError(err)

	stockDoc, _ := goquery.NewDocumentFromReader(stockPage.Body)

	priceBuyStr, err := stockDoc.Find(".ask.price-divide").Html()
	CheckError(err)
	priceSellStr, err := stockDoc.Find(".bid.price-divide").Html()
	CheckError(err)

	selectionPcChange := stockDoc.Find("span.change-divide > span:nth-child(3)")
	percentChangeStr, err := selectionPcChange.Html()
	CheckError(err)
	percentChangeStr = strings.TrimSpace(percentChangeStr)
	reg, _ := regexp.Compile("[^0-9.]+")
	percentChangeStr = reg.ReplaceAllString(percentChangeStr, "")

	isNegative := selectionPcChange.HasClass("negative change")
	if isNegative { percentChangeStr = "-" + percentChangeStr }
	percentChange, err := NewFromString(percentChangeStr)
	CheckError(err)

	if len(priceSellStr) == 0 || len(priceSellStr) == 0 {
		Log("Failed to get a price for stock")
		message := fmt.Sprint(stock.Description, " Url: ", stock.Url, " Buy ", priceBuyStr, " Sell ", priceSellStr)
		Log(message)
	}

	if len(percentChangeStr) == 0 {
		Log("Failed to get a percent change for stock")
		message := fmt.Sprint(stock.Description, " Url: ", stock.Url, " String ", percentChangeStr)
		Log(message)
	}

	return WatchDetail{
		ChangePercent: percentChange,
		History: PriceHistory{
			Eods: []EodMarketStack {
				{
					Date:       timeMarketStack{time.Now()},
					PriceClose: parsePrice(priceSellStr),
				},
			},
		},
	}
}

func BuildWatchDetailMarketStack(client HttpSource, stock Stock) WatchDetail {
	log := fmt.Sprintf("Getting price history for %v from URL: %v", stock.ToString(), stock.Url)
	Log(log)

	response, err := client.HttpGet(stock.Url)
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

	var responseDays ResponseMarketStack
	err = json.Unmarshal(responseData, &responseDays)
	CheckError(err)

	var wd WatchDetail
	wd.History.Eods = responseDays.Data
	return wd
}

func buildWatchDetailIex(stock Stock) WatchDetail {
	response, err := http.Get(stock.Url)
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

	var quote WatchDetail
	err = json.Unmarshal(responseData, &quote)
	CheckError(err)

	return quote
}

func (stock *Stock) populateFromHl() {
	fullUrl := stock.GetPriceUrl()

	stockPage, err := http.Get(fullUrl)
	CheckError(err)

	stockDoc, _ := goquery.NewDocumentFromReader(stockPage.Body)

	priceBuyStr, err := stockDoc.Find(".ask.price-divide").Html()
	CheckError(err)
	priceSellStr, err := stockDoc.Find(".bid.price-divide").Html()
	CheckError(err)

	if len(priceSellStr) == 0 || len(priceSellStr) == 0 {
		Log("Failed to get a price for stock")
		message := fmt.Sprint(stock.HlName, " Url: ", fullUrl, " Buy ", priceBuyStr, " Sell ", priceSellStr)
		Log(message)
	}

	stock.PriceBuy =  parsePrice(priceBuyStr)
	stock.PriceSell= parsePrice(priceSellStr)
}

func parsePrice(priceStr string) Decimal {
	priceStr = strings.ReplaceAll(priceStr, "p", "")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	if len(priceStr) == 0 {
		priceStr = "0"
	}

	if strings.HasPrefix(priceStr, "£") {
		priceStr = strings.ReplaceAll(priceStr, "£", "")
		ixDecimal := strings.Index(priceStr, ".")
		dpGiven := len(priceStr) - (ixDecimal + 1)
		zeroesNeeded := 2 - dpGiven
		for i := 0; i <zeroesNeeded; i++ {
			priceStr += "0"
		}
		priceStr = strings.ReplaceAll(priceStr, ".", "")
	}

	price, err := NewFromString(priceStr)
	CheckError(err)
	return price
}

func (stock *Stock) populateFromMarketStack() {
	var httpClient DefaultHttp
	watchDetail := BuildWatchDetailMarketStack(&httpClient, *stock)
	stock.PriceBuy = watchDetail.GetPriceLastClose()
	stock.PriceSell = watchDetail.GetPriceLastClose()
}

func getIexUrl(stock Stock) string {
	return fmt.Sprintf("https://cloud.iexapis.com/stable/stock/%v/quote?token=%v", stock.Symbol, os.Getenv(EnvTokenIex))
}

