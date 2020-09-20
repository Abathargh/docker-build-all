package main

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	DockerfileComponentNum = 2
	ArchIndex              = 1

	buildCmd     = "buildx build --platform %s --rm -f %s -t %s ."
	pushImageCmd = "push %s"

	createManifestCmd = "manifest create %s"
	pushManifestCmd   = "manifest push %s -p"
)

// A builder builds a series of images based on the passed architectures and on the
// options related to pushing and creating a manifest.
// Errors that occurs in the building/pushing stages are forwarded to the error channel.
type Builder struct {
	buildList []*DockerBuild
	manifest  *Manifest
	push      bool
	errChan   chan error
}

// Creates a new builder, taking care of setting it up in case a manifest has to be created too
func NewBuilder(buildList []*DockerBuild, imageName string, tag string, manifest bool, push bool, errChan chan error) *Builder {
	var nManifest *Manifest

	if manifest {
		manifestName := fmt.Sprintf("%s:%s", imageName, tag)
		nManifest = &Manifest{
			Builds: buildList,
			Name:   manifestName,
		}
	}

	return &Builder{
		buildList: buildList,
		manifest:  nManifest,
		push:      push,
		errChan:   errChan,
	}
}

// Executes a build (+ eventual push) of a single build, catching errors and
// forwarding them onto the error channel
func (b *Builder) single(build Buildable, push bool, errChan chan error) {
	if err := build.Build(); err != nil {
		errChan <- err
		return
	}

	if push {
		if err := build.Push(); err != nil {
			errChan <- err
			return
		}
	}
}

// Builds every image (+ eventual push and manifest creation)
// Every build + push is executed in parallel, the manifest creation + push
// is executed at the ened if everything is successful
func (b *Builder) BuildAll() {
	if b.buildList == nil || len(b.buildList) == 0 {
		fmt.Println("No Dockerfile found")
		return
	}

	var wGroup sync.WaitGroup
	wGroup.Add(len(b.buildList))

	defer close(b.errChan)

	for _, build := range b.buildList {
		go func(build DockerBuild, push bool) {
			defer wGroup.Done()
			b.single(build, push, b.errChan)
		}(*build, b.push)
	}

	wGroup.Wait()
	if b.manifest != nil {
		b.single(b.manifest, b.push, b.errChan)
	}
}

// Interface that describes the set of operation
// of a buildable object (image builds or manifest builds)
type Buildable interface {
	Build() error
	Push() error
}

// A Dockerbuild describes a docker image build of a specific
// architecture, with its complete image name.
type DockerBuild struct {
	Dockerfile string
	Arch       string
	BuildxArch string
	ImageName  string
}

// Creates a new docker build validating the dockerfile name and the architecture
func NewDockerBuild(dockerfile string, imageName string, tag string) (*DockerBuild, error) {
	dSplitted := strings.Split(dockerfile, ".")
	if len(dSplitted) != DockerfileComponentNum {
		return nil, errors.New("Wrong Format")
	}
	if !isSupported(dSplitted[ArchIndex]) {
		return nil, errors.New(fmt.Sprintf("Unsupported arch '%s'", dSplitted[1]))
	}

	arch, buildxArch := archFromFile(dockerfile)
	completeImageName := fmt.Sprintf("%s:%s-%s", imageName, arch, tag)
	return &DockerBuild{
		Dockerfile: dockerfile,
		Arch:       arch,
		BuildxArch: buildxArch,
		ImageName:  completeImageName,
	}, nil
}

func (b DockerBuild) Build() error {
	cmdString := fmt.Sprintf(buildCmd, b.BuildxArch, b.Dockerfile, b.ImageName)
	err := execDockerCmd(cmdString, "Building", b.Dockerfile)

	if err != nil {
		return err
	}
	return nil
}

func (b DockerBuild) Push() error {
	cmdString := fmt.Sprintf(pushImageCmd, b.ImageName)
	err := execDockerCmd(cmdString, "Pushing", b.ImageName)
	if err != nil {
		return err
	}
	return nil
}

// A manifest describes a manifest build with its imulti-arch
// images list to refer to.
type Manifest struct {
	Name   string
	Builds []*DockerBuild
}

func (m Manifest) Build() error {
	var strBuilder strings.Builder
	strBuilder.WriteString(m.Name)
	strBuilder.WriteString(" ")

	for i, build := range m.Builds {
		strBuilder.WriteString("--amend ")
		strBuilder.WriteString(build.ImageName)
		if i < len(m.Builds)-1 {
			strBuilder.WriteString(" ")
		}
	}

	cmdString := fmt.Sprintf(createManifestCmd, strBuilder.String())
	err := execDockerCmd(cmdString, "Building manifest", m.Name)
	if err != nil {
		return err
	}
	return nil
}

func (m Manifest) Push() error {
	cmdString := fmt.Sprintf(pushManifestCmd, m.Name)
	err := execDockerCmd(cmdString, "Pushing manifest", m.Name)
	if err != nil {
		return err
	}
	return nil
}
