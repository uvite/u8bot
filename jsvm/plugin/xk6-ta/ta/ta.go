// Package timers is implementing setInterval setTimeout and co.
package ta

import (
	"fmt"

	"sync"

	"github.com/uvite/u8/js/modules"

	"github.com/c9s/bbgo/pkg/types"
	"github.com/uvite/u8/tart/floats"

	. "github.com/uvite/u8/tart"
	"reflect"
)

// RootModule is the global module instance that will create module
// instances for each VU.
type RootModule struct{}

// Ta represents an instance of the timers module.
type Ta struct {
	vu modules.VU

	timerStopCounter uint32
	timerStopsLock   sync.Mutex
	timerStops       map[uint32]chan struct{}
}

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Ta{}
)

// New returns a pointer to a new RootModule instance.
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Ta{
		vu:         vu,
		timerStops: make(map[uint32]chan struct{}),
	}
}

// Exports returns the exports of the k6 module.
func (c *Ta) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"change":     c.Change,
			"series":     NewSeries,
			"slice":      floats.NewSlice,
			"crossover":  c.CrossOver,
			"crossunder": c.CrossUnder,

			//"alma": c.Alma,
			"hma":  c.Hma,
			"jma":  c.Jma,
			"dwma": c.DWma,

			"sma":      c.Sma,
			"atr":      c.Atr,
			"ema":      c.Ema,
			"rsi":      c.Rsi,
			"wma":      c.Wma,
			"willR":    c.WillR,
			"tr":       c.TRange,
			"stochRsi": c.StochRsi,
			"dev":      c.Dev,
			"stdDev":   c.StdDev,
			"Roc":      c.Roc,
			"obv":      c.Obv,
			"natr":     c.Natr,
			"macd":     c.Macd,
			"kama":     c.Kama,
			"adx":      c.Adx,
			"diff":     c.Diff,
			"ppo":      c.Ppo,
			"dema":     c.Dema,
			"cci":      c.Cci,
			"boll":     c.BBands,
			"aroon":    c.Aroon,
		},
	}
}

func noop() error { return nil }

func (ta Ta) CrossOver(s floats.Slice, t floats.Slice) bool {

	return s.Index(0)-t.Index(0) > 0 && s.Index(1)-t.Index(1) < 0

}
func (ta Ta) CrossUnder(s floats.Slice, t floats.Slice) bool {
	return s.Index(0)-t.Index(0) < 0 && s.Index(1)-t.Index(1) > 0

}
func (ta Ta) Change(args ...any) any {
	var res Series
	var len int = 1
	for _, arg := range args {
		//fmt.Println(arg)
		switch arg.(type) {
		case Series:
			res = arg.(Series)
		case []float64:
			res = NewSeries()
			for _, key := range arg.([]float64) {
				res.Push(key)
			}
			//fmt.Println(res.Tail(5))
		case int:
			len = arg.(int)
		case int64:
			len = int(arg.(int64))

		default:
			fmt.Println(arg, "error")
		}
	}
	if res.Length() > 0 {
		//fmt.Println(len, "leng")
		//fmt.Println(reflect.TypeOf(res.Index(0)).String(), "typeof")
		if reflect.TypeOf(res.Index(0)).String() == "float64" {
			//fmt.Println(res.Index(0).(float64))
			diff := res.Index(0).(float64) - res.Index(len).(float64)
			return diff
		}
		if reflect.TypeOf(res.Index(0)).String() == "string" {
			diff := res.Index(len).(string) != res.Index(0).(string)
			return diff
		}

	}
	return ""
}

// alma
//func (ta Ta) Alma(in floats.Slice, n int64, offset float64, sigma int) floats.Slice {
//
//	//out := make(floats.Slice, len(in))
//
//	alma := inc.ALMA{
//		IntervalWindow: types.IntervalWindow{Window: int(n)},
//		Offset:         offset,
//		Sigma:          sigma,
//	}
//	//alma.CalculateAndUpdate(tt.kLines)
//	//s := NewSma(n)
//	fmt.Println(alma.Length(), "alma.length")
//	for _, v := range in {
//		alma.Update(v)
//	}
//
//	return alma.Values
//}

// alma
func (ta Ta) Hma(in floats.Slice, n int64) floats.Slice {

	//out := make(floats.Slice, len(in))

	//hma := indicator.HULL{
	//	IntervalWindow: types.IntervalWindow{Window: int(n)},
	//}
	//
	//for _, v := range in {
	//	hma.Update(v)
	//}
	//
	//return hma.Values

	out := make([]float64, len(in))

	///k := 2.0 / float64(n+1)
	t := NewHma(n)
	for i, v := range in {
		out[i] = t.Update(v)
	}

	return out

}

