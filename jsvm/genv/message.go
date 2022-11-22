package genv

import (
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

var OrderOpenLong string = "Order.Open.Long"
var OrderCloseLong string = "Order.Close.Long"
var OrderOpenShort string = "Order.Open.Short"
var OrderCloseShort string = "Order.Close.Short"
var MessageShow string = "Message.Debug"

type OrderPayload struct {
	Symbol   string           `json:"symbol" db:"symbol"`
	Side     types.SideType   `json:"side" db:"side"`
	Quantity fixedpoint.Value `json:"quantity" db:"quantity"`
	Price    fixedpoint.Value `json:"price" db:"price"`
}
