PKGS = $(shell go list ./... | grep -v /vendor/)
GOFILES = $(shell find . -name '*.go' -and -not -path "./vendor/*")
UNFMT = $(shell gofmt -l ${GOFILES})
GIT_COMMIT = $(shell git rev-parse --short HEAD)

CGO_ENABLED = 0

XC_OS ?= linux darwin
XC_ARCH ?= amd64

lint: fmtcheck
	@golangci-lint run --config golangci.yml

.PHONY: lint

test:
	@echo "==> Running tests"
	@go test -v -count=1 -timeout=300s ${PKGS}

.PHONY: test

fmt:
	@echo "==> Fixing code with gofmt"
	@gofmt -s -w ${GOFILES}

.PHONY: fmt

fmtcheck:
	@echo "==> Checking code for gofmt compliance"
	@[ -z "${UNFMT}" ] || ( echo "Following files are not gofmt compliant.\n\n${UNFMT}\n\nRun 'make fmt' for reformat code"; exit 1 )

.PHONY: fmtcheck

build:
	@go build -a -o bin/corn -ldflags "-s -w -extldflags \"-static\" -X 'cmd.Version=${GIT_COMMIT}'"

.PHONY: build

define xc-target
  $1/$2:
	@printf "%s%20s %s\n" "-->" "${1}/${2}:" "corn"
	@CGO_ENABLED=0 GOOS="${1}" GOARCH="${2}" \
		go build -a -o "bin/corn-${1}_${2}" \
			-ldflags "-s -w -extldflags \"-static\" -X 'cmd.Version=${GIT_COMMIT}'"
  .PHONY: $1/$2

  $1:: $1/$2
  .PHONY: $1

  xc:: $1/$2
  .PHONY: xc
endef
$(foreach goarch,$(XC_ARCH),$(foreach goos,$(XC_OS),$(eval $(call xc-target,$(goos),$(goarch)))))