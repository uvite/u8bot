package server

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/pkg/accounting/pnl"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/service"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log2 "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) botPnls(c *gin.Context) {

	if s.Environ.TradeService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database is not configured"})
		return
	}

	exchange := c.Query("exchange")
	symbol := c.Query("symbol")
	gidStr := c.DefaultQuery("gid", "0")
	lastGID, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Error("last gid parse error")
		c.Status(http.StatusBadRequest)
		return
	}
	sinceOpt := c.Query("since")
	//limit := c.Query("limit")
	//includeTransfer := c.Query("includeTransfer")

	session, ok := s.Environ.Session(exchange)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", exchange)})
		return
	}

	since := time.Now().AddDate(-1, 0, 0)

	if sinceOpt != "" {
		lt, err := types.ParseLooseFormatTime(sinceOpt)
		logrus.WithError(err).Error("sinceOpt error")

		since = lt.Time()
	}

	//until := time.Now()
	//exchange := session.Exchange
	//market, _ := session.Market(symbol)
	//
	var trades []types.Trade

	trades, err = s.Environ.TradeService.Query(service.QueryTradesOptions{
		Exchange: types.ExchangeName(exchange),
		Symbol:   symbol,
		LastGID:  lastGID,
		Since:    &since,
		Ordering: "DESC",
	})

	trades = types.SortTradesAscending(trades)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickers, err := session.Exchange.QueryTickers(ctx, symbol)
	logrus.WithError(err).Error("sync error")

	currentTick, ok := tickers[symbol]
	if !ok {
		logrus.WithError(err).Error("no ticker data for current price")

	}

	market, ok := session.Market(symbol)
	if !ok {
		logrus.WithError(err).Error("market not found: %s, %s", symbol, session.Exchange.Name())

	}
	//tradingFeeCurrency := session.Exchange.PlatformFeeCurrency()
	currentPrice := currentTick.Last
	//calculator := &pnl.AverageCostCalculator{
	//	TradingFeeCurrency: tradingFeeCurrency,
	//	Market:             market,
	//}
	//
	//report := calculator.Calculate(symbol, trades, currentPrice)
	//report.Print()

	stats := types.NewTradeStats(symbol)

	// copy trades, so that we can truncate it.
	var bidVolume = fixedpoint.Zero
	var askVolume = fixedpoint.Zero
	var feeUSD = fixedpoint.Zero
	var grossProfit = fixedpoint.Zero
	var grossLoss = fixedpoint.Zero
	var pnlReport = pnl.AverageCostPnLReport{}
	var position = types.NewPositionFromMarket(market)

	makerFeeRate := 0.075 * 0.01

	position.SetFeeRate(types.ExchangeFee{
		// binance vip 0 uses 0.075%
		MakerFeeRate: fixedpoint.NewFromFloat(makerFeeRate),
		TakerFeeRate: fixedpoint.NewFromFloat(0.075 * 0.01),
	})

	if len(trades) == 0 {
		pnlReport = pnl.AverageCostPnLReport{
			Symbol:     symbol,
			Market:     market,
			LastPrice:  currentPrice,
			NumTrades:  0,
			Position:   position,
			BuyVolume:  bidVolume,
			SellVolume: askVolume,
			FeeInUSD:   feeUSD,
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"report": pnlReport, "stats": stats}})

		//c.JSON(http.StatusOK, gin.H{"report": pnlReport, "stats": stats})
		return
	}

	var currencyFees = map[string]fixedpoint.Value{}

	// TODO: configure the exchange fee rate here later
	// position.SetExchangeFeeRate()
	var totalProfit fixedpoint.Value
	var totalNetProfit fixedpoint.Value

	var tradeIDs = map[uint64]types.Trade{}

	for _, trade := range trades {
		if _, exists := tradeIDs[trade.ID]; exists {
			log2.Warnf("duplicated trade: %+v", trade)
			continue
		}

		if trade.Symbol != symbol {
			continue
		}

		profit, netProfit, madeProfit := position.AddTrade(trade)
		if madeProfit {
			totalProfit = totalProfit.Add(profit)
			totalNetProfit = totalNetProfit.Add(netProfit)
		}

		if profit.Sign() > 0 {
			grossProfit = grossProfit.Add(profit)
		} else if profit.Sign() < 0 {
			grossLoss = grossLoss.Add(profit)
		}

		if trade.IsBuyer {
			bidVolume = bidVolume.Add(trade.Quantity)
		} else {
			askVolume = askVolume.Add(trade.Quantity)
		}

		if _, ok := currencyFees[trade.FeeCurrency]; !ok {
			currencyFees[trade.FeeCurrency] = trade.Fee
		} else {
			currencyFees[trade.FeeCurrency] = currencyFees[trade.FeeCurrency].Add(trade.Fee)
		}

		tradeIDs[trade.ID] = trade
		fmt.Println(trade.OrderID, profit)
		pp := &types.Profit{OrderID: trade.OrderID, Profit: profit, Symbol: symbol}
		stats.Add(pp)

	}

	unrealizedProfit := currentPrice.Sub(position.AverageCost).
		Mul(position.GetBase())

	pnlReport = pnl.AverageCostPnLReport{
		Symbol:    symbol,
		Market:    market,
		LastPrice: currentPrice,
		NumTrades: len(trades),
		StartTime: time.Time(trades[0].Time),
		Position:  position,

		BuyVolume:  bidVolume,
		SellVolume: askVolume,

		BaseAssetPosition: position.GetBase(),
		Profit:            totalProfit,
		NetProfit:         totalNetProfit,
		UnrealizedProfit:  unrealizedProfit,

		GrossProfit: grossProfit,
		GrossLoss:   grossLoss,

		AverageCost:  position.AverageCost,
		FeeInUSD:     totalProfit.Sub(totalNetProfit),
		CurrencyFees: currencyFees,
	}
	fmt.Println(pnlReport, stats.String())
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"report": pnlReport, "stats": stats}})
}
