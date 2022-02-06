package main

import (
	"fmt"
	"os/exec"
)

func main()  {
	cmd := exec.Command("D:\\software\\cygwin64\\bin\\bash.exe","-c","echo 1")
	err := cmd.Run()
	fmt.Println(err)
}
