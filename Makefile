# Copyright 2016 The Cockroach Authors.
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
POSTGRES_TEST_TAG ?= 20170227-1358
EXAMPLES_ORMS_PATH = /examples-orms
DOCKER = docker run --volume="$(shell pwd)":/examples-orms cockroachdb/postgres-test:$(POSTGRES_TEST_TAG)

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
dockertest:
		$(DOCKER) make -C $(EXAMPLES_ORMS_PATH) deps test

# Run `git clean` in Docker to remove leftover files that are owned by root.
# This must be run after `dockertest` to ensure that successive CI runs don't
# fail.
.PHONY: dockergitclean
dockergitclean:
		$(DOCKER) /bin/bash -c "cd $(EXAMPLES_ORMS_PATH) && git clean -f -d -x ."

.PHONY: deps
deps:
	# TODO(nvanbenschoten) The following two lines are required for CI until
	# the Azure-Agents get updated next. If you are reading this, the lines
	# can probably be deleted now.
	rm -rf ../../lib/pq
	rm -rf ../../cockroachdb/cockroach-go
	$(GO) get -d -t ./...
	$(MAKE) deps -C ./java/hibernate
	$(MAKE) deps -C ./ruby/activerecord
