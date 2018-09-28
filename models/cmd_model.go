package models

import "os/exec"

var GloableCmdList *exec.Cmd

func SetGloableCmdList(cmd *exec.Cmd) {
	GloableCmdList = cmd
}

func GetGloableCmdList() *exec.Cmd {
	return GloableCmdList
}
