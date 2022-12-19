package utils

import (
	"fmt"
	"os"
	"os/exec"
)

type OpenFaasFunc struct {
	Name             string
	FuncOpt          string
	IsCreateFromFile bool
	YmlFileAddress   string
}

func checkCmdExist(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err != nil {
		fmt.Println("Error: cannot find the cmd! Please check if installed or registered in $PATH")
		return false
	}
	return true
}

func (ofc *OpenFaasFunc) ExecOpenFaasCmd() error {
	var cmd *exec.Cmd
	fmt.Println(ofc.YmlFileAddress + ofc.Name + ".yml")
	if ofc.IsCreateFromFile {
		cmd = exec.Command(OpenFaasCmd, ofc.FuncOpt, CreateFromFile, ofc.YmlFileAddress+ofc.Name+".yml")
	} else {
		cmd = exec.Command(OpenFaasCmd, ofc.FuncOpt, CreateFromFile, ofc.YmlFileAddress+ofc.Name+".yml")
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
