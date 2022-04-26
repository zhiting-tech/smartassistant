package exec

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	sedCmd  SedCmdStr
	sedOnce sync.Once
)

type SedCmdStr string

func GetSedCmd() *SedCmdStr {
	sedOnce.Do(func() {
		path, err := exec.LookPath("sed")
		if err != nil {
			panic("sed lookPath err")
		}
		sedCmd = SedCmdStr(path)
	})
	return &sedCmd
}

func (sed *SedCmdStr) Exec(arg ...string) *ECmdStr {
	fmt.Println("arg:", arg)
	return &ECmdStr{
		Cmd: exec.Command(string(*sed), arg...),
		Arg: arg,
	}
}
