# docker-build-all

A simple tool to build docker images for multiple target architectures.
It searches for Dockerfiles in the format "Dockerfile.architecture" in the current directory and builds images with the passedname:architecture tag.


```bash
go get github.com/abathargh/docker-build-all

# usage
docker-build-all -n abathargh/test
```