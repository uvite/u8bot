package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Bots struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DataType    string `json:"dataType"`
	DataVersion string `json:"dataVersion"`
	DataDesc    string `json:"dataDesc"`
}

func (s *Server) bots(c *gin.Context) {

	bots := []Bots{}
	bots = append(bots, Bots{
		Id:          1,
		Name:        "均线剥头皮",
		DataType:    "long",
		DataVersion: "1",
		DataDesc:    "五分钟BTC 做多",
	})
	bots = append(bots, Bots{
		Id:          2,
		Name:        "均线交叉",
		DataType:    "short",
		DataVersion: "1",
		DataDesc:    "30分钟eth 做空",
	})

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": bots})
}
