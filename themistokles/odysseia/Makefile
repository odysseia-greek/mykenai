.PHONY: build

SHELL := /bin/bash
CHART_DIRS := $(shell find ./charts -mindepth 1 -maxdepth 1 -type d)

build:
	@for dir in $(CHART_DIRS); do \
		echo "Building dependencies for $$dir"; \
		helm dependency build $$dir; \
	done
