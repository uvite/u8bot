package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Coin struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

func (s *Server) coin(c *gin.Context) {
	coins := []Coin{}
	coins = append(coins, Coin{
		Id:    1,
		Name:  "BTCUSDT",
		Base:  "BTC",
		Quote: "USDT",
	})
	coins = append(coins, Coin{
		Id:    2,
		Name:  "ETHUSDT",
		Base:  "ETH",
		Quote: "USDT",
	})

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": coins})
}
