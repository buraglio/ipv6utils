NAME    := ipv6utils
VERSION ?= 4
LDFLAGS  = -ldflags "-s -w -X main.version=$(VERSION)"
DIST    := dist

# Practical cross-compile targets for a networking utility.
# Each entry is OS/ARCH as understood by GOOS/GOARCH.
PLATFORMS := \
	linux/386        linux/amd64      linux/arm       linux/arm64    \
	linux/mips       linux/mipsle     linux/mips64    linux/mips64le \
	linux/ppc64le    linux/riscv64    linux/s390x                    \
	darwin/amd64     darwin/arm64                                    \
	windows/386      windows/amd64    windows/arm64                  \
	freebsd/amd64    freebsd/arm64                                   \
	openbsd/amd64    openbsd/arm64                                   \
	netbsd/amd64     netbsd/arm64                                    \
	dragonfly/amd64                                                  \
	solaris/amd64

.PHONY: all dist clean help $(PLATFORMS)

## all: build native binary in the current directory
all:
	go build $(LDFLAGS) -o $(NAME) .

## dist: build binaries for all supported platforms into ./dist/
dist: $(PLATFORMS)

## <os/arch>: build for a specific platform, e.g.  make linux/arm64
#  Windows targets automatically receive the .exe extension.
#  Shell variable expansion is used for parallel-build safety.
$(PLATFORMS):
	@set -e; \
	 mkdir -p "$(DIST)"; \
	 os="$(word 1,$(subst /, ,$@))"; \
	 arch="$(word 2,$(subst /, ,$@))"; \
	 ext=""; test "$$os" = "windows" && ext=".exe"; \
	 out="$(DIST)/$(NAME)_$${os}_$${arch}$$ext"; \
	 printf "building %-45s" "$$out ..."; \
	 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o "$$out" . \
	   && echo "ok" || echo "FAILED"

## clean: remove ./dist/ and the local native binary
clean:
	rm -rf $(DIST) $(NAME)

## help: list available targets
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST)
