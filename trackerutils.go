package common

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func CheckApiKeyRequest(request *http.Request) {

	method := request.Method

	var apiKeys []string
	if method == "GET" {
		queryValues, err := url.ParseQuery(request.URL.RawQuery)
		CheckError(err)

		apiKeys = queryValues["api_key"]
	} else if (method == "POST") {
		request.ParseForm()
		apiKeys = request.Form["api_key"]
	} else {
		panic(fmt.Sprintf("Unknown method %v", method))
	}
	if len(apiKeys) != 1 {
		panic("Expected 1 apikey, got " + strconv.Itoa(len(apiKeys)))
	}
	apiKey := apiKeys[0]
	checkApiKey(apiKey)
}

func checkApiKey(apiKey string) {
	secret := GetSecret(EnvSecretMyApiKey)
	if secret != apiKey || len(apiKey) == 0 {
		panic(fmt.Sprintf("Invalid api key [%v]", apiKey))
	}
}

func GetStocksReference(db *mongo.Database) map[string]*Stock {
	Log("Getting stocks reference data...")

	ctx := context.TODO()

	collectionStock := db.Collection("stock")

	cursor, err := collectionStock.Find(ctx, bson.M{})
	CheckError(err)

	//var docsStock []bson.M
	//err = cursor.All(ctx, &docsStock)

	stocks := map[string]*Stock{}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var stock Stock
		err = cursor.Decode(&stock)
		CheckError(err)

		urlToUse := stock.GetPriceUrl()
		stock.Url = urlToUse

		debugOverride := os.Getenv("DEBUG_STOCK")
		if len(debugOverride) > 0 &&  debugOverride != stock.StockId {
			continue
		}

		stocks[stock.StockId] = &stock
	}

	Log(fmt.Sprint("Got ", len(stocks), " stocks"))
	return stocks
}
