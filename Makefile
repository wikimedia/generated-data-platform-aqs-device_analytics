# Copyright 2021 Nikki Nikkhoui <nnikkhoui@wikimedia.org> and Wikimedia Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

APPNAME     = unique-devices
VERSION     = $(shell /usr/bin/git describe --always)
BUILD_DATE  = $(shell date -u +%Y-%m-%dT%T:%Z)
HOSTNAME    = $(shell hostname)

CONFIG     ?= config.yaml

GO_LDFLAGS  = -X main.version=$(if $(VERSION),$(VERSION),unknown)
GO_LDFLAGS += -X main.buildDate=$(if $(BUILD_DATE),$(BUILD_DATE),unknown)
GO_LDFLAGS += -X main.buildHost=$(if $(HOSTNAME),$(HOSTNAME),unknown)


build:
	@echo "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
	@echo "VERSION ......: $(VERSION)"
	@echo "BUILD HOST ...: $(HOSTNAME)"
	@echo "BUILD DATE ...: $(BUILD_DATE)"
	@echo "GO VERSION ...: $(word 3, $(shell go version))"
	@echo "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
	go build -ldflags "$(GO_LDFLAGS)" -o $(APPNAME) .

run:
	go run -ldflags "$(GO_LDFLAGS)" . -config $(CONFIG)

test:
	go test

check:
	@if [ -n "`goimports -l *.go`" ]; then \
	    echo "goimports: format errors detected" >&2; \
	    false; \
	fi
	@if [ -n "`gofmt -l *.go`" ]; then \
	    echo "gofmt: format errors detected" >&2; \
	    false; \
	fi
	go vet ./...

clean:
	rm -f $(APPNAME)

.PHONY: build run test clean
