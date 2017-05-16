# Copyright 2017 The Cockroach Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
# implied. See the License for the specific language governing
# permissions and limitations under the License. See the AUTHORS file
# for names of contributors.
#
# Author: Nathan VanBenschoten (nvanbenschoten@gmail.com)

GO ?= go
POSTGRES_TEST_TAG ?= 20170308-1644
DOCKER_GOPATH = /root/go
DOCKER_REPO_PATH = $(DOCKER_GOPATH)/src/github.com/cockroachdb/examples-orms
DOCKER = docker run --volume="$(shell pwd)/../../../..":$(DOCKER_GOPATH) docker.io/cockroachdb/postgres-test:$(POSTGRES_TEST_TAG)
#                                          ^  ^  ^  ^~ GOPATH
#                                          |  |  |~ GOPATH/src
#                                          |  |~ GOPATH/src/github.com
#                                          |~ GOPATH/src/github.com/cockroachdb

.PHONY: all
all: test

ifneq ($(COCKROACH_BINARY),)
BINARYFLAG = -cockroach-binary=$(COCKROACH_BINARY)
endif

.PHONY: test
test:
	$(GO) test -v -i ./testing
	$(GO) test -v ./testing $(BINARYFLAG)

.PHONY: dockertest
dockertest: godeps
		$(DOCKER) make -C $(DOCKER_REPO_PATH) ormdeps test

# Run `git clean` in Docker to remove leftover files that are owned by root.
# This must be run after `dockertest` to ensure that successive CI runs don't
# fail.
.PHONY: dockergitclean
dockergitclean:
		$(DOCKER) /bin/bash -c "cd $(DOCKER_REPO_PATH) && git clean -f -d -x ."

.PHONY: deps
deps: godeps ormdeps

.PHONY: godeps
godeps:
	$(GO) get -d -u -t ./...

.PHONY: ormdeps
ormdeps:
	$(MAKE) deps -C ./java/hibernate
	$(MAKE) deps -C ./node/sequelize
	$(MAKE) deps -C ./python/sqlalchemy
	$(MAKE) deps -C ./ruby/activerecord
