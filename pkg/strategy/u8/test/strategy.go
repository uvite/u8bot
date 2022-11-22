package test

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/jsvm"
	"github.com/sirupsen/logrus"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

// ID is the unique strategy ID, it needs to be in all lower case
// For example, grid strategy uses "grid"
const ID = "u8"

// log is a logrus.Entry that will be reused.
// This line attaches the strategy field to the logger with our ID, so that the logs from this strategy will be tagged with our ID
var log = logrus.WithField("strategy", ID)

var ten = fixedpoint.NewFromInt(10)

// init is a special function of golang, it will be called when the program is started
// importing this package will trigger the init function call.
func init() {
	// Register our struct type to BBGO
	// Note that you don't need to field the fields.
	// BBGO uses reflect to parse your type information.
	bbgo.RegisterStrategy(ID, &Strategy{})
}

// State is a struct contains the information that we want to keep in the persistence layer,
// for example, redis or json file.
type State struct {
	Counter int `json:"counter,omitempty"`
}

// Strategy is a struct that contains the settings of your strategy.
// These settings will be loaded from the BBGO YAML config file "bbgo.yaml" automatically.
type Strategy struct {
	Symbol string `json:"symbol"`
	types.IntervalWindow

	JsVm *jsvm.JsVm
	// State is a state of your strategy
	// When BBGO shuts down, everything in the memory will be dropped
	// If you need to store something and restore this information back,
	// Simply define the "persistence" tag
	State *State `persistence:"state"`
}

// ID should return the identity of this strategy
func (s *Strategy) ID() string {
	return ID
}

// InstanceID returns the identity of the current instance of this strategy.
// You may have multiple instance of a strategy, with different symbols and settings.
// This value will be used for persistence layer to separate the storage.
//
// Run:
//
//	redis-cli KEYS "*"
//
// And you will see how this instance ID is used in redis.
func (s *Strategy) InstanceID() string {
	return ID + ":" + s.Symbol
}

//func (s *Strategy) JsVm(jsvm *vm.JsVm) {
//	// We want 1m kline data of the symbol
//	// It will be BTCUSDT 1m if our s.Symbol is BTCUSDT
//	fmt.Println(222)
//	s.Vm = jsvm
//
//	//s.Vm.Default(s.Symbol)
//}

// Subscribe method subscribes specific market data from the given session.
// Before BBGO is connected to the exchange, we need to collect what we want to subscribe.
// Here the strategy needs kline data, so it adds the kline subscription.
func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	// We want 1m kline data of the symbol
	// It will be BTCUSDT 1m if our s.Symbol is BTCUSDT

	fmt.Println(3333)
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: "1m"})
}

// This strategy simply spent all available quote currency to buy the symbol whenever kline gets closed
func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	// Initialize the default value for state
	//
	//if ok := s.JsVm.Vu.RunOnce(); ok != nil {
	//
	//	return fmt.Errorf("jsvm run err : %w", ok)
	//
	//}
	if s.State == nil {
		s.State = &State{Counter: 1}
	}

	// here we define a kline callback
	// when a kline is closed, we will do something
	session.MarketDataStream.OnKLine(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {

	}))
	//session.MarketDataStream.OnKLineClosed(callback)

	// if you need to do something when the user data stream is ready
	// note that you only receive order update, trade update, balance update when the user data stream is connect.
	session.UserDataStream.OnStart(func() {
		log.Infof("connected")
	})

	return nil
}
