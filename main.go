package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pborman/getopt/v2"
)

const (
	valid    = "Dockerfile."
	buildCmd = "buildx build --platform %s --rm -f %s -t %s:%s ."
	usage    = `Build an image for each Dockerfile present in the folder.
Dockerfiles are in the Dockerfile.architecture format, where architecture is one between:
- arm32v7
- arm64
- amd64
`
)

// maybe aliases liek aarch64 is ok and refers to arm64?
var (
	dockerfiles []string
	wGroup      sync.WaitGroup
	errChan     chan error
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

func searchArchFiles(path string, f os.FileInfo, err error) error {
	if strings.HasPrefix(path, valid) {
		dSplitted := strings.Split(path, ".")
		if len(dSplitted) != 2 {
			return errors.New("Wrong Format")
		}
		if !isSupported(dSplitted[1]) {
			return errors.New("Unsupported arch")
		}
		dockerfiles = append(dockerfiles, path)
		//take out only final part
	}
	return nil
}

func buildImages(imagename string) {
	wGroup.Add(len(dockerfiles))

	defer close(errChan)
	defer wGroup.Wait()

	for _, dockerfile := range dockerfiles {
		go func(dockerfile string) {
			defer wGroup.Done()
			arch, formattedArch := archFromFile(dockerfile)
			cmdString := fmt.Sprintf(buildCmd, formattedArch, dockerfile, imagename, arch)
			fmt.Println(cmdString)
			cmdArgs := strings.Split(cmdString, " ")

			fmt.Printf("Building %s\n", dockerfile)
			cmd := exec.Command("docker", cmdArgs...)
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				var errContent string
				if stderr.String() == "" {
					errContent = out.String()
				} else {
					errContent = stderr.String()
				}
				strErr := fmt.Sprintf("%s: %s", err.Error(), errContent)
				errChan <- errors.New(strErr)
				return
			}
			fmt.Printf("Building %s: done\n", dockerfile)
		}(dockerfile)
	}
}

func main() {
	imgName := getopt.StringLong("name", 'n', "", "The base name of the images; this will generate images in the $name:tag format")
	help := getopt.BoolLong("help", 'h', "")
	getopt.Parse()

	if *help {
		fmt.Println(usage)
		getopt.PrintUsage(os.Stdout)
		return
	}

	if *imgName == "" {
		fmt.Println("a name is required")
		getopt.PrintUsage(os.Stdout)
		return
	}

	err := filepath.Walk(".", searchArchFiles)
	if err != nil {
		fmt.Println(err)
		getopt.PrintUsage(os.Stdout)
		return
	}

	errChan = make(chan error)
	go buildImages(*imgName)

	for err := range errChan {
		fmt.Print(err)
		return
	}
	fmt.Println("Done")
}
