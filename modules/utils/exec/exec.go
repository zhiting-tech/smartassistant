package exec

import (
	"io/ioutil"
	"os"
	"os/exec"
)

type ECmdStr struct {
	*exec.Cmd
	Arg []string
}

func GetCmd(cmd string) ECmd {
	switch cmd {
	case "sed":
		return GetSedCmd()
	case "echo":
		return GetEchoCmd()
	default:
		return nil
	}
}

type ECmd interface {
	Exec(arg ...string) *ECmdStr
}

func (ec *ECmdStr) ExecOutPut() error {
	_, err := ec.Output()
	return err
}

func (ec *ECmdStr) ExecRedirect() error {
	stout, err := ec.StdoutPipe()
	if err != nil {
		return err
	}
	defer stout.Close()
	if err = ec.Start(); err != nil {
		return err
	}
	data, err := ioutil.ReadAll(stout)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(ec.Arg[len(ec.Arg)-1], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return err
	}
	return nil
}
