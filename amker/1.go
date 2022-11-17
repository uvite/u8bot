package main

import (
	"fmt"
	"github.com/c9s/bbgo/pkg/types"
	"time"
)

func main() {

	now := time.Now()
	since := now.Add(-5 * time.Minute)
	now.Before()
	time.Now().After(k.EndTime.Time())

	aa := types.ParseInterval(types.Interval1d)
	fmt.Println(since, aa)
}
