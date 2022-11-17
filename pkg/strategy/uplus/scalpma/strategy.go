package scalpma

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	internal "github.com/c9s/bbgo/u8/nats"
	"math"
	"os"
	"sync"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/strategy/uplus/indi"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/c9s/bbgo/pkg/util"
	"github.com/sirupsen/logrus"
)

const ID = "scalpma"

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type SourceFunc func(*types.KLine) fixedpoint.Value

type ScalpPrice struct {
	top2     float64 `persistence:"top2_price"`
	top3     float64 `persistence:"top3_price"`
	bott2    float64 `persistence:"bott2_price"`
	bott3    float64 `persistence:"bott3_price"`
	buyPrice float64 `persistence:"stop_long_price"`
	tpLong   float64 `persistence:"tp_long_price"`
	stopLong float64 `persistence:"stop_long_price"`

	sellPrice float64 `persistence:"stop_long_price"`
	tpShort   float64 `persistence:"tp_short_price"`
	stopShort float64 `persistence:"stop_long_price"`
}

type Strategy struct {
	Symbol string `json:"symbol"`

	bbgo.OpenPositionOptions
	bbgo.StrategyController
	bbgo.SourceSelector
	types.Market
	Session  *bbgo.ExchangeSession
	Leverage fixedpoint.Value `json:"leverage"`

	Interval  types.Interval   `json:"interval"`
	Stoploss  fixedpoint.Value `json:"stoploss" modifiable:"true"`
	WindowATR int              `json:"windowATR"`
	WindowJMA int              `json:"windowJMA"`
	WindowRSX int              `json:"windowRSX"`
	JmaPhase  float64          `json:"jmaPhase"`
	JmaPower  float64          `json:"jmaPower"`
	Sigma     float64          `json:"sigma"`
	AtrSigma  float64          `json:"atrSigma"`
	Side      string           `json:"side"`

	PendingMinutes int `json:"pendingMinutes" modifiable:"true"`

	// whether to draw graph or not by the end of backtest
	DrawGraph          bool   `json:"drawGraph"`
	GraphIndicatorPath string `json:"graphIndicatorPath"`
	GraphPNLPath       string `json:"graphPNLPath"`
	GraphCumPNLPath    string `json:"graphCumPNLPath"`

	*bbgo.Environment
	*bbgo.GeneralOrderExecutor
	*types.Position    `persistence:"position"`
	*types.ProfitStats `persistence:"profit_stats"`
	*types.TradeStats  `persistence:"trade_stats"`

	atr *indicator.ATR

	change    *indi.Slice
	jma       *indi.JMA
	rsx       *indi.RSX
	PriceOpen *types.Queue

	orders *types.Queue

	getLastPrice func() fixedpoint.Value

	// for smart cancel
	orderPendingCounter map[uint64]int
	startTime           time.Time
	minutesCounter      int

	// for position
	buyPrice float64 `persistence:"buy_price"`

	buyTime      time.Time
	sellPrice    float64 `persistence:"sell_price"`
	highestPrice float64 `persistence:"highest_price"`
	lowestPrice  float64 `persistence:"lowest_price"`

	TrailingCallbackRate    []float64          `json:"trailingCallbackRate" modifiable:"true"`
	TrailingActivationRatio []float64          `json:"trailingActivationRatio" modifiable:"true"`
	ExitMethods             bbgo.ExitMethodSet `json:"exits"`

	midPrice fixedpoint.Value
	lock     sync.RWMutex `ignore:"true"`

	ScalpPrice ScalpPrice

	//止损时间
	StopTime time.Time
	Stop     bool
	Pubsub   internal.PubSub
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) InstanceID() string {
	return fmt.Sprintf("%s:%s:%v", ID, s.Symbol, bbgo.IsBackTesting)
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {

	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: types.Interval1m,
	})
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: s.Interval})

	if !bbgo.IsBackTesting {
		session.Subscribe(types.BookTickerChannel, s.Symbol, types.SubscribeOptions{})
	}
	s.ExitMethods.SetAndSubscribe(session, s)
}

func (s *Strategy) CurrentPosition() *types.Position {
	return s.Position
}

func (s *Strategy) ClosePosition(ctx context.Context, percentage fixedpoint.Value, tag string) error {

	if tag == "stop" {
		s.Stop = true
		timeDur := time.Duration(1*60*1) * time.Minute //四小时

		s.StopTime = time.Now().Add(timeDur) //四小时
		bbgo.Notify("止损 。。。。。%v", s.StopTime)

	}
	//fmt.Println("时间差：", s.buyTime.Add(5*time.Minute), time.Now(), time.Now().After(s.buyTime.Add(5*time.Minute)))
	//fmt.Println("仓位：", s.Position.IsLong(), s.Position.IsShort())
	//
	//if !time.Now().After(s.buyTime.Add(5 * time.Minute)) {
	//	fmt.Println("1分钟内不能平仓")
	//	return nil
	//}
	//fmt.Println("可以平了")
	order := s.Position.NewMarketCloseOrder(percentage)
	if order == nil {
		return nil
	}
	order.Tag = tag
	order.TimeInForce = ""
	balances := s.GeneralOrderExecutor.Session().GetAccount().Balances()
	baseBalance := balances[s.Market.BaseCurrency].Available
	price := s.getLastPrice()
	if !s.Session.Futures {
		if order.Side == types.SideTypeBuy {
			quoteAmount := balances[s.Market.QuoteCurrency].Available.Div(price)
			if order.Quantity.Compare(quoteAmount) > 0 {
				order.Quantity = quoteAmount
			}
		} else if order.Side == types.SideTypeSell && order.Quantity.Compare(baseBalance) > 0 {
			order.Quantity = baseBalance
		}
	}

	order.ReduceOnly = true
	order.MarginSideEffect = types.SideEffectTypeAutoRepay
	//if price.Float64() > s.buyPrice {
	//	order.Type = types.OrderTypeTakeProfitLimit
	//
	//}
	//order.Type = types.OrderTypeLimit
	order.Price = price
	//fmt.Println("--------------------", order.Type)
	if s.Position.IsLong() {
		s.Pubsub.Publish(internal.OrderCloseLong, internal.OrderPayload{
			Symbol:   s.Symbol,
			Side:     types.SideTypeBuy,
			Quantity: s.Quantity,
			Price:    price,
		})
	}
	if s.Position.IsShort() {
		s.Pubsub.Publish(internal.OrderCloseLong, internal.OrderPayload{
			Symbol:   s.Symbol,
			Side:     types.SideTypeSell,
			Quantity: s.Quantity,
			Price:    price,
		})
	}

	for {

		if s.Market.IsDustQuantity(order.Quantity, price) {
			return nil
		}

		_, err := s.GeneralOrderExecutor.SubmitOrders(ctx, *order)
		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}

}

func (s *Strategy) initIndicators(store *bbgo.MarketDataStore) error {
	s.change = &indi.Slice{}

	s.atr = &indicator.ATR{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowATR}}

	s.jma = &indi.JMA{IntervalWindow: types.IntervalWindow{Window: s.WindowJMA}, Phase: s.JmaPhase, Power: s.JmaPower}
	s.rsx = &indi.RSX{IntervalWindow: types.IntervalWindow{Window: s.WindowRSX}}
	klines, ok := store.KLinesOfInterval(s.Interval)

	//fmt.Println("klines", klines)
	klineLength := len(*klines)

	if !ok || klineLength == 0 {
		return errors.New("klines not exists")
	}
	tmpK := (*klines)[klineLength-1]
	s.startTime = tmpK.StartTime.Time().Add(tmpK.Interval.Duration())

	for _, kline := range *klines {
		s.AddKline(kline)

	}
	return nil
}
func (s *Strategy) AddKline(kline types.KLine) {
	source := s.GetSource(&kline).Float64()
	s.atr.PushK(kline)
	s.jma.Update(source)
	s.rsx.Update(source)
	s.PriceOpen.Update(kline.Open.Float64())

	//fmt.Println("atr,jma,rsx", source, s.atr.Last(), s.jma.Last(), s.rsx.Last())
}

