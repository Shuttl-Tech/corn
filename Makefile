PKGS = $(shell go list ./... | grep -v /vendor/)
GOFILES = $(shell find . -name '*.go' -and -not -path "./vendor/*")
UNFMT = $(shell gofmt -l ${GOFILES})
GIT_COMMIT = $(shell git rev-parse --short HEAD)

CGO_ENABLED = 0

XC_OS ?= linux darwin
XC_ARCH ?= amd64

VERSION ?= ${GIT_COMMIT}
GPG_KEY ?=

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

release: clean xc _compress _checksum _sign


define xc-target
  $1/$2:
	@printf "%s%20s %s\n" "-->" "${1}/${2}:" "corn"
	@CGO_ENABLED=0 GOOS="${1}" GOARCH="${2}" \
		go build -a -o "bin/corn-${1}_${2}" \
			-ldflags "-s -w -extldflags \"-static\" -X 'cmd.Version=${GIT_COMMIT}-${GIT_COMMIT}'"
  .PHONY: $1/$2

  $1:: $1/$2
  .PHONY: $1

  xc:: $1/$2
  .PHONY: xc
endef
$(foreach goarch,$(XC_ARCH),$(foreach goos,$(XC_OS),$(eval $(call xc-target,$(goos),$(goarch)))))

_compress:
	@echo "--> Compressing artifacts"
	@for each in bin/*; do \
       zip "$${each}.zip" "$${each}" &>/dev/null; \
       rm -f "$${each}"; \
    done

.PHONY: _compress

_checksum:
	@echo "--> Generating checksum"
	@cd bin/ && \
		shasum --algorithm 256 * > "corn_${VERSION}_SHA256SUM" && \
		cd - &>/dev/null

.PHONY: _checksum

_sign:
ifeq ($(VERSION),$(GIT_COMMIT))
	@echo "==> ERROR: Release version is not set. This release cannot be tagged using git commit hash"
	@echo "           Set the environment variable VERSION to the version number for this release"
	@echo "           and try again"
	@exit 127
endif

ifndef GPG_KEY
	@echo "==> ERROR: No GPG key specified! Without a GPG key, this release cannot"
	@echo "           be signed. Set the environment variable GPG_KEY to the ID of"
	@echo "           the GPG key to continue."
	@exit 127
else
	@echo "--> Tagging and signing the release"
	@gpg --default-key "${GPG_KEY}" --detach-sig "bin/corn_${VERSION}_SHA256SUM"
	@git commit --allow-empty --gpg-sign="${GPG_KEY}" --quiet --signoff --message "Release v${VERSION}"
	@git tag --annotate --create-reflog --local-user "${GPG_KEY}" --message "Version ${VERSION}" --sign "v${VERSION}" master
	@echo "--> Run following command to push tag"
	@echo ""
	@echo "    git push && git push --tags"
	@echo ""
	@echo "Then upload the binaries and checksum in bin/"
endif

.PHONY: _sign

clean:
	@rm -f bin/*