.PHONY: build

SHELL := /bin/bash
CHART_DIRS := $(shell find ./charts -mindepth 1 -maxdepth 1 -type d)

build:
	@for dir in $(CHART_DIRS); do \
		echo "Building dependencies for $$dir"; \
		helm dependency build $$dir; \
	done

.PHONY: setup
setup:
	kubectl create namespace odysseia
	kubectl -n elastic-system apply -f https://download.elastic.co/downloads/eck/2.10.0/crds.yaml
	kubectl -n elastic-system apply -f https://download.elastic.co/downloads/eck/2.10.0/operator.yaml
