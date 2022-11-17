package jmacross

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/data/tsv"
	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/interact"
	"github.com/c9s/bbgo/pkg/strategy/uplus/indi"
	"github.com/c9s/bbgo/pkg/types"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const ID = "jmacross"

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type Strategy struct {
	//bbgo.OpenPositionOptions

	bbgo.SourceSelector

	Session  *bbgo.ExchangeSession
	Leverage fixedpoint.Value `json:"leverage"`
	*bbgo.Environment

	Symbol string `json:"symbol"`
	Market types.Market

	types.IntervalWindow
	//bbgo.OpenPositionOptions

	// persistence fields
	Position    *types.Position    `persistence:"position"`
	ProfitStats *types.ProfitStats `persistence:"profit_stats"`
	TradeStats  *types.TradeStats  `persistence:"trade_stats"`

	ExitMethods bbgo.ExitMethodSet `json:"exits"`

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor

	bbgo.QuantityOrAmount

	// StrategyController
	bbgo.StrategyController

	grid   *GRID
	Phase  float64 `json:"phase"`
	Power  float64 `json:"power"`
	dmi    *indicator.DMI
	atr    *indicator.ATR
	hma    *indi.HMA
	jma    *indi.JMA
	change *indi.Slice

	Side string `json:"side"`

	PriceLine *types.Queue `json:"-"`

	getLastPrice func() fixedpoint.Value

	WindowATR int `json:"windowATR"`
	WindowDMI int `json:"windowDMI"`
	WindowHMA int `json:"windowHMA"`
	WindowJMA int `json:"windowJMA"`
	Smoothing int `json:"smoothing"`

	buyTime time.Time

	midPrice fixedpoint.Value
	lock     sync.RWMutex `ignore:"true"`

	AccountValueCalculator *bbgo.AccountValueCalculator

	// whether to draw graph or not by the end of backtest
	DrawGraph       bool   `json:"drawGraph"`
	GraphPNLPath    string `json:"graphPNLPath"`
	GraphCumPNLPath string `json:"graphCumPNLPath"`

	// for position
	buyPrice     float64 `persistence:"buy_price"`
	sellPrice    float64 `persistence:"sell_price"`
	highestPrice float64 `persistence:"highest_price"`
	lowestPrice  float64 `persistence:"lowest_price"`

	// Accumulated profit report
	AccumulatedProfitReport *AccumulatedProfitReport `json:"accumulatedProfitReport"`
}

// AccumulatedProfitReport For accumulated profit report output
type AccumulatedProfitReport struct {
	// AccumulatedProfitMAWindow Accumulated profit SMA window, in number of trades
	AccumulatedProfitMAWindow int `json:"accumulatedProfitMAWindow"`

	// IntervalWindow interval window, in days
	IntervalWindow int `json:"intervalWindow"`

	// NumberOfInterval How many intervals to output to TSV
	NumberOfInterval int `json:"NumberOfInterval"`

	// TsvReportPath The path to output report to
	TsvReportPath string `json:"tsvReportPath"`

	// AccumulatedDailyProfitWindow The window to sum up the daily profit, in days
	AccumulatedDailyProfitWindow int `json:"accumulatedDailyProfitWindow"`

	// Accumulated profit
	accumulatedProfit         fixedpoint.Value
	accumulatedProfitPerDay   floats.Slice
	previousAccumulatedProfit fixedpoint.Value

	// Accumulated profit MA
	accumulatedProfitMA       *indicator.SMA
	accumulatedProfitMAPerDay floats.Slice

	// Daily profit
	dailyProfit floats.Slice

	// Accumulated fee
	accumulatedFee       fixedpoint.Value
	accumulatedFeePerDay floats.Slice

	// Win ratio
	winRatioPerDay floats.Slice

	// Profit factor
	profitFactorPerDay floats.Slice

	// Trade number
	dailyTrades               floats.Slice
	accumulatedTrades         int
	previousAccumulatedTrades int
}

func (r *AccumulatedProfitReport) Initialize() {
	if r.AccumulatedProfitMAWindow <= 0 {
		r.AccumulatedProfitMAWindow = 60
	}
	if r.IntervalWindow <= 0 {
		r.IntervalWindow = 7
	}
	if r.AccumulatedDailyProfitWindow <= 0 {
		r.AccumulatedDailyProfitWindow = 7
	}
	if r.NumberOfInterval <= 0 {
		r.NumberOfInterval = 1
	}
	r.accumulatedProfitMA = &indicator.SMA{IntervalWindow: types.IntervalWindow{Interval: types.Interval1d, Window: r.AccumulatedProfitMAWindow}}
}

func (r *AccumulatedProfitReport) RecordProfit(profit fixedpoint.Value) {
	r.accumulatedProfit = r.accumulatedProfit.Add(profit)
}

