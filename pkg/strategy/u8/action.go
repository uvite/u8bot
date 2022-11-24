package u8

import (
	"errors"
	"fmt"
	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

/*
strategy.entry(id, direction, qty, limit, stop, oca_name, oca_type, comment, alert_message)
开单
*/

type Options struct {
	qty     fixedpoint.Value `json:"qty"`
	limit   fixedpoint.Value `json:"limit"`
	stop    fixedpoint.Value `json:"stop"`
	comment string           `json:"comment"`
	tag     string           `json:"tag"`
}

/*
*
处理止赢止损
*/
func (s *Strategy) CheckLimitStop() {
	s.Session.MarketDataStream.OnKLine(func(kline types.KLine) {
		s.Price = kline.Close
		//fmt.Println(s.Price.Float64(), s.sellPrice, s.shortLimitStop.limit.Float64(), s.sellPrice-s.shortLimitStop.limit.Float64())
		//fmt.Println("多空:", s.Position.IsLong(), s.Position.IsShort())
		if s.Position.IsLong() && s.longLimitStop.limit > 0 {
			if s.Price.Float64() > s.buyPrice+s.longLimitStop.limit.Float64() {
				fmt.Println("止赢平多")
				s.Exit("long")
			}
		}
		if s.Position.IsShort() && s.shortLimitStop.limit > 0 {

			if s.Price.Float64() < s.sellPrice-s.shortLimitStop.limit.Float64() {
				fmt.Println("止赢平空")
				s.Exit("short")
			}
		}

	})
}

// 多单处理
func (s *Strategy) OpenOrder(side Side, options *Options) {
	price := s.getLastPrice()
	direct := types.SideTypeBuy
	if side == SideLong {

		if s.Position.IsLong() {
			bbgo.Notify("%s position is already opened, skip", s.Symbol)
			return
		}
		direct = types.SideTypeBuy
		s.longLimitStop = &LimtStop{
			limit: options.limit,
			stop:  options.stop,
		}
	} else if side == SideShort {

		if s.Position.IsShort() {
			bbgo.Notify("%s position is already opened, skip", s.Symbol)
			return
		}
		direct = types.SideTypeSell

		s.shortLimitStop = &LimtStop{
			limit: options.limit,
			stop:  options.stop,
		}
	}

	if err := s.GeneralOrderExecutor.GracefulCancel(s.ctx); err != nil {
		log.WithError(err).Errorf("cannot cancel orders")
		return
	}
	quantity := s.Quantity
	if options.qty > 0 {
		quantity = options.qty
	}

	fmt.Println("s.longLimitStop, s.shortLimitStop", s.longLimitStop, s.shortLimitStop)

	createdOrders, err := s.GeneralOrderExecutor.SubmitOrders(s.ctx, types.SubmitOrder{
		Symbol:   s.Symbol,
		Side:     direct,
		Quantity: quantity,
		Type:     types.OrderTypeMarket,
		Tag:      options.comment,
	})

	if err != nil {
		if _, ok := err.(types.ZeroAssetError); ok {
			return
		}
		log.WithError(err).Errorf("cannot place buy order: %v", price)
		return
	}
	if createdOrders != nil {
		s.orderPendingCounter[createdOrders[0].OrderID] = s.counter
	}
	return

}

//
//// 开空处理
//func (s *Strategy) OpenShort(options *Options) {
//	fmt.Println(options.qty, options.limit, options.comment)
//	return
//
//}

func (s *Strategy) Entry(id string, side Side, data map[string]interface{}) {
	//option := Options{}
	params := Keys(data)

	fmt.Println(params, data)
	option := &Options{}
	for _, k := range params {
		switch k {
		case "qty":
			option.qty, _ = fixedpoint.NewFromString(data["qty"].(string))

		case "limit":
			option.limit, _ = fixedpoint.NewFromString(data["limit"].(string))
		case "stop":
			option.stop, _ = fixedpoint.NewFromString(data["stop"].(string))
		case "comment":
			option.comment = data["comment"].(string)
		case "tag":
			option.tag = data["tag"].(string)

		}
	}

	//fmt.Println(option)

	s.OpenOrder(side, option)

}

/**
strategy.order(id, direction, qty, limit, stop, oca_name, oca_type, comment, alert_message)
*/

func (s *Strategy) Order(id string, direction Side, args ...any) {

}

/**
strategy.cancel(id) → void
*/

func (s *Strategy) Cancel(id string, direction Side, args ...any) {

}

/*
*
strategy.cancel_all() → void
*/
func (s *Strategy) CancelAll(id string, direction Side, args ...any) {

}

func (s *Strategy) CloseOrder(percentage fixedpoint.Value) error {
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

		_, err := s.GeneralOrderExecutor.SubmitOrders(s.ctx, *order)
		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}
	return errors.New("exceed retry limit")
}

/*
*strategy.close(id, comment, qty, qty_percent, alert_message, immediately) → void
 */
func (s *Strategy) Close(args ...any) {
	s.CloseOrder(fixedpoint.One)
}

/*
strategy.exit(id, from_entry, qty, qty_percent, profit, limit, loss, stop, trail_price, trail_points, trail_offset, oca_name, comment, comment_profit, comment_loss, comment_trailing, alert_message, alert_profit, alert_loss, alert_trailing)
*/

func (s *Strategy) Exit(tag string) error {
	percentage := fixedpoint.One
	order := s.p.NewMarketCloseOrder(percentage)

	if order == nil {
		return nil
	}
	order.Tag = tag
	order.TimeInForce = ""

	order.MarginSideEffect = types.SideEffectTypeAutoRepay
	for i := 0; i < closeOrderRetryLimit; i++ {
		price := s.getLastPrice()
		if !s.Session.Futures {
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
		}
		order.ReduceOnly = true

		if s.Market.IsDustQuantity(order.Quantity, price) {
			return nil
		}

		_, err := s.GeneralOrderExecutor.SubmitOrders(s.ctx, *order)

		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}
	return errors.New("exceed retry limit")
}
func (s *Strategy) CloseAll(id string, direction Side, args ...any) {

}

//
//func (s *Strategy) Position() *types.Position {
//	return s.Position
//}