func (ta Ta) Jma(in floats.Slice, n int64, phase float64, power float64) floats.Slice {

	//out := make(floats.Slice, len(in))

	jma := JMA{
		IntervalWindow: types.IntervalWindow{Window: int(n)},
		Phase:          phase,
		Power:          power,
	}
	//inc.Ma = &indi.JMA{IntervalWindow: types.IntervalWindow{Window: inc.Window}, Phase: inc.Phase, Power: inc.Power}

	//
	for _, v := range in {
		jma.Update(v)
	}
	//
	return floats.Slice(jma.Values)

}

// 双重wma
func (ta Ta) DWma(in floats.Slice, n int64) floats.Slice {
	//ta.wma(ta.wma(_src, _length), _length)

	out := ta.Wma(in, n)

	return ta.Wma(out, n)
}

// {"Wma","WillR","Var","UltOsc","Trix","Trima","TRange","StochRsi","StdDev","Roc","Obv",
// "Natr","Macd","Kama","Dx","Diff","Dev","ppo"}
//func (ta Ta) Sma(in floats.Slice, n int64) floats.Slice {
//
//	alma := indicator.ALMA{
//		IntervalWindow: types.IntervalWindow{Window: 5},
//		Offset:         0.9,
//		Sigma:          6,
//	}
//	alma.CalculateAndUpdate(tt.kLines)
//	return out
//}

// todo 以上为bbgo 补充的指标，以后要统一下

func (ta Ta) Sma(in floats.Slice, n int64) floats.Slice {

	out := make(floats.Slice, len(in))

	s := NewSma(n)
	for i, v := range in {
		out[i] = s.Update(v)
	}

	return out
}
func (ta Ta) Atr(h, l, c floats.Slice, n int64) floats.Slice {

	out := make(floats.Slice, len(c))

	a := NewAtr(n)
	for i := 0; i < len(c); i++ {
		//fmt.Println(h[i], l[i], c[i])
		out[i] = a.Update(h[i], l[i], c[i])
	}

	return out
}
func (ta Ta) Ema(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	k := 2.0 / float64(n+1)
	e := NewEma(n, k)
	for i, v := range in {
		out[i] = e.Update(v)
	}

	return out
}
func (ta Ta) Rsi(in floats.Slice, n int64) floats.Slice {

	out := make(floats.Slice, len(in))

	r := NewRsi(n)
	for i, v := range in {
		out[i] = r.Update(v)
	}

	return out
}
func (ta Ta) Wma(in floats.Slice, n int64) floats.Slice {

	out := make(floats.Slice, len(in))

	w := NewWma(n)
	for i, v := range in {
		out[i] = w.Update(v)
	}

	return out
}
func (ta Ta) WillR(h, l, c floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(c))

	w := NewWillR(n)
	for i := 0; i < len(c); i++ {
		out[i] = w.Update(h[i], l[i], c[i])
	}

	return out
}
func (ta Ta) Var(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	s := NewVar(n)
	for i, v := range in {
		out[i] = s.Update(v)
	}

	return out
}

func (ta Ta) UltOsc(h, l, c floats.Slice, n1, n2, n3 int64) floats.Slice {
	out := make(floats.Slice, len(c))

	u := NewUltOsc(n1, n2, n3)
	for i := 0; i < len(c); i++ {
		out[i] = u.Update(h[i], l[i], c[i])
	}

	return out
}
func (ta Ta) Trix(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	t := NewTrix(n)
	for i, v := range in {
		out[i] = t.Update(v)
	}

	return out
}
func (ta Ta) Trima(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	t := NewTrima(n)
	for i, v := range in {
		out[i] = t.Update(v)
	}

	return out
}
func (ta Ta) TRange(h, l, c floats.Slice) floats.Slice {
	out := make(floats.Slice, len(c))

	t := NewTRange()
	for i := 0; i < len(c); i++ {
		out[i] = t.Update(h[i], l[i], c[i])
	}

	return out
}

func (ta Ta) StochRsi(in floats.Slice, n, kN int64, dt MaType, dN int64) (floats.Slice, floats.Slice) {
	k := make(floats.Slice, len(in))
	d := make(floats.Slice, len(in))

	s := NewStochRsi(n, kN, dt, dN)
	for i, v := range in {
		k[i], d[i] = s.Update(v)
	}

	return k, d
}

func (ta Ta) StdDev(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	s := NewStdDev(n)
	for i, v := range in {
		out[i] = s.Update(v)
	}

	return out
}
func (ta Ta) Roc(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	r := NewRoc(n)
	for i, v := range in {
		out[i] = r.Update(v)
	}

	return out
}

