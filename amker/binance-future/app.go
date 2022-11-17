package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	//timeDur := time.Duration(1*60*4) * time.Minute //四小时
	timeDur := time.Duration(1) * time.Minute //四小时
	nowAddTime := now.Add(timeDur)

	//时间比较
	//时间之前比较
	//fmt.Println(nowAddTime.After(now)) //true
	////时间之后比较
	//fmt.Println(nowAddTime.Before(now)) //false
	////时间相等比较
	//fmt.Println(now.Equal(now)) //true
	tick := time.NewTicker(1 * time.Second)
	ch := make(chan int, 1024)

	for {
		select {

		case <-tick.C:
			fmt.Printf(": %t \n", nowAddTime.Before(time.Now()))
		}

		time.Sleep(200 * time.Millisecond)
	}
	close(ch)
	tick.Stop()

}
