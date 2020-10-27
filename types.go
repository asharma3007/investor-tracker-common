package investor_tracker_common

import (
	"database/sql"
	"fmt"
	. "github.com/shopspring/decimal"
	"os"
)

const (
	EnvDatabaseUrl      = "DATABASE_URL"
	EnvDatabasePassword = "DATABASE_PASSWORD"
	port                = "3306"
	user                = "root"
	dbname              = "tracker"

	EnvUrlEmailQueue 	= "URL_EMAIL_QUEUE"

	EnvTokenMarketStack = "TOKEN_MARKETSTACK"

	CloudfunctionSourceDir = "serverless_function_source_code"
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


type MessageSendEmail struct {
	Html string
	PlainText string
	SenderName string
	Subject string
}



func FixTemplateSourceDir() {
	fileInfo, err := os.Stat(CloudfunctionSourceDir)
	if err == nil && fileInfo.IsDir() {
		_ = os.Chdir(CloudfunctionSourceDir)
	}
}

func Log(message string) {
	fmt.Println(message)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ConnectDb() *sql.DB {
	Log("Connecting to db")

	dbUrl := os.Getenv(EnvDatabaseUrl)
	password := os.Getenv(EnvDatabasePassword)

	Log("Connection string")
	// connection string
	mysqlconn := user + ":" + password + "@tcp(" + dbUrl + ":" + port +")/" + dbname

	Log("Opening db")
	// open database
	db, err := sql.Open("mysql", mysqlconn)
	CheckError(err)

	Log("Doing ping")
	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	return db
}
