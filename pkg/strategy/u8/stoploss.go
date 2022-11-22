package u8

func (s *Strategy) CheckStopLoss() bool {
	if s.UseStopLoss {
		stoploss := s.StopLoss.Float64()
		if s.sellPrice > 0 && s.sellPrice*(1.+stoploss) <= s.highestPrice ||
			s.buyPrice > 0 && s.buyPrice*(1.-stoploss) >= s.lowestPrice {
			return true
		}
	}
	// todo 暂时可能用不到
	//if s.UseAtr {
	//	atr := s.atr.Last()
	//	if s.sellPrice > 0 && s.sellPrice+atr <= s.highestPrice ||
	//		s.buyPrice > 0 && s.buyPrice-atr >= s.lowestPrice {
	//		return true
	//	}
	//}
	return false
}
