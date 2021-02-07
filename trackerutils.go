package common

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

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
