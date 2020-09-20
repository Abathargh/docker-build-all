package main

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pborman/getopt/v2"
)

const (
	valid = "Dockerfile."
	usage = `Build an image for each Dockerfile present in the folder.
Dockerfiles are in the Dockerfile.architecture format, where architecture is one between:
- arm32v7
- arm64
- amd64
`
)

func main() {
	imgName := getopt.StringLong("name", 'n', "", "The base name of the images; this will generate images in the $name:tag format")
	imgTag := getopt.StringLong("tag", 't', "latest", "The tag to append to the image tag; the images are generated in the name:arch-tag, with the default being 'latest'")
	manifest := getopt.BoolLong("manifest", 'm', "Create a manifest including all the built images")
	push := getopt.BoolLong("push", 'p', "Push the images after building them; if a manifest is also created, it is pushed too")
	help := getopt.BoolLong("help", 'h', "Show info about the app")

	getopt.Parse()

	if *help {
		fmt.Println(usage)
		getopt.PrintUsage(os.Stdout)
		return
	}

	if *imgName == "" {
		fmt.Println("You have to pass a name and a tag")
		getopt.PrintUsage(os.Stdout)
		return
	}

	var buildList []*DockerBuild

	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if strings.HasPrefix(path, valid) {
			build, err := NewDockerBuild(path, *imgName, *imgTag)
			if err != nil {
				return err
			}
			buildList = append(buildList, build)
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		getopt.PrintUsage(os.Stdout)
		return
	}

	errChan := make(chan error, len(buildList))
	builder := NewBuilder(buildList, *imgName, *imgTag, *manifest, *push, errChan)
	go builder.BuildAll()

	// Eiher wait for an error on errChan or for
	// the Builder to close it
	for err := range errChan {
		fmt.Println(err)
		return
	}
	fmt.Println("Done")
}
