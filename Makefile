.PHONY: clean lint test build k8s-up k8s-down
BIN_NAMES=controlplane redirect
CMD_DIRECTORY := ./cmd

# GIT_REPO := github.com/xvzf/lightpath
GIT_REPO := github.com/xvzf/lightpath
IMAGE_REPO := ghcr.io/xvzf/lightpath

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

# Default build target
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
DOCKER_BUILD_PLATFORMS ?= linux/amd64,linux/arm64

default: clean lint test build

lint:
	go vet ./...
	clang-format --Werror bpf/*.c bpf/headers/*.h -n
# golangci-lint run

clean:
	rm -rf cover.out

test: clean
	opa test -v  $(shell find . -name "*.rego")
	go test -v -race -cover ./...

dist:
	mkdir dist

build: clean dist $(patsubst %, build-%, $(BIN_NAMES))

build-%:
	@echo [$*] SHA: $(SHA) $(BUILD_DATE)
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -v \
		-ldflags '-X "${GIT_REPO}/internal/version.commit=${SHA}" -X "${GIT_REPO}/internal/version.date=${BUILD_DATE}" -X "${GIT_REPO}/internal/version.tag=${TAG_NAME}"' \
		-o "./dist/${GOOS}/${GOARCH}/${BIN_NAME}" ${CMD_DIRECTORY}/$*


build-linux-arm64: export GOOS := linux
build-linux-arm64: export GOARCH := arm64
build-linux-arm64:
	make build
build-linux-amd64: export GOOS := linux
build-linux-amd64: export GOARCH := amd64
build-linux-amd64:
	make build

## Build Multi archs Docker image
container-image-%: build-linux-amd64 build-linux-arm64
	make IMAGE_TAG=$* $(patsubst %, container-image-target-%, $(BIN_NAMES))

container-image-target-%:
	docker buildx build $(DOCKER_BUILDX_ARGS) --progress=chain -t $(IMAGE_REPO)/$*:${IMAGE_TAG} --platform=$(DOCKER_BUILD_PLATFORMS) -f $*.Dockerfile .

.PHONY: k8s-up
k8s-up:
	kind create cluster --config ./hack/ci/kind-cluster.yaml --wait 120s --name=lightpath-ci
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
	kubectl config set-context --current --namespace=lightpath-system

.PHONY: k8s-down
k8s-down:
	kind delete cluster --name=lightpath-ci

.PHONY: deploy
deploy:
	kubectl apply -k deploy/default/
