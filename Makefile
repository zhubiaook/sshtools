# 定义全局 Makefile 变量方便后面引用
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR)/ && pwd -P))
OUTPUT_DIR := $(ROOT_DIR)/_output

# Platform相关变量
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PLATFORM ?= $(GOOS)_$(GOARCH)
PLATFORMS ?= linux_amd64 linux_arm64 $(PLATFORM)

# 编译程序变量
COMMANDS ?= $(filter-out %.md, $(wildcard ${ROOT_DIR}/cmd/*))
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))

# 版本相关变量
VERSION_PACKAGE=sshtools/pkg/version

ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --tags --always --match='v*')
endif

GIT_TREE_STATE:="dirty"
ifeq (, $(shell git status --porcelain 2>/dev/null))
	GIT_TREE_STATE="clean"
endif

GIT_COMMIT:=$(shell git rev-parse HEAD)

GO_LDFLAGS += \
	-X $(VERSION_PACKAGE).GitVersion=$(VERSION) \
	-X $(VERSION_PACKAGE).GitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PACKAGE).GitTreeState=$(GIT_TREE_STATE) \
	-X $(VERSION_PACKAGE).BuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')


.PHONY: all
all: build

## build: Build source code for host platform.
.PHONY: build
build: $(addprefix build., $(addprefix $(PLATFORM)., $(BINS)))

## build.multiarch: Build source code for multiple platforms. See option PLATFORMS.
.PHONY: build.multiarch
build.multiarch: $(foreach p,$(PLATFORMS),$(addprefix build., $(addprefix $(p)., $(BINS))))

.PHONY: build.%
build.%: tidy 
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/bin/$(OS)/$(ARCH)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "$(GO_LDFLAGS)" -o $(OUTPUT_DIR)/bin/$(OS)/$(ARCH)/$(COMMAND) $(ROOT_DIR)/cmd/$(COMMAND)
	@mkdir -p $(OUTPUT_DIR)/release
	@cp $(OUTPUT_DIR)/bin/$(OS)/$(ARCH)/$(COMMAND) $(OUTPUT_DIR)/release/$(COMMAND)-$(VERSION)-$(OS)-$(ARCH)
	@cd $(OUTPUT_DIR)/release && sha256sum $(COMMAND)-$(VERSION)-$(OS)-$(ARCH) >> $(OUTPUT_DIR)/release/SHA256SUMS

.PHONY: tidy
tidy: 
	@go mod tidy

## clean: Remove all files that are created by building.
.PHONY: clean
clean: 
	@-rm -vrf $(OUTPUT_DIR)

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "Usage: make <TARGETS>\n\nTargets:\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
