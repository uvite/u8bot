package irr

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/data/tsv"
	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/sirupsen/logrus"
)

const ID = "irr"

var one = fixedpoint.One
var zero = fixedpoint.Zero

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type Strategy struct {
	Environment *bbgo.Environment
	Symbol      string `json:"symbol"`
	Market      types.Market

	types.IntervalWindow

	// persistence fields
	Position    *types.Position    `persistence:"position"`
	ProfitStats *types.ProfitStats `persistence:"profit_stats"`
	TradeStats  *types.TradeStats  `persistence:"trade_stats"`

	activeOrders *bbgo.ActiveOrderBook

	ExitMethods bbgo.ExitMethodSet `json:"exits"`

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor

	bbgo.QuantityOrAmount

	Interval int `json:"hftInterval"`

	// realtime book ticker to submit order
	obBuyPrice  uint64
	obSellPrice uint64
	// for getting close price
	currentTradePrice uint64
	// for negative return rate
	openPrice  float64
	closePrice float64

	stopC chan struct{}

	// StrategyController
	bbgo.StrategyController

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
	if !bbgo.IsBackTesting {
		session.Subscribe(types.AggTradeChannel, s.Symbol, types.SubscribeOptions{})
		session.Subscribe(types.BookTickerChannel, s.Symbol, types.SubscribeOptions{})
	}
	//session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: s.Interval})
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) InstanceID() string {
	return fmt.Sprintf("%s:%s", ID, s.Symbol)
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
		_ = s.orderExecutor.ClosePosition(ctx, fixedpoint.One)
	})

	// initial required information
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
		session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, types.Interval1d, func(kline types.KLine) {
			s.AccumulatedProfitReport.DailyUpdate(s.TradeStats)
		}))
	}

	// For drawing
	profitSlice := floats.Slice{1., 1.}
	price, _ := session.LastPrice(s.Symbol)
	initAsset := s.CalcAssetValue(price).Float64()
	cumProfitSlice := floats.Slice{initAsset, initAsset}
	profitDollarSlice := floats.Slice{0, 0}
	cumProfitDollarSlice := floats.Slice{0, 0}

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
		profitDollarSlice.Update(profit.Float64())
		cumProfitDollarSlice.Update(profitDollarSlice.Sum())
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

	s.InitDrawCommands(&profitSlice, &cumProfitSlice, &cumProfitDollarSlice)

	s.orderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		bbgo.Sync(ctx, s)
	})
	s.orderExecutor.Bind()
	s.activeOrders = bbgo.NewActiveOrderBook(s.Symbol)

	atomic.SwapUint64(&s.currentTradePrice, 0.)
	s.closePrice = 0.
	s.openPrice = 0.
	klinDirections := types.NewQueue(100)
	started := false
	boxOpenPrice := 0.
	boxClosePrice := 0.
	boxCounter := 0

	if !bbgo.IsBackTesting {

		s.session.MarketDataStream.OnBookTickerUpdate(func(bt types.BookTicker) {
			// quote order book price
			newBid := uint64(bt.Buy.Float64())
			newAsk := uint64(bt.Sell.Float64())
			atomic.SwapUint64(&s.obBuyPrice, newBid)
			atomic.SwapUint64(&s.obSellPrice, newAsk)
		})

		s.session.MarketDataStream.OnAggTrade(func(trade types.Trade) {
			tradePrice := uint64(trade.Price.Float64())
			atomic.SwapUint64(&s.currentTradePrice, tradePrice)
		})

		closeTime := <-time.After(time.Duration(s.Interval-int(time.Now().UnixMilli())%s.Interval) * time.Millisecond)
		log.Infof("kline close timing synced @ %s", closeTime.Format("2006-01-02 15:04:05.000000"))
		go func() {
			intervalCloseTicker := time.NewTicker(time.Duration(s.Interval) * time.Millisecond)
			defer intervalCloseTicker.Stop()
			for {
				select {
				case <-intervalCloseTicker.C:
					log.Infof("kline close time @ %s", time.Now().Format("2006-01-02 15:04:05.000000"))

					s.orderExecutor.CancelOrders(context.Background())

					if s.currentTradePrice > 0 {
						s.closePrice = float64(s.currentTradePrice)
						log.Infof("Close Price: %f", s.closePrice)
						if s.closePrice > 0 && s.openPrice > 0 {
							direction := s.closePrice - s.openPrice
							klinDirections.Update(direction)
							regimeShift := klinDirections.Index(0)*klinDirections.Index(1) < 0
							if regimeShift && !started {
								boxOpenPrice = s.openPrice
								started = true
								boxCounter = 0
								log.Infof("box started at price: %f", boxOpenPrice)
							} else if regimeShift && started {
								boxClosePrice = s.closePrice
								started = false
								log.Infof("box ended at price: %f with time length: %d", boxClosePrice, boxCounter)
								// box ending, should re-balance position
								nirr := fixedpoint.NewFromFloat(((boxOpenPrice - boxClosePrice) / boxOpenPrice) / (float64(boxCounter) + 1))
								log.Infof("Alpha: %f", nirr.Float64())
								if nirr.Float64() < 0 {
									_, err := s.orderExecutor.SubmitOrders(context.Background(), types.SubmitOrder{
										Symbol:   s.Symbol,
										Side:     types.SideTypeSell,
										Quantity: s.Quantity,
										Type:     types.OrderTypeLimitMaker,
										Price:    fixedpoint.NewFromFloat(float64(s.obSellPrice)),
										Tag:      "irrSell",
									})
									if err != nil {
										log.WithError(err)
									}
								} else if nirr.Float64() > 0 {
									_, err := s.orderExecutor.SubmitOrders(context.Background(), types.SubmitOrder{
										Symbol:   s.Symbol,
										Side:     types.SideTypeBuy,
										Quantity: s.Quantity,
										Type:     types.OrderTypeLimitMaker,
										Price:    fixedpoint.NewFromFloat(float64(s.obBuyPrice)),
										Tag:      "irrBuy",
									})
									if err != nil {
										log.WithError(err)
									}
								}
							} else {
								boxCounter++
							}
						}
					}
				case <-s.stopC:
					log.Warnf("%s goroutine stopped, due to the stop signal", s.Symbol)
					return

				case <-ctx.Done():
					log.Warnf("%s goroutine stopped, due to the cancelled context", s.Symbol)
					return
				}
			}

		}()

		openTime := <-time.After(time.Duration(s.Interval-int(time.Now().UnixMilli())%s.Interval) * time.Millisecond)
		log.Infof("kline open timing synced @ %s", openTime.Format("2006-01-02 15:04:05.000000"))
		go func() {
			intervalOpenTicker := time.NewTicker(time.Duration(s.Interval) * time.Millisecond)
			defer intervalOpenTicker.Stop()
			for {
				select {
				case <-intervalOpenTicker.C:
					time.Sleep(10 * time.Millisecond)
					log.Infof("kline open time @ %s", time.Now().Format("2006-01-02 15:04:05.000000"))

					if s.currentTradePrice > 0 && s.closePrice > 0 {
						s.openPrice = float64(s.currentTradePrice)
						log.Infof("Open Price: %f", s.openPrice)
					}
				case <-s.stopC:
					log.Warnf("%s goroutine stopped, due to the stop signal", s.Symbol)
					return

				case <-ctx.Done():
					log.Warnf("%s goroutine stopped, due to the cancelled context", s.Symbol)
					return
				}
			}
		}()
	}

	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		// Output accumulated profit report
		if bbgo.IsBackTesting {
			defer s.AccumulatedProfitReport.Output(s.Symbol)

			if s.DrawGraph {
				if err := s.Draw(&profitSlice, &cumProfitSlice); err != nil {
					log.WithError(err).Errorf("cannot draw graph")
				}
			}
		} else {
			close(s.stopC)
		}
		_, _ = fmt.Fprintln(os.Stderr, s.TradeStats.String())
		_ = s.orderExecutor.GracefulCancel(ctx)
	})
	return nil
}

func (s *Strategy) CalcAssetValue(price fixedpoint.Value) fixedpoint.Value {
	balances := s.session.GetAccount().Balances()
	return balances[s.Market.BaseCurrency].Total().Mul(price).Add(balances[s.Market.QuoteCurrency].Total())
}
