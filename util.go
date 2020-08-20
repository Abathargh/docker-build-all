package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func isSupported(arch string) bool {
	var supportedArchs = []string{"arm32v7", "arm64", "amd64"}
	for i := 0; i < len(supportedArchs); i++ {
		if supportedArchs[i] == arch {
			return true
		}
	}
	return false
}

func archFromFile(file string) (string, string) {
	arch := strings.Split(file, ".")[1]
	mapping := map[string]string{
		"amd64":   "linux/amd64",
		"arm32v7": "linux/arm/v7",
		"arm64":   "linux/arm64",
	}
	return arch, mapping[arch]
}

func execDockerCmd(cmdString string, cmdName string, targetName string) error {
	cmdArgs := strings.Split(cmdString, " ")
	cmd := exec.Command("docker", cmdArgs...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	fmt.Printf("%s %s\n", cmdName, targetName)
	err := cmd.Run()

	// TODO comment on usage of stdout vs stderr
	if err != nil {
		var errContent string
		if stderr.String() == "" {
			errContent = out.String()
		} else {
			errContent = stderr.String()
		}
		strErr := fmt.Sprintf("%s: %s", err.Error(), errContent)
		return errors.New(strErr)
	}
	fmt.Printf("%s %s: done\n", cmdName, targetName)
	return nil
}
