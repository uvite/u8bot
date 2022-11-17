package dmima

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/c9s/bbgo/pkg/util"
)

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

func (s *Strategy) ClosePosition(ctx context.Context, percentage fixedpoint.Value) error {

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
