package genv

import (
	"github.com/uvite/u8/plugin/xk6-ta/ta"
	"github.com/uvite/u8/tart/floats"
)

type Strategy struct {
	close *floats.Slice
	high  floats.Slice
}

func NewStragegy() *Strategy {

	return &Strategy{}
}

func (s *Strategy) Run() {
}
func (s *Strategy) GetSMA() floats.Slice {

	ta := ta.Ta{}
	sma := ta.Sma(*s.close, 14)
	return sma
}

func (s *Strategy) Close(close *floats.Slice) {
	s.close = close
}