func (ta Ta) Obv(c, v floats.Slice) floats.Slice {
	out := make(floats.Slice, len(c))

	o := NewObv()
	for i := 0; i < len(c); i++ {
		out[i] = o.Update(c[i], v[i])
	}

	return out
}
func (ta Ta) Natr(h, l, c floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(c))

	a := NewNatr(n)
	for i := 0; i < len(c); i++ {
		out[i] = a.Update(h[i], l[i], c[i])
	}

	return out
}

func (ta Ta) Macd(in floats.Slice, fastN, slowN, signalN int64) (floats.Slice, floats.Slice, floats.Slice) {
	macd := make(floats.Slice, len(in))
	signal := make(floats.Slice, len(in))
	hist := make(floats.Slice, len(in))

	m := NewMacd(fastN, slowN, signalN)
	for i, v := range in {
		macd[i], signal[i], hist[i] = m.Update(v)
	}

	return macd, signal, hist
}
func (ta Ta) Kama(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	k := NewKama(n)
	for i, v := range in {
		out[i] = k.Update(v)
	}

	return out
}
func (ta Ta) Dx(h, l, c floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(c))

	d := NewDx(n)
	for i := 0; i < len(c); i++ {
		out[i] = d.Update(h[i], l[i], c[i])
	}

	return out
}

func (ta Ta) Diff(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	d := NewDiff(n)
	for i, v := range in {
		out[i] = d.Update(v)
	}

	return out
}

func (ta Ta) Dev(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	d := NewDev(n)
	for i, v := range in {
		out[i] = d.Update(v)
	}

	return out
}
func (ta Ta) Ppo(in floats.Slice, t MaType, fastN, slowN int64) floats.Slice {
	out := make(floats.Slice, len(in))

	p := NewPpo(t, fastN, slowN)
	for i, v := range in {
		out[i] = p.Update(v)
	}

	return out
}

func (ta Ta) Dema(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	k := 2.0 / float64(n+1)
	d := NewDema(n, k)
	for i, v := range in {
		out[i] = d.Update(v)
	}

	return out
}
func (ta Ta) Cmo(in floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(in))

	c := NewCmo(n)
	for i, v := range in {
		out[i] = c.Update(v)
	}

	return out
}

func (ta Ta) Cci(h, l, c floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(h))

	d := NewCci(n)
	for i := 0; i < len(h); i++ {
		out[i] = d.Update(h[i], l[i], c[i])
	}

	return out
}
func ChangeMaType(ma string) MaType {
	var maType MaType
	switch ma {
	case "SMA":
		maType = SMA
	case "EMA":
		maType = EMA
	}
	return maType
	//case SMA:
	//	mu = NewSma(n)
	//	case EMA:
	//	mu = NewEma(n, k)
	//	case WMA:
	//	mu = NewWma(n)
	//	case DEMA:
	//	mu = NewDema(n, k)
	//	case TEMA:
	//	mu = NewTema(n, k)
	//	case TRIMA:
	//	mu = NewTrima(n)
	//	case KAMA:
	//	mu = NewKama(n)
	//
}
func (ta Ta) BBands(ma string, in floats.Slice, n int64, upNStdDev, dnNStdDev float64) map[string]floats.Slice {

	//fmt.Println(t, "tama")
	t := ChangeMaType(ma)
	//fmt.Println(t, ma)
	m := make(floats.Slice, len(in))
	u := make(floats.Slice, len(in))
	l := make(floats.Slice, len(in))

	b := NewBBands(t, n, upNStdDev, dnNStdDev)
	for i, v := range in {
		u[i], m[i], l[i] = b.Update(v)
	}

	ret := make(map[string]floats.Slice)
	ret["u"] = u
	ret["m"] = m
	ret["l"] = l
	return ret
}
func (ta Ta) AroonOsc(h, l floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(h))

	a := NewAroonOsc(n)
	for i := 0; i < len(h); i++ {
		out[i] = a.Update(h[i], l[i])
	}

	return out
}
func (ta Ta) Aroon(h, l floats.Slice, n int64) (floats.Slice, floats.Slice) {
	dn := make(floats.Slice, len(h))
	up := make(floats.Slice, len(h))

	a := NewAroon(n)
	for i := 0; i < len(h); i++ {
		dn[i], up[i] = a.Update(h[i], l[i])
	}

	return dn, up
}
func (ta Ta) Apo(t MaType, in floats.Slice, fastN, slowN int64) floats.Slice {
	out := make(floats.Slice, len(in))

	a := NewApo(t, fastN, slowN)
	for i, v := range in {
		out[i] = a.Update(v)
	}

	return out
}

func (ta Ta) Adx(h, l, c floats.Slice, n int64) floats.Slice {
	out := make(floats.Slice, len(c))

	a := NewAdx(n)
	for i := 0; i < len(c); i++ {
		out[i] = a.Update(h[i], l[i], c[i])
	}

	return out
}
