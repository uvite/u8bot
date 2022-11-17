package main

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/pkg/cmd/cmdutil"
	"github.com/c9s/bbgo/pkg/exchange/binance"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
	"syscall"
	"time"
)

func init() {
	rootCmd.PersistentFlags().String("binance-api-key", "", "binance api key")
	rootCmd.PersistentFlags().String("binance-api-secret", "", "binance api secret")
	rootCmd.PersistentFlags().String("symbol", "ETHUSDT", "symbol")
	rootCmd.PersistentFlags().Float64("price", 20.0, "order price")
	rootCmd.PersistentFlags().Float64("quantity", 10.0, "order quantity")
}

var rootCmd = &cobra.Command{
	Use:   "binance-future",
	Short: "binance future",

	// SilenceUsage is an option to silence usage when an error occurs.
	SilenceUsage: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		key, secret := viper.GetString("binance-api-key"), viper.GetString("binance-api-secret")
		if len(key) == 0 || len(secret) == 0 {
			return errors.New("empty key or secret")
		}

		symbol, err := cmd.Flags().GetString("symbol")
		if err != nil {
			return err
		}

		price, err := cmd.Flags().GetFloat64("price")
		if err != nil {
			return err
		}

		quantity, err := cmd.Flags().GetFloat64("quantity")
		if err != nil {
			return err
		}

		var exchange = binance.New(key, secret)

		fmt.Println(symbol, price, quantity)

		//exchange.IsFutures = true
		//
		markets, err := exchange.QueryMarkets(ctx)
		if err != nil {
			return err
		}

		//
		market, ok := markets[symbol]
		if !ok {
			return fmt.Errorf("market %s is not defined", symbol)
		}
		//
		//marginAccount, err := exchange.QueryFuturesAccount(ctx)
		//if err != nil {
		//	return err
		//}
		//fmt.Println(market, marginAccount)
		go long(exchange, market)
		//go short(exchange, market)

		////
		////log.Infof("margin account: %+v", marginAccount)
		////
		////isolatedMarginAccount, err := exchange.QueryIsolatedMarginAccount(ctx)
		////if err != nil {
		////	return err
		////}
		////
		////log.Infof("isolated margin account: %+v", isolatedMarginAccount)
		////
		stream := exchange.NewStream()

		log.Info("connecting websocket...")
		if err := stream.Connect(ctx); err != nil {
			log.Fatal(err)
		}
		//
		time.Sleep(time.Second)
		//
		cmdutil.WaitForSignal(ctx, syscall.SIGINT, syscall.SIGTERM)
		return nil
	},
}

// 开多自带止赢止损
func long(exchange *binance.Exchange, market types.Market) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	price := 1340.0
	quantity := 0.9
	//市价单买入
	createdOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol: "ETHUSDT",
		Market: market,
		Side:   types.SideTypeBuy,
		Type:   types.OrderTypeMarket,
		Price:  fixedpoint.NewFromFloat(price),
		//StopPrice:        fixedpoint.NewFromFloat(price),
		Quantity: fixedpoint.NewFromFloat(quantity),

		//TimeInForce: "GTC",
		//ReduceOnly:       true,
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("create----:\n", createdOrder)
	//exchange.CancelOrders(ctx,)
	//限价单止赢

	createdTackProfitOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol:    "ETHUSDT",
		Market:    market,
		Side:      types.SideTypeSell,
		Type:      types.OrderTypeTakeProfitLimit,
		Price:     fixedpoint.NewFromFloat(price + 20),
		StopPrice: fixedpoint.NewFromFloat(price + 20),
		Quantity:  fixedpoint.NewFromFloat(quantity),

		TimeInForce: "GTC",
		ReduceOnly:  true,
	})
	fmt.Println("take,profit----::\n", createdTackProfitOrder)
	//限价单止损

	createdStopOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol:    "ETHUSDT",
		Market:    market,
		Side:      types.SideTypeSell,
		Type:      types.OrderTypeStopLimit,
		Price:     fixedpoint.NewFromFloat(price - 20),
		StopPrice: fixedpoint.NewFromFloat(price - 20),
		Quantity:  fixedpoint.NewFromFloat(quantity),

		TimeInForce: "GTC",
		ReduceOnly:  true,
	})
	fmt.Println("stop-----:", createdStopOrder)

	//exchange.QueryOrderTrades()
}

// 开空自带止赢止损
func short(exchange *binance.Exchange, market types.Market) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	price := 1340.0
	quantity := 0.9
	//市价单买入
	createdOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol: "ETHUSDT",
		Market: market,
		Side:   types.SideTypeSell,
		Type:   types.OrderTypeMarket,
		Price:  fixedpoint.NewFromFloat(price),
		//StopPrice:        fixedpoint.NewFromFloat(price),
		Quantity: fixedpoint.NewFromFloat(quantity),

		TimeInForce: "GTC",
		//ReduceOnly:       true,
	})
	if err != nil {
		fmt.Println(err)
	}

	log.Info(createdOrder)

	//限价单止赢

	createdTackProfitOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol:    "ETHUSDT",
		Market:    market,
		Side:      types.SideTypeBuy,
		Type:      types.OrderTypeTakeProfitLimit,
		Price:     fixedpoint.NewFromFloat(price - 20),
		StopPrice: fixedpoint.NewFromFloat(price - 20),
		Quantity:  fixedpoint.NewFromFloat(quantity),

		TimeInForce: "GTC",
		ReduceOnly:  true,
	})
	fmt.Println(createdTackProfitOrder)
	//限价单止损

	createdStopOrder, err := exchange.SubmitOrder(ctx, types.SubmitOrder{
		Symbol:    "ETHUSDT",
		Market:    market,
		Side:      types.SideTypeBuy,
		Type:      types.OrderTypeStopLimit,
		Price:     fixedpoint.NewFromFloat(price + 20),
		StopPrice: fixedpoint.NewFromFloat(price + 20),
		Quantity:  fixedpoint.NewFromFloat(quantity),

		TimeInForce: "GTC",
		ReduceOnly:  true,
	})
	fmt.Println(createdStopOrder)
}

func main() {
	if _, err := os.Stat(".env.local"); err == nil {
		if err := godotenv.Load(".env.local"); err != nil {
			log.Fatal(err)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Error("bind pflags error")
	}

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.WithError(err).Error("cmd error")
	}
}
