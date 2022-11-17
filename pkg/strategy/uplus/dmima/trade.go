package dmima

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"time"
)

func (s *Strategy) PostionCheck(ctx context.Context) {
	s.session.MarketDataStream.OnKLine(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {

		shortCondition := true

		if s.Position.IsShort() {
			exitCondition := s.PriceLine.CrossOver(s.hma).Last() // || (s.change.Change() && s.change.Last() == BUY)

			if exitCondition {
				fmt.Println("平空")
				_ = s.ClosePosition(ctx, fixedpoint.One)
			}
		}
		if shortCondition { // && ((previousRegime < zeroThreshold && previousRegime > -zeroThreshold) || s.grid.Index(1) > 0)
			fmt.Println("开空")
			if s.Position.IsLong() {
				_ = s.orderExecutor.GracefulCancel(ctx)
				s.orderExecutor.ClosePosition(ctx, fixedpoint.One, "close long position")
			}
			_, err := s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   s.Symbol,
				Side:     types.SideTypeSell,
				Quantity: s.Quantity,
				Type:     types.OrderTypeMarket,
				Tag:      "grid short: sell in",
			})
			//if err == nil {
			//	_, err = s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
			//		Symbol:   s.Symbol,
			//		Side:     types.SideTypeBuy,
			//		Quantity: s.Quantity,
			//		//Price:    fixedpoint.NewFromFloat(s.grid.Lows.Tail(100).Min()),
			//		Type: types.OrderTypeMarket,
			//		Tag:  "grid short: buy back",
			//	})
			//}
			if err != nil {
				log.Errorln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}))
}

func (s *Strategy) ClosePosition1(ctx context.Context, percentage fixedpoint.Value) error {

	//fmt.Println("时间差：", s.buyTime.Add(5*time.Minute), time.Now(), time.Now().After(s.buyTime.Add(5*time.Minute)))
	//fmt.Println("仓位：", s.Position.IsLong(), s.Position.IsShort())
	//
	//if !time.Now().After(s.buyTime.Add(5 * time.Minute)) {
	//	fmt.Println("1分钟内不能平仓")
	//	return nil
	//}
	//fmt.Println("可以平了")

	price := s.getLastPrice()
	order := s.Position.NewMarketCloseOrder(percentage)
	if order == nil {
		return nil
	}
	order.Tag = "close"
	order.TimeInForce = ""
	balances := s.orderExecutor.Session().GetAccount().Balances()
	baseBalance := balances[s.Market.BaseCurrency].Available

	order.ReduceOnly = true
	order.MarginSideEffect = types.SideEffectTypeAutoRepay
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
	//order.Type = types.OrderTypeLimit
	order.Price = price
	fmt.Println("--------------------", order.Type)

	for {

		if s.Market.IsDustQuantity(order.Quantity, price) {
			return nil
		}

		_, err := s.orderExecutor.SubmitOrders(ctx, *order)
		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}

}
