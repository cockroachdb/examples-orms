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

ADDR ?= cockroachdb://root@localhost:26257/company_sqlalchemy?sslmode=disable&disable_cockroachdb_telemetry=true

.PHONY: start
start:
	ADDR=$(ADDR) python3 ./server.py --port=6543

.PHONY: deps
deps:
	# To avoid permissions errors, the following should be run in a virtualenv
	# (preferred) or as root.
	pip3 install flask-sqlalchemy==2.5.1 sqlalchemy-cockroachdb==1.4.1 psycopg2 sqlalchemy==1.4.46
