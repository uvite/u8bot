package scalpma

import (
	"github.com/c9s/bbgo/pkg/indicator"
	"time"

	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/types"
)

/*
macd implements moving average convergence divergence indicator

Moving Average Convergence Divergence (MACD)
- https://www.investopedia.com/terms/m/macd.asp
- https://school.stockcharts.com/doku.php?id=technical_indicators:macd-histogram
*/
type CrossType string

const (
	OnlineOver   = CrossType("ONLINEOVER")
	OnlineUnder  = CrossType("ONLINEUNDER")
	OfflineOver  = CrossType("OFFLINEOVER")
	OfflineUnder = CrossType("OFFLINEUNDER")

	CrossOver  = CrossType("CROSSOVER")
	CrossUnder = CrossType("CROSSUNDER")

	NoThing = CrossType("NOTHING")
)

func (cross CrossType) String() string {
	return string(cross)
}

type OK struct {
	types.IntervalWindow
	types.SeriesBase

	Values floats.Slice
}

var _ types.SeriesExtend = &OK{}

func (inc *OK) Update(value float64) {

	if len(inc.Values) == 0 {
		inc.SeriesBase.Series = inc
		inc.Values.Push(value)
		return
	}
	inc.Values.Push(value)
}

//go:generate callbackgen -type MACD
type DEMACD struct {
	types.IntervalWindow     // 9
	ShortPeriod          int // 12
	LongPeriod           int // 26
	DeaPeriod            int
	MaType               string
	FastMa               types.UpdatableSeriesExtend
	SlowMa               types.UpdatableSeriesExtend
	Dea                  types.UpdatableSeriesExtend
	Dif                  floats.Slice

	//Dif floats.Slice

	//Macd floats.Slice

	Values floats.Slice

	EndTime time.Time

	updateCallbacks []func(value float64)
}

func (inc *DEMACD) Update(x float64) {
	//var FastMa
	if len(inc.Values) == 0 {
		switch inc.MaType {
		case "EWMA":
			inc.FastMa = &indicator.EWMA{IntervalWindow: types.IntervalWindow{Window: inc.ShortPeriod}}
			inc.SlowMa = &indicator.EWMA{IntervalWindow: types.IntervalWindow{Window: inc.LongPeriod}}
			inc.Dea = &indicator.EWMA{IntervalWindow: types.IntervalWindow{Window: inc.DeaPeriod}}
		case "DEMA":
			inc.FastMa = &indicator.DEMA{IntervalWindow: types.IntervalWindow{Window: inc.ShortPeriod}}
			inc.SlowMa = &indicator.DEMA{IntervalWindow: types.IntervalWindow{Window: inc.LongPeriod}}
			inc.Dea = &indicator.DEMA{IntervalWindow: types.IntervalWindow{Window: inc.DeaPeriod}}

		}
	}

	// update fast and slow ema
	inc.FastMa.Update(x)
	inc.SlowMa.Update(x)

	// update macd
	dif := inc.FastMa.Last() - inc.SlowMa.Last()

	//values 相当于dif sigal 相当于慢线 histogram 相当于macd
	inc.Dif.Push(dif)
	// update signal line
	inc.Dea.Update(dif)

	// update histogram  dea
	inc.Values.Push((dif - inc.Dea.Last()) * 2)
}

func (inc *DEMACD) Last() float64 {
	if len(inc.Values) == 0 {
		return 0.0
	}

	return inc.Values[len(inc.Values)-1]
}

func (inc *DEMACD) Length() int {
	return len(inc.Values)
}

func (inc *DEMACD) PushK(k types.KLine) {
	inc.Update(k.Close.Float64())
}

func (inc *DEMACD) Difs() float64 {
	return inc.Dif.Last()
}
func (inc *DEMACD) Deas() float64 {
	return inc.Dea.Last()
}

func (inc *DEMACD) Cross() CrossType {

	//floats.CrossOver(inc.Dif, inc.Dea.Array())

	crossMaOver := floats.CrossOver(inc.Dif, inc.Dea.Array())
	crossMaUnder := floats.CrossUnder(inc.Dif, inc.Dea.Array())
	if crossMaOver {
		return CrossOver
	}
	if crossMaUnder {
		return CrossUnder
	}
	return NoThing
	//if crossMaOver.Last() {
	//	return CrossOver
	//}
	//if crossMaUnder.Last() {
	//	return CrossUnder
	//}
	//crossSignOver := floats.CrossOver(inc.Dif, inc.Dea.Array())
	//crossMaUnder := types.CrossUnder(inc.FastMa, inc.SlowMa)
	//crossSignUnder := floats.CrossUnder(inc.Dif, inc.Dea.Array())
	//
	//if crossMaOver.Last() && crossSignOver {
	//	return OnlineOver
	//}
	//if crossMaOver.Last() && crossSignUnder {
	//	return OfflineOver
	//}
	//
	//if crossMaUnder.Last() && crossSignOver {
	//	return OnlineUnder
	//}
	//if crossMaUnder.Last() && crossSignUnder {
	//	return OfflineUnder
	//}
	return NoThing

}
