package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func main()  {
	type res struct {
		err error
		output *[]byte
	}
	var (
		cmd *exec.Cmd
		err error
		ctx context.Context
		cancelFunc context.CancelFunc
		output []byte
		chanRes chan *res
	)


	chanRes = make(chan *res,10)
	ctx, cancelFunc = context.WithCancel(context.Background())
	cmd = exec.CommandContext(ctx,"bash.exe","-c","echo begin; sleep 3; ls -l")
	go func() {
		output,err = cmd.CombinedOutput()
		res := res{
			err:    err,
			output: &output,
		}
		chanRes <- &res
	}()

	time.Sleep(1)
	cancelFunc()
	temp := <- chanRes

	fmt.Println(temp.err)
	fmt.Println(string(*temp.output))

	cancelFunc()


}
