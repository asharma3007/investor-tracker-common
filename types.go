package investor_tracker_common

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	. "github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"os"
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

func (lot Lot) GetValueTotal() Decimal {
	return lot.PriceBought.Mul(lot.Units)
}

func (holding Holding) GetPriceAverage() Decimal {
	totalValue := NewFromInt(0)
	for _, lot := range holding.Lots {
		lotValue := lot.GetValueTotal()
		totalValue = totalValue.Add(lotValue)
	}

	totalUnits := holding.GetUnitsTotal()
	return totalValue.Div(totalUnits)
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

func buildWatchDetailMarketStack(client HttpSource, stock *Stock) WatchDetail {
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
	watchDetail := buildWatchDetailMarketStack(&httpClient, stock)
	stock.PriceBuy = watchDetail.GetPriceLastClose()
	stock.PriceSell = watchDetail.GetPriceLastClose()
}
