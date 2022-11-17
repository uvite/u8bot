package scalpma

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	internal "github.com/c9s/bbgo/u8/nats"
	"github.com/davecgh/go-spew/spew"

	"time"
)

func (s *Strategy) PostionCheck(ctx context.Context) {
	s.Session.MarketDataStream.OnKLine(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {

		shortCondition := true

		if s.Position.IsShort() {

			fmt.Println("平空")
			_ = s.ClosePosition(ctx, fixedpoint.One, "")
			time.Sleep(10 * time.Second)
		}
		if shortCondition { // && ((previousRegime < zeroThreshold && previousRegime > -zeroThreshold) || s.grid.Index(1) > 0)
			createdOrders, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   s.Symbol,
				Side:     types.SideTypeSell,
				Quantity: s.Quantity,
				Type:     types.OrderTypeMarket,
				Tag:      "grid short: sell in",
			})

			fmt.Println(createdOrders)
			if err != nil {
				log.Errorln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}))
}

func (s *Strategy) PostionCheck1(ctx context.Context) {
	s.Session.MarketDataStream.OnKLine(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {

		shortCondition := true
		spew.Dump(s.Position)
		spew.Dump(s.Position.IsLong(), s.Position.IsOpened(1))
		if s.Position.IsLong() {

			fmt.Println("平多")

			s.Pubsub.Publish(internal.OrderCloseLong, internal.OrderPayload{
				Symbol:   s.Symbol,
				Side:     types.SideTypeBuy,
				Quantity: s.Quantity,
				Price:    kline.Close,
			})
			_ = s.ClosePosition(ctx, fixedpoint.One, "")
			time.Sleep(10 * time.Second)
		}
		if shortCondition { // && ((previousRegime < zeroThreshold && previousRegime > -zeroThreshold) || s.grid.Index(1) > 0)
			createdOrders, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   s.Symbol,
				Side:     types.SideTypeBuy,
				Quantity: s.Quantity,
				Type:     types.OrderTypeMarket,
				Tag:      "grid short: sell in",
			})
			s.Pubsub.Publish(internal.OrderOpenLong, internal.OrderPayload{
				Symbol:   s.Symbol,
				Side:     types.SideTypeBuy,
				Quantity: s.Quantity,
				Price:    kline.Close,
			})
			fmt.Println(createdOrders)
			if err != nil {
				log.Errorln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}))
}
