# docker-build-all

A simple tool to build docker images for multiple target architectures.

It searches for Dockerfiles in the format "Dockerfile.architecture" in the current directory and builds images with the passed_name:architecture tag.

## Requirements

In orderd for the program to work, you have to enable docker experimental features, since it uses docker buildx/manifest.

If you use the manifest/push features, you have to be logged in with your docker hub account through

```bash
docker login
```

## Instal and Run

```bash
go get -u github.com/abathargh/docker-build-all

# you can also clone it an build it manually
git clone https://github.com/Abathargh/docker-build-all
cd docker-build-all
go install

# usage
docker-build-all -n abathargh/test
```

## Usage

- **-n**, the name of the image, required;
- **-t**, images are tagged in the name:arch-tag format and name:tag (for manifests), pass a custom tag with -t, defaults to "latest";
- **-p**, specifies that you want to push after building.
- **-m**, specifies that you want to create a manifest that includes every built image after the build phase.