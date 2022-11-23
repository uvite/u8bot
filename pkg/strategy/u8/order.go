package u8

import (
	"context"
	"errors"
	"fmt"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"math"
)

func (s *Strategy) ClosePosition(ctx context.Context, percentage fixedpoint.Value) error {
	order := s.p.NewMarketCloseOrder(percentage)
	if order == nil {
		return nil
	}
	order.Tag = "close"
	order.TimeInForce = ""

	order.MarginSideEffect = types.SideEffectTypeAutoRepay
	for i := 0; i < closeOrderRetryLimit; i++ {
		price := s.getLastPrice()
		balances := s.GeneralOrderExecutor.Session().GetAccount().Balances()
		baseBalance := balances[s.Market.BaseCurrency].Available
		if order.Side == types.SideTypeBuy {
			quoteAmount := balances[s.Market.QuoteCurrency].Available.Div(price)
			if order.Quantity.Compare(quoteAmount) > 0 {
				order.Quantity = quoteAmount
			}
		} else if order.Side == types.SideTypeSell && order.Quantity.Compare(baseBalance) > 0 {
			order.Quantity = baseBalance
		}
		order.ReduceOnly = true
		if s.Market.IsDustQuantity(order.Quantity, price) {
			return nil
		}
		fmt.Println(order)
		_, err := s.GeneralOrderExecutor.SubmitOrders(ctx, *order)
		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}
	return errors.New("exceed retry limit")
}

func (s *Strategy) klineHandlerMin(ctx context.Context, kline types.KLine, counter int) {

	if s.Status != types.StrategyStatusRunning {
		return
	}
	// for doing the trailing stoploss during backtesting
	atr := s.atr.Last()
	price := s.getLastPrice()
	pricef := price.Float64()

	lowf := math.Min(kline.Low.Float64(), pricef)
	highf := math.Max(kline.High.Float64(), pricef)
	s.positionLock.Lock()
	if s.lowestPrice > 0 && lowf < s.lowestPrice {
		s.lowestPrice = lowf
	}
	if s.highestPrice > 0 && highf > s.highestPrice {
		s.highestPrice = highf
	}
	s.positionLock.Unlock()

	numPending := 0
	var err error
	if numPending, err = s.smartCancel(ctx, pricef, atr, counter); err != nil {
		log.WithError(err).Errorf("cannot cancel orders")
		return
	}
	if numPending > 0 {
		return
	}

	if s.NoTrailingStopLoss {
		return
	}
	fmt.Println("check stop ")
	exitCondition := s.CheckStopLoss() || s.trailingCheck(highf, "short") || s.trailingCheck(lowf, "long")
	if exitCondition {
		_ = s.ClosePosition(ctx, fixedpoint.One)
	}
}
