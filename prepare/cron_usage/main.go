package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main()  {

	//crontab支持5位，这个库支持7个 包括了秒和年
	//分钟(0-59) 小时(0-23) 天(1-31) 月(1-12) 星期(0-6)

	expr,err := cronexpr.Parse("*/3 * * * * * *")
	if err != nil {
		fmt.Println(err)
		return
	}

	now := time.Now()
	nextTime := expr.Next(now)

	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("被调度了")
	})
	fmt.Println(now,nextTime)
	time.Sleep(10 *time.Second)


}