func (s *Strategy) smartCancel(ctx context.Context, pricef float64) int {
	nonTraded := s.GeneralOrderExecutor.ActiveMakerOrders().Orders()
	if len(nonTraded) > 0 {
		left := 0
		for _, order := range nonTraded {
			if order.Status != types.OrderStatusNew && order.Status != types.OrderStatusPartiallyFilled {
				continue
			}
			log.Warnf("%v | counter: %d, system: %d", order, s.orderPendingCounter[order.OrderID], s.minutesCounter)
			toCancel := false
			if s.minutesCounter-s.orderPendingCounter[order.OrderID] >= s.PendingMinutes {
				toCancel = true
			} else if order.Side == types.SideTypeBuy {
				if order.Price.Float64()+s.atr.Last()*2 <= pricef {
					toCancel = true
				}
			} else if order.Side == types.SideTypeSell {
				// 75% of the probability
				if order.Price.Float64()-s.atr.Last()*2 >= pricef {
					toCancel = true
				}
			} else {
				panic("not supported side for the order")
			}
			if toCancel {
				err := s.GeneralOrderExecutor.GracefulCancel(ctx, order)
				if err == nil {
					delete(s.orderPendingCounter, order.OrderID)
				} else {
					log.WithError(err).Errorf("failed to cancel %v", order.OrderID)
				}
				log.Warnf("cancel %v", order.OrderID)
			} else {
				left += 1
			}
		}
		return left
	}
	return len(nonTraded)
}

func (s *Strategy) trailingCheck(price float64, direction string) bool {
	if s.highestPrice > 0 && s.highestPrice < price {
		s.highestPrice = price
	}
	if s.lowestPrice > 0 && s.lowestPrice > price {
		s.lowestPrice = price
	}
	isShort := direction == "short"
	for i := len(s.TrailingCallbackRate) - 1; i >= 0; i-- {
		trailingCallbackRate := s.TrailingCallbackRate[i]
		trailingActivationRatio := s.TrailingActivationRatio[i]
		if isShort {
			if (s.sellPrice-s.lowestPrice)/s.lowestPrice > trailingActivationRatio {
				return (price-s.lowestPrice)/s.lowestPrice > trailingCallbackRate
			}
		} else {
			if (s.highestPrice-s.buyPrice)/s.buyPrice > trailingActivationRatio {
				return (s.highestPrice-price)/price > trailingCallbackRate
			}
		}
	}
	return false
}

func (s *Strategy) initTickerFunctions() {
	if s.IsBackTesting() {
		s.getLastPrice = func() fixedpoint.Value {
			lastPrice, ok := s.Session.LastPrice(s.Symbol)
			if !ok {
				log.Error("cannot get lastprice")
			}
			return lastPrice
		}
	} else {
		s.Session.MarketDataStream.OnBookTickerUpdate(func(ticker types.BookTicker) {
			bestBid := ticker.Buy
			bestAsk := ticker.Sell
			if !util.TryLock(&s.lock) {
				return
			}
			if !bestAsk.IsZero() && !bestBid.IsZero() {
				s.midPrice = bestAsk.Add(bestBid).Div(Two)
			} else if !bestAsk.IsZero() {
				s.midPrice = bestAsk
			} else if !bestBid.IsZero() {
				s.midPrice = bestBid
			}
			s.lock.Unlock()
		})
		s.getLastPrice = func() (lastPrice fixedpoint.Value) {
			var ok bool
			s.lock.RLock()
			defer s.lock.RUnlock()
			if s.midPrice.IsZero() {
				lastPrice, ok = s.Session.LastPrice(s.Symbol)
				if !ok {
					log.Error("cannot get lastprice")
					return lastPrice
				}
			} else {
				lastPrice = s.midPrice
			}
			return lastPrice
		}
	}
}

func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	instanceID := s.InstanceID()
	if s.Position == nil {
		s.Position = types.NewPositionFromMarket(s.Market)
	}
	if s.ProfitStats == nil {
		s.ProfitStats = types.NewProfitStats(s.Market)
	}
	if s.TradeStats == nil {
		s.TradeStats = types.NewTradeStats(s.Symbol)
	}
	s.ScalpPrice = ScalpPrice{
		buyPrice: 0.0,
		tpLong:   0.0,
		stopLong: 0.0,

		sellPrice: 0.0,
		tpShort:   0.0,
		stopShort: 0.0,
	}
	s.Stop = false

	s.Pubsub = internal.NewPubSub()
	//defer s.Pubsub.Close()

	// StrategyController
	s.Status = types.StrategyStatusRunning
	s.OnSuspend(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
	})
	s.OnEmergencyStop(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
		_ = s.ClosePosition(ctx, fixedpoint.One, "stop")
	})
	s.GeneralOrderExecutor = bbgo.NewGeneralOrderExecutor(session, s.Symbol, ID, instanceID, s.Position)
	s.GeneralOrderExecutor.BindEnvironment(s.Environment)
	s.GeneralOrderExecutor.BindProfitStats(s.ProfitStats)
	s.GeneralOrderExecutor.BindTradeStats(s.TradeStats)
	s.GeneralOrderExecutor.TradeCollector().OnPositionUpdate(func(p *types.Position) {
		//fmt.Println("", p)
		bbgo.Sync(ctx, s)
	})
	s.GeneralOrderExecutor.Bind()

	s.orderPendingCounter = make(map[uint64]int)
	s.minutesCounter = 0

	for _, method := range s.ExitMethods {
		method.Bind(session, s.GeneralOrderExecutor)
	}

	s.PriceOpen = types.NewQueue(300)
	s.StopTime = time.Now()
	profit := floats.Slice{1., 1.}
	price, _ := s.Session.LastPrice(s.Symbol)
	initAsset := s.CalcAssetValue(price).Float64()
	cumProfit := floats.Slice{initAsset, initAsset}
	modify := func(p float64) float64 {
		return p
	}

	s.GeneralOrderExecutor.TradeCollector().OnTrade(func(trade types.Trade, _profit, _netProfit fixedpoint.Value) {

		price := trade.Price.Float64()
		if s.buyPrice > 0 {
			profit.Update(modify(price / s.buyPrice))
			cumProfit.Update(s.CalcAssetValue(trade.Price).Float64())
		} else if s.sellPrice > 0 {
			profit.Update(modify(s.sellPrice / price))
			cumProfit.Update(s.CalcAssetValue(trade.Price).Float64())
		}
		if s.Position.IsDust(trade.Price) {
			s.buyPrice = 0
			s.sellPrice = 0
			s.highestPrice = 0
			s.lowestPrice = 0
		} else if s.Position.IsLong() {
			s.buyPrice = price
			s.buyTime = time.Now()
			s.sellPrice = 0
			s.highestPrice = s.buyPrice
			s.lowestPrice = 0
		} else {
			s.sellPrice = price
			s.buyPrice = 0
			s.highestPrice = 0
			s.lowestPrice = s.sellPrice
		}
	})
	s.initTickerFunctions()

	startTime := s.Environment.StartTime()
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1d, startTime))
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1w, startTime))

	s.initOutputCommands()

	// event trigger order: s.Interval => Interval1m
	store, ok := session.MarketDataStore(s.Symbol)
	if !ok {
		panic("cannot get 1m history")
	}
	if err := s.initIndicators(store); err != nil {
		log.WithError(err).Errorf("initIndicator failed")
		return nil
	}

	s.change.Push(HOLD)
	//if s.Side == "short" {
	//	s.PostionCheck(ctx)
	//}
	//if s.Side == "long" {
	//	s.PostionCheck1(ctx)
	//}

	store.OnKLineClosed(func(kline types.KLine) {
		s.minutesCounter = int(kline.StartTime.Time().Add(kline.Interval.Duration()).Sub(s.startTime).Minutes())
		if s.Stop {
			if s.StopTime.Before(time.Now()) {
				s.Stop = false
				bbgo.Notify("静默期 。。。。。")

			}
		}
		if kline.Interval == s.Interval && !s.Stop {
			s.klineHandler(ctx, kline)
			s.klineHandler1m(ctx, kline)
		}

	})

	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {

		var buffer bytes.Buffer

		s.Print(&buffer, true, true)

		fmt.Fprintln(&buffer, "--- NonProfitable Dates ---")
		for _, daypnl := range s.TradeStats.IntervalProfits[types.Interval1d].GetNonProfitableIntervals() {
			fmt.Fprintf(&buffer, "%s\n", daypnl)
		}
		fmt.Fprintln(&buffer, s.TradeStats.BriefString())

		os.Stdout.Write(buffer.Bytes())

		//if s.GenerateGraph {
		//	s.Draw(s.frameKLine.StartTime, &profit, &cumProfit)
		//}
		wg.Done()
	})

	return nil
}

func (s *Strategy) CalcAssetValue(price fixedpoint.Value) fixedpoint.Value {
	balances := s.Session.GetAccount().Balances()
	return balances[s.Market.BaseCurrency].Total().Mul(price).Add(balances[s.Market.QuoteCurrency].Total())
}

