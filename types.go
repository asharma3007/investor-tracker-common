package investor_tracker_common

import (
	. "github.com/shopspring/decimal"
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
