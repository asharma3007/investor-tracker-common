package investor_tracker_common

import (
	. "github.com/shopspring/decimal"
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
