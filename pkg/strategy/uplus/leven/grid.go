package leven

import (
	"github.com/c9s/bbgo/pkg/strategy/uplus/indi"
	"time"

	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
)

var zeroTime time.Time

//go:generate callbackgen -type GRID
type GRID struct {
	types.IntervalWindow
	types.SeriesBase
	MaType    string
	Ma        types.UpdatableSeriesExtend
	Dmi       *indicator.DMI
	Smoothing int
	Power     float64
	Phase     float64

	Values floats.Slice

	EndTime time.Time

	updateCallbacks []func(value float64)
}

var _ types.SeriesExtend = &GRID{}

func (inc *GRID) Update(high, low, price float64) {
	if inc.SeriesBase.Series == nil {
		inc.SeriesBase.Series = inc

		switch inc.MaType {
		case "EWMA":
			inc.Ma = &indicator.EWMA{IntervalWindow: types.IntervalWindow{Window: inc.Window}}

		case "DEMA":
			inc.Ma = &indicator.DEMA{IntervalWindow: types.IntervalWindow{Window: inc.Window}}
		case "JMA":
			inc.Ma = &indi.JMA{IntervalWindow: types.IntervalWindow{Window: inc.Window}, Phase: inc.Phase, Power: inc.Power}
		case "HMA":
			inc.Ma = &indicator.HULL{IntervalWindow: types.IntervalWindow{Window: inc.Window}}
		case "ALMA":
			inc.Ma = &indicator.ALMA{IntervalWindow: types.IntervalWindow{Window: inc.Window}, Offset: 0.9, Sigma: 6}

		}

		inc.Ma = &indi.JMA{IntervalWindow: types.IntervalWindow{Window: 21}, Phase: 100, Power: 1}

		inc.Dmi = &indicator.DMI{IntervalWindow: types.IntervalWindow{Window: inc.Window}, ADXSmoothing: inc.Smoothing}

	}
	//fmt.Println(inc.MaType, inc.Window, inc.PhaseRatio, inc.Power)
	inc.Dmi.Update(high, low, price)
	result := inc.Dmi.GetDIPlus().Last() - inc.Dmi.GetDIMinus().Last()
	//fmt.Println(price, inc.Dmi.GetDIPlus().Last(), inc.Dmi.GetDIMinus().Last(), result)
	inc.Ma.Update(result)
	//fmt.Println(price, result, inc.Ma.Last())
	//inc.Ma.Update(price)

	inc.Values.Push(inc.Ma.Last())

}

func (inc *GRID) Last() float64 {
	if len(inc.Values) == 0 {
		return 0
	}

	return inc.Values[len(inc.Values)-1]
}

func (inc *GRID) Index(i int) float64 {
	if i >= len(inc.Values) {
		return 0
	}

	return inc.Values[len(inc.Values)-1-i]
}

func (inc *GRID) Length() int {
	return len(inc.Values)
}

func (inc *GRID) BindK(target indicator.KLineClosedEmitter, symbol string, interval types.Interval) {
	target.OnKLineClosed(types.KLineWith(symbol, interval, inc.PushK))
}

func (inc *GRID) PushK(k types.KLine) {
	if inc.EndTime != zeroTime && k.EndTime.Before(inc.EndTime) {
		return
	}

	inc.Update(indicator.KLineHighPriceMapper(k), indicator.KLineLowPriceMapper(k), indicator.KLineClosePriceMapper(k))
	inc.EndTime = k.EndTime.Time()
	inc.EmitUpdate(inc.Last())
}

func (inc *GRID) LoadK(allKLines []types.KLine) {
	for _, k := range allKLines {
		inc.PushK(k)
	}
	inc.EmitUpdate(inc.Last())
}
