package main

import (
	"fmt"
	"os/exec"
)

func main()  {
	var (
		cmd *exec.Cmd
		err error
		res []byte
	)
	cmd = exec.Command("D:\\software\\cygwin64\\bin\\bash.exe","-c","ls -l")
	if res,err = cmd.CombinedOutput();err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(res))

}
