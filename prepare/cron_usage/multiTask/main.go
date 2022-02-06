package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type Task struct {
	expr *cronexpr.Expression
	nextTime time.Time
}

func main()  {
	expr,err := cronexpr.Parse("*/5 * * * * * * ")
	if err != nil {
		fmt.Println(err)
		return
	}
	expr1 := cronexpr.MustParse("*/5 * * * * * * ")


	forerver := make(chan bool)

	taskMap := make(map[string]*Task)
	taskMap["Job1"] = &Task{
		expr:     expr,
		nextTime: expr.Next(time.Now()),
	}
	taskMap["Job2"] = &Task{
		expr:     expr1,
		nextTime: expr1.Next(time.Now()),
	}
	var now time.Time
	go func() {
		for  {
			now = time.Now()
			for key,val := range taskMap {
				if val.nextTime.Before(now) || val.nextTime.Equal(now) {
					go func(name string) {
						fmt.Println(time.Now(),name)
					}(key)
					val.nextTime = val.expr.Next(now)
				}
			}
			select {
				case <- time.NewTimer(200 * time.Millisecond).C:
			}
		}


	}()
	<-forerver
}