func (s *Strategy) klineHandler1m(ctx context.Context, kline types.KLine) {
	if s.Status != types.StrategyStatusRunning {
		return
	}
	price := s.getLastPrice()
	pricef := price.Float64()
	lowf := math.Min(kline.Low.Float64(), pricef)
	highf := math.Max(kline.High.Float64(), pricef)

	if s.Position.IsLong() {
		tag := ""
		if lowf < s.ScalpPrice.stopLong {
			tag = "stop"
		}
		if highf > s.ScalpPrice.tpLong {
			tag = "limit"
		}
		exitCondition := (lowf < s.ScalpPrice.stopLong || highf > s.ScalpPrice.tpLong) || (s.change.Change() && s.change.Last() == SELL)
		if exitCondition {
			fmt.Println("平多")
			err := s.ClosePosition(ctx, fixedpoint.One, tag)
			if err != nil {
				fmt.Println("----", err)
			}
		}
	}

	if s.Position.IsShort() {
		tag := ""
		if highf > s.ScalpPrice.stopShort {
			tag = "stop"
		}
		if lowf < s.ScalpPrice.tpShort {
			tag = "limit"
		}
		exitCondition := (lowf < s.ScalpPrice.tpShort || highf > s.ScalpPrice.stopShort) || (s.change.Change() && s.change.Last() == BUY)
		if exitCondition {
			fmt.Println("平空")
			err := s.ClosePosition(ctx, fixedpoint.One, tag)
			if err != nil {
				fmt.Println("----", err)
			}
		}
	}
	//if s.Position.IsShort() {
	//	exitCondition := (lowf > s.ScalpPrice.tpShort || highf > s.ScalpPrice.stopShort) || (s.change.Change() && s.change.Last() == BUY)
	//	if exitCondition {
	//		fmt.Println("平空")
	//		_ = s.ClosePosition(ctx, fixedpoint.One)
	//	}
	//}

}
func (s *Strategy) checkPrice() {
	atr := s.atr.Last() * s.AtrSigma

	top3 := s.jma.Last() + atr*4.236
	top2 := s.jma.Last() + atr*2.618
	//top1 := s.ma.Last() + atr*1.618
	//bott1 := s.ma.Last() - atr*1.618
	bott2 := s.jma.Last() - atr*2.618
	bott3 := s.jma.Last() - atr*4.236

	k := s.Sigma

	fmt.Println("s.Sigma", s.Sigma, s.AtrSigma)

	tpLong := top2 * (1 - k)
	stopLong := bott3 * (1 - k*0.5)

	tpShort := bott2 * (1 + k)
	stopShort := top3 * (1 + k*0.5)
	s.ScalpPrice = ScalpPrice{
		top2:     top2,
		top3:     top3,
		bott2:    bott2,
		bott3:    bott3,
		buyPrice: 0.0,
		tpLong:   tpLong,
		stopLong: stopLong,

		sellPrice: 0.0,
		tpShort:   tpShort,
		stopShort: stopShort,
	}
	bbgo.Notify("long sma:%4.f,bott2:%.4f,limit:%.4f,stop:%.4f,atr:%.4f,atrx:%.4f", s.jma.Last(), bott2, tpLong, stopLong, s.atr.Last(), atr)
	bbgo.Notify("short sma:%4.f,top2:%.4f,limit:%.4f,stop:%.4f,atr:%.4f,,atrx:%.4f", s.jma.Last(), top2, tpShort, stopShort, s.atr.Last(), atr)
	payload := map[string]string{
		"long":  fmt.Sprint("long sma:%4.f,bott2:%.4f,limit:%.4f,stop:%.4f,atr:%.4f,atrx:%.4f", s.jma.Last(), bott2, tpLong, stopLong, s.atr.Last(), atr),
		"short": fmt.Sprint("short sma:%4.f,top2:%.4f,limit:%.4f,stop:%.4f,atr:%.4f,,atrx:%.4f", s.jma.Last(), top2, tpShort, stopShort, s.atr.Last(), atr),
	}

	s.Pubsub.Publish(internal.MessageShow, payload)
}

func (s *Strategy) klineHandler(ctx context.Context, kline types.KLine) {

	source := s.GetSource(&kline)
	sourcef := source.Float64()
	s.AddKline(kline)
	if s.Status != types.StrategyStatusRunning {
		return
	}

	//stoploss := s.Stoploss.Float64()
	price := s.getLastPrice()
	pricef := price.Float64()

	balances := s.GeneralOrderExecutor.Session().GetAccount().Balances()
	bbgo.Notify("source: %.4f, price: %.4f lowest: %.4f highest: %.4f sp %.4f bp %.4f", sourcef, pricef, s.lowestPrice, s.highestPrice, s.sellPrice, s.buyPrice)
	bbgo.Notify("balances: [Total] %v %s [Base] %s(%v %s) [Quote] %s",
		s.CalcAssetValue(price),
		s.Market.QuoteCurrency,
		balances[s.Market.BaseCurrency].String(),
		balances[s.Market.BaseCurrency].Total().Mul(price),
		s.Market.QuoteCurrency,
		balances[s.Market.QuoteCurrency].String(),
	)
	s.checkPrice()

	up := kline.Low.Float64() < s.ScalpPrice.bott2 && kline.Low.Float64() > s.ScalpPrice.bott3 && s.PriceOpen.Index(0) > s.PriceOpen.Index(1)
	down := kline.High.Float64() > s.ScalpPrice.top2 && kline.High.Float64() < s.ScalpPrice.top3 && s.PriceOpen.Index(0) < s.PriceOpen.Index(1)

	sig := s.change.Last()

	if up {

		sig = BUY
	}
	if down {

		sig = SELL
	}

	s.change.Push(sig)

	switch s.Side {
	case "long":
		s.ProcessLong(ctx, kline)
	case "short":
		s.ProcessShort(ctx, kline)
	case "both":

	default:
		log.Panicf("undefined side: %s", s.Side)
	}

	//changed := s.change.Change()

}

// 多单处理
func (s *Strategy) ProcessLong(ctx context.Context, kline types.KLine) {

	source := s.GetSource(&kline)
	price := s.getLastPrice()

	longCondition := s.change.Last() == BUY && s.rsx.Last() < 35
	if s.Position.IsLong() {
		bbgo.Notify("%s position is already opened, skip", s.Symbol)
		return
	}

	if longCondition {
		if err := s.GeneralOrderExecutor.GracefulCancel(ctx); err != nil {
			log.WithError(err).Errorf("cannot cancel orders")
			return
		}
		if source.Compare(price) > 0 {
			source = price
		}
		createdOrders, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
			Symbol:   s.Symbol,
			Side:     types.SideTypeBuy,
			Quantity: s.Quantity,
			Type:     types.OrderTypeMarket,
			Tag:      "short: sell in",
		})

		s.Pubsub.Publish(internal.OrderOpenLong, internal.OrderPayload{
			Symbol:   s.Symbol,
			Side:     types.SideTypeBuy,
			Quantity: s.Quantity,
			Price:    price,
		})

		if err != nil {
			if _, ok := err.(types.ZeroAssetError); ok {
				return
			}
			log.WithError(err).Errorf("cannot place buy order: %v %v", source, kline)
			return
		}
		if createdOrders != nil {
			s.orderPendingCounter[createdOrders[0].OrderID] = s.minutesCounter
		}
		return
	}
}

// 空单处理
func (s *Strategy) ProcessShort(ctx context.Context, kline types.KLine) {

	source := s.GetSource(&kline)
	price := s.getLastPrice()

	shortCondition := s.change.Last() == SELL && s.rsx.Last() > 65
	if s.Position.IsShort() {
		bbgo.Notify("%s position is already opened, skip", s.Symbol)
		return
	}

	if shortCondition {
		bbgo.Notify("开空")
		if err := s.GeneralOrderExecutor.GracefulCancel(ctx); err != nil {
			log.WithError(err).Errorf("cannot cancel orders")
			return
		}
		if source.Compare(price) < 0 {
			source = price
		}

		createdOrders, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
			Symbol:   s.Symbol,
			Side:     types.SideTypeSell,
			Quantity: s.Quantity,
			Type:     types.OrderTypeMarket,
			Tag:      "short: sell in",
		})

		s.Pubsub.Publish(internal.OrderOpenShort, internal.OrderPayload{
			Symbol:   s.Symbol,
			Side:     types.SideTypeSell,
			Quantity: s.Quantity,
			Price:    price,
		})
		//opt := s.OpenPositionOptions
		//opt.LimitOrder = false
		//opt.Short = true
		//opt.Price = source
		//opt.Leverage = s.Leverage
		//opt.Tags = []string{"short"}
		//log.Infof("source in short %v %v", source, price)
		//createdOrders, err := s.GeneralOrderExecutor.OpenPosition(ctx, opt)
		if err != nil {
			if _, ok := err.(types.ZeroAssetError); ok {
				return
			}
			log.WithError(err).Errorf("cannot place sell order: %v %v", source, kline)
			return
		}
		if createdOrders != nil {
			s.orderPendingCounter[createdOrders[0].OrderID] = s.minutesCounter
		}
		return
	}
}
