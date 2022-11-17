package leven

import "github.com/c9s/bbgo/pkg/fixedpoint"

var Two fixedpoint.Value = fixedpoint.NewFromInt(2)
var Three fixedpoint.Value = fixedpoint.NewFromInt(3)
var Four fixedpoint.Value = fixedpoint.NewFromInt(4)
var Delta fixedpoint.Value = fixedpoint.NewFromFloat(0.00001)
var BUY = "1"
var SELL = "-1"
var HOLD = "0"
var holdingMax = 5
