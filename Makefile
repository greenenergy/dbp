# For the version tagging, you must use annotated tags (or git gets unhappy).
# So:
# git tag -a vX.Y.Z
# <enter annotation - doesn't matter what you say, there just has to be one>

SRC = $(shell find . -name "*.go")
PROG = dbp
GIT_VERSION = $(shell git describe --long | sed 's/-g[0-9a-f]\{7,\}$$//')

build/amd64/$(PROG): $(SRC)
	@echo "Building version $(GIT_VERSION) for amd64"
	mkdir -p build/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -extldflags '-static' -X main.Version=$(GIT_VERSION)" -a -installsuffix cgo -o build/amd64/$(PROG)

build/arm64/$(PROG): $(SRC)
	@echo "Building version $(GIT_VERSION) for arm64"
	mkdir -p build/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-w -s -extldflags '-static' -X main.Version=$(GIT_VERSION)" -a -installsuffix cgo -o build/arm64/$(PROG)

all: build/arm64/$(PROG) build/amd64/$(PROG)

.PHONY: install
install:
	go install -v -ldflags "-w -s -X main.Version=$(GIT_VERSION)"

.PHONY: clean
clean:
	@rm -rf build/amd64 build/arm64

# Updated docker target for multi-arch build
docker: all
	@echo "Building multi-arch Docker image version $(GIT_VERSION)"
	(cd build && \
	docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -t livewireholdings/dbp:$(GIT_VERSION) \
        --push \
        -f Dockerfile \
        .)

# Local build option
docker-local: all
	@echo "Building multi-arch Docker image locally version $(GIT_VERSION)"
	(cd build && \
	docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -t livewireholdings/dbp:$(GIT_VERSION) \
        -t livewireholdings/dbp:latest \
        --load \
        -f Dockerfile \
        .)

docker-push:
	docker push livewireholdings/dbp:$(GIT_VERSION)

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-html: test
	go tool cover -html=coverage.out

versiontag:
	@-echo "$(GIT_VERSION)"