func (r *AccumulatedProfitReport) RecordTrade(fee fixedpoint.Value) {
	r.accumulatedFee = r.accumulatedFee.Add(fee)
	r.accumulatedTrades += 1
}

func (r *AccumulatedProfitReport) DailyUpdate(tradeStats *types.TradeStats) {
	// Daily profit
	r.dailyProfit.Update(r.accumulatedProfit.Sub(r.previousAccumulatedProfit).Float64())
	r.previousAccumulatedProfit = r.accumulatedProfit

	// Accumulated profit
	r.accumulatedProfitPerDay.Update(r.accumulatedProfit.Float64())

	// Accumulated profit MA
	r.accumulatedProfitMA.Update(r.accumulatedProfit.Float64())
	r.accumulatedProfitMAPerDay.Update(r.accumulatedProfitMA.Last())

	// Accumulated Fee
	r.accumulatedFeePerDay.Update(r.accumulatedFee.Float64())

	// Win ratio
	r.winRatioPerDay.Update(tradeStats.WinningRatio.Float64())

	// Profit factor
	r.profitFactorPerDay.Update(tradeStats.ProfitFactor.Float64())

	// Daily trades
	r.dailyTrades.Update(float64(r.accumulatedTrades - r.previousAccumulatedTrades))
	r.previousAccumulatedTrades = r.accumulatedTrades
}

// Output Accumulated profit report to a TSV file
func (r *AccumulatedProfitReport) Output(symbol string) {
	if r.TsvReportPath != "" {
		tsvwiter, err := tsv.AppendWriterFile(r.TsvReportPath)
		if err != nil {
			panic(err)
		}
		defer tsvwiter.Close()
		// Output symbol, total acc. profit, acc. profit 60MA, interval acc. profit, fee, win rate, profit factor
		_ = tsvwiter.Write([]string{"#", "Symbol", "accumulatedProfit", "accumulatedProfitMA", fmt.Sprintf("%dd profit", r.AccumulatedDailyProfitWindow), "accumulatedFee", "winRatio", "profitFactor", "60D trades"})
		for i := 0; i <= r.NumberOfInterval-1; i++ {
			accumulatedProfit := r.accumulatedProfitPerDay.Index(r.IntervalWindow * i)
			accumulatedProfitStr := fmt.Sprintf("%f", accumulatedProfit)
			accumulatedProfitMA := r.accumulatedProfitMAPerDay.Index(r.IntervalWindow * i)
			accumulatedProfitMAStr := fmt.Sprintf("%f", accumulatedProfitMA)
			intervalAccumulatedProfit := r.dailyProfit.Tail(r.AccumulatedDailyProfitWindow+r.IntervalWindow*i).Sum() - r.dailyProfit.Tail(r.IntervalWindow*i).Sum()
			intervalAccumulatedProfitStr := fmt.Sprintf("%f", intervalAccumulatedProfit)
			accumulatedFee := fmt.Sprintf("%f", r.accumulatedFeePerDay.Index(r.IntervalWindow*i))
			winRatio := fmt.Sprintf("%f", r.winRatioPerDay.Index(r.IntervalWindow*i))
			profitFactor := fmt.Sprintf("%f", r.profitFactorPerDay.Index(r.IntervalWindow*i))
			trades := r.dailyTrades.Tail(60+r.IntervalWindow*i).Sum() - r.dailyTrades.Tail(r.IntervalWindow*i).Sum()
			tradesStr := fmt.Sprintf("%f", trades)

			_ = tsvwiter.Write([]string{fmt.Sprintf("%d", i+1), symbol, accumulatedProfitStr, accumulatedProfitMAStr, intervalAccumulatedProfitStr, accumulatedFee, winRatio, profitFactor, tradesStr})
		}
	}
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: s.Interval})

	if !bbgo.IsBackTesting {
		session.Subscribe(types.MarketTradeChannel, s.Symbol, types.SubscribeOptions{})
	}

	s.ExitMethods.SetAndSubscribe(session, s)
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) InstanceID() string {
	return fmt.Sprintf("%s:%s", ID, s.Symbol)
}

func (s *Strategy) CalcAssetValue(price fixedpoint.Value) fixedpoint.Value {
	balances := s.session.GetAccount().Balances()
	return balances[s.Market.BaseCurrency].Total().Mul(price).Add(balances[s.Market.QuoteCurrency].Total())
}

func (s *Strategy) initIndicators(store *bbgo.MarketDataStore) error {
	s.change = &indi.Slice{}

	s.atr = &indicator.ATR{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowATR}}
	s.hma = &indi.HMA{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowHMA}}
	s.jma = &indi.JMA{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowJMA}, Phase: s.Phase, Power: s.Power}
	//fmt.Println(s.WindowHMA, s.WindowJMA, s.Power, s.Phase)
	//s.dmi = &indicator.DMI{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowDMI}, ADXSmoothing: s.WindowDMI}

	klines, ok := store.KLinesOfInterval(s.Interval)

	//fmt.Println("klines", klines)
	klineLength := len(*klines)

	if !ok || klineLength == 0 {
		return errors.New("klines not exists")
	}

	for _, kline := range *klines {
		s.AddKline(kline)

	}
	return nil
}
func (s *Strategy) AddKline(kline types.KLine) {

	s.atr.PushK(kline)
	s.hma.Update(kline.Close.Float64())
	s.jma.Update(kline.Close.Float64())
	s.PriceLine.Update(kline.Close.Float64())

	//fmt.Println("hma,jma", kline.EndTime, kline.Close, s.hma.Last(), s.jma.Last())

}
func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	var instanceID = s.InstanceID()

	if s.Position == nil {
		s.Position = types.NewPositionFromMarket(s.Market)
	}

	if s.ProfitStats == nil {
		s.ProfitStats = types.NewProfitStats(s.Market)
	}

	if s.TradeStats == nil {
		s.TradeStats = types.NewTradeStats(s.Symbol)
	}

	// StrategyController
	s.Status = types.StrategyStatusRunning

	s.OnSuspend(func() {
		// Cancel active orders
		_ = s.orderExecutor.GracefulCancel(ctx)
	})

	s.OnEmergencyStop(func() {
		// Cancel active orders
		_ = s.orderExecutor.GracefulCancel(ctx)
		// Close 100% position
		//_ = s.ClosePosition(ctx, fixedpoint.One)
	})

	s.session = session

	// Set fee rate
	if s.session.MakerFeeRate.Sign() > 0 || s.session.TakerFeeRate.Sign() > 0 {
		s.Position.SetExchangeFeeRate(s.session.ExchangeName, types.ExchangeFee{
			MakerFeeRate: s.session.MakerFeeRate,
			TakerFeeRate: s.session.TakerFeeRate,
		})
	}

	s.orderExecutor = bbgo.NewGeneralOrderExecutor(session, s.Symbol, ID, instanceID, s.Position)
	s.orderExecutor.BindEnvironment(s.Environment)
	s.orderExecutor.BindProfitStats(s.ProfitStats)
	s.orderExecutor.BindTradeStats(s.TradeStats)

	// AccountValueCalculator
	s.AccountValueCalculator = bbgo.NewAccountValueCalculator(s.session, s.Market.QuoteCurrency)

	// Accumulated profit report
	if bbgo.IsBackTesting {
		if s.AccumulatedProfitReport == nil {
			s.AccumulatedProfitReport = &AccumulatedProfitReport{}
		}
		s.AccumulatedProfitReport.Initialize()
		s.orderExecutor.TradeCollector().OnProfit(func(trade types.Trade, profit *types.Profit) {
			if profit == nil {
				return
			}

			s.AccumulatedProfitReport.RecordProfit(profit.Profit)
		})
		// s.orderExecutor.TradeCollector().OnTrade(func(trade types.Trade, profit fixedpoint.Value, netProfit fixedpoint.Value) {
		// 	s.AccumulatedProfitReport.RecordTrade(trade.Fee)
		// })
		session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, types.Interval1d, func(kline types.KLine) {
			s.AccumulatedProfitReport.DailyUpdate(s.TradeStats)
		}))
	}

	// For drawing
	profitSlice := floats.Slice{1., 1.}
	price, _ := session.LastPrice(s.Symbol)
	initAsset := s.CalcAssetValue(price).Float64()
	cumProfitSlice := floats.Slice{initAsset, initAsset}

	s.orderExecutor.TradeCollector().OnTrade(func(trade types.Trade, profit fixedpoint.Value, netProfit fixedpoint.Value) {
		if bbgo.IsBackTesting {
			s.AccumulatedProfitReport.RecordTrade(trade.Fee)
		}

		// For drawing/charting
		price := trade.Price.Float64()
		if s.buyPrice > 0 {
			profitSlice.Update(price / s.buyPrice)
			cumProfitSlice.Update(s.CalcAssetValue(trade.Price).Float64())
		} else if s.sellPrice > 0 {
			profitSlice.Update(s.sellPrice / price)
			cumProfitSlice.Update(s.CalcAssetValue(trade.Price).Float64())
		}
		if s.Position.IsDust(trade.Price) {
			s.buyPrice = 0
			s.sellPrice = 0
			s.highestPrice = 0
			s.lowestPrice = 0
		} else if s.Position.IsLong() {
			s.buyPrice = price
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

	s.orderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		bbgo.Sync(ctx, s)
	})

	s.orderExecutor.Bind()

	for _, method := range s.ExitMethods {
		method.Bind(session, s.orderExecutor)
	}
	s.PriceLine = types.NewQueue(300)
	kLineStore, ok := s.session.MarketDataStore(s.Symbol)
	if !ok {
		panic("cannot get 1m history")
	}
	if err := s.initIndicators(kLineStore); err != nil {
		log.WithError(err).Errorf("initIndicator failed")
		return nil
	}
	s.initTickerFunctions()

	startTime := s.Environment.StartTime()

	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1d, startTime))
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1w, startTime))

	//s.PostionCheck(ctx)
	s.session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {
		s.AddKline(kline)

		//log.Infof("Shark Score: %f, Current Price: %f", s.grid.Last(), kline.Close.Float64())

		up := s.hma.CrossOver(s.jma)
		down := s.hma.CrossUnder(s.jma)
		fmt.Println("up,down", up.Last(), down.Last())

		sig := s.change.Last()

		if up.Last() {

			sig = BUY
		}
		if down.Last() {

			sig = SELL
		}
		s.change.Push(sig)

		//switch s.Side {
		//case "long":
		//	s.ProcessLong(ctx, kline)
		//case "short":
		//	s.ProcessShort(ctx, kline)
		//case "both":
		//	s.ProcessLong(ctx, kline)
		//	s.ProcessShort(ctx, kline)
		//
		//default:
		//	log.Panicf("undefined side: %s", s.Side)
		//}

		fmt.Printf("changed: %t,changlast:%s \n", s.change.Change(), s.change.Last())

	}))

	bbgo.RegisterCommand("/config", "Show latest config", func(reply interact.Reply) {
		var buffer bytes.Buffer
		s.Print(&buffer, false)
		reply.Message(buffer.String())
	})
	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {

		var buffer bytes.Buffer

		//s.Print(&buffer, true, true)

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

/*
*
平空
检测多单是否有仓位
*/
func (s *Strategy) ProcessLong(ctx context.Context, kline types.KLine) {

	source := s.GetSource(&kline)
	price := s.getLastPrice()
	longCondition := s.change.Change() && s.change.Last() == BUY
	exitCondition := s.change.Change() && s.change.Last() == SELL

	if s.Position.IsShort() || exitCondition {
		_ = s.orderExecutor.GracefulCancel(ctx)
		s.orderExecutor.ClosePosition(ctx, fixedpoint.One, "close long position")
	}
	if s.Position.IsLong() {
		bbgo.Notify("%s position is already opened, skip", s.Symbol)
		fmt.Println("%s position is already opened, skip", s.Symbol)
		return
	}

	if longCondition {
		if err := s.orderExecutor.GracefulCancel(ctx); err != nil {
			log.WithError(err).Errorf("cannot cancel orders")
			return
		}
		if source.Compare(price) > 0 {
			source = price
		}
		_, err := s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
			Symbol:   s.Symbol,
			Side:     types.SideTypeBuy,
			Quantity: s.Quantity,
			Type:     types.OrderTypeMarket,
			Tag:      "long",
		})
		if err != nil {
			log.Errorln(err)
		}
		return
	}
}

// 空单处理
func (s *Strategy) ProcessShort(ctx context.Context, kline types.KLine) {

	source := s.GetSource(&kline)
	price := s.getLastPrice()

	shortCondition := s.change.Change() && s.change.Last() == SELL
	exitCondition := s.change.Change() && s.change.Last() == BUY

	if s.Position.IsLong() || exitCondition {
		_ = s.orderExecutor.GracefulCancel(ctx)
		s.orderExecutor.ClosePosition(ctx, fixedpoint.One, "close long position")
	}
	if s.Position.IsShort() {
		bbgo.Notify("%s position is already opened, skip", s.Symbol)
		fmt.Println("%s short position is already opened, skip", s.Symbol)

		return
	}

	if shortCondition {
		bbgo.Notify("开空")
		if err := s.orderExecutor.GracefulCancel(ctx); err != nil {
			log.WithError(err).Errorf("cannot cancel orders")
			return
		}
		if source.Compare(price) < 0 {
			source = price
		}

		createdOrders, err := s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
			Symbol:   s.Symbol,
			Side:     types.SideTypeSell,
			Quantity: s.Quantity,
			Type:     types.OrderTypeMarket,
			Tag:      "short: sell in",
		})

		if err != nil {
			if _, ok := err.(types.ZeroAssetError); ok {
				return
			}
			log.WithError(err).Errorf("cannot place sell order: %v %v", source, kline)
			return
		}
		if createdOrders != nil {

		}
		return
	}
}
