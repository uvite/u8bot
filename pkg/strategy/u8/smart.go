package u8

import (
	"context"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/c9s/bbgo/pkg/util"
)

func (s *Strategy) smartCancel(ctx context.Context, pricef, atr float64, syscounter int) (int, error) {
	nonTraded := s.GeneralOrderExecutor.ActiveMakerOrders().Orders()
	if len(nonTraded) > 0 {
		if len(nonTraded) > 1 {
			log.Errorf("should only have one order to cancel, got %d", len(nonTraded))
		}
		toCancel := false

		for _, order := range nonTraded {
			if order.Status != types.OrderStatusNew && order.Status != types.OrderStatusPartiallyFilled {
				continue
			}
			s.pendingLock.Lock()
			counter := s.orderPendingCounter[order.OrderID]
			s.pendingLock.Unlock()

			log.Warnf("%v | counter: %d, system: %d", order, counter, syscounter)
			if syscounter-counter > s.PendingMinInterval {
				toCancel = true
			} else if order.Side == types.SideTypeBuy {
				// 75% of the probability
				if order.Price.Float64()+atr*2 <= pricef {
					toCancel = true
				}
			} else if order.Side == types.SideTypeSell {
				// 75% of the probability
				if order.Price.Float64()-atr*2 >= pricef {
					toCancel = true
				}
			} else {
				panic("not supported side for the order")
			}
		}
		if toCancel {
			err := s.GeneralOrderExecutor.FastCancel(ctx)
			// TODO: clean orderPendingCounter on cancel/trade
			for _, order := range nonTraded {
				s.pendingLock.Lock()
				counter := s.orderPendingCounter[order.OrderID]
				delete(s.orderPendingCounter, order.OrderID)
				s.pendingLock.Unlock()
				if order.Side == types.SideTypeSell {
					if s.maxCounterSellCanceled < counter {
						s.maxCounterSellCanceled = counter
					}
				} else {
					if s.maxCounterBuyCanceled < counter {
						s.maxCounterBuyCanceled = counter
					}
				}
			}
			log.Warnf("cancel all %v", err)
			return 0, err
		}
	}
	return len(nonTraded), nil
}

func (s *Strategy) trailingCheck(price float64, direction string) bool {
	if s.highestPrice > 0 && s.highestPrice < price {
		s.highestPrice = price
	}
	if s.lowestPrice > 0 && s.lowestPrice > price {
		s.lowestPrice = price
	}
	isShort := direction == "short"
	if isShort && s.sellPrice == 0 || !isShort && s.buyPrice == 0 {
		return false
	}
	for i := len(s.TrailingCallbackRate) - 1; i >= 0; i-- {
		trailingCallbackRate := s.TrailingCallbackRate[i]
		trailingActivationRatio := s.TrailingActivationRatio[i]
		if isShort {
			if (s.sellPrice-s.lowestPrice)/s.lowestPrice > trailingActivationRatio {
				return (price-s.lowestPrice)/s.lowestPrice > trailingCallbackRate
			}
		} else {
			if (s.highestPrice-s.buyPrice)/s.buyPrice > trailingActivationRatio {
				return (s.highestPrice-price)/s.buyPrice > trailingCallbackRate
			}
		}
	}
	return false
}

func (s *Strategy) initTickerFunctions(ctx context.Context) {
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
			} else {
				s.midPrice = bestBid
			}
			s.lock.Unlock()

			// we removed realtime stoploss and trailingStop.

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
