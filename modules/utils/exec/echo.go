package exec

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	echoCmd  EchoCmdStr
	echoOnce sync.Once
)

type EchoCmdStr string

func GetEchoCmd() *EchoCmdStr {
	echoOnce.Do(func() {
		path, err := exec.LookPath("echo")
		if err != nil {
			panic("echo lookPath err")
		}
		echoCmd = EchoCmdStr(path)
	})
	return &echoCmd
}

func (echo *EchoCmdStr) Exec(arg ...string) *ECmdStr {
	fmt.Println("arg:", arg)
	return &ECmdStr{
		Cmd: exec.Command("sh", arg...),
		Arg: arg,
	}
}
