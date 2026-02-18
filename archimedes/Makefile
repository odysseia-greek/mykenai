project_name=archimedes
install_dir=$(HOME)/bin
binary_path=$(install_dir)/$(project_name)
VERSION ?=1.0.8

.PHONY: build build-auto build-custom

build: build-auto

build-auto:
	@current_version="$$( \
		$(binary_path) --version 2>/dev/null | sed -E 's/.*version[[:space:]]+([0-9]+\.[0-9]+\.[0-9]+).*/\1/' || true \
	)"; \
	if [ -z "$$current_version" ]; then \
		current_version="$$( \
			$(project_name) --version 2>/dev/null | sed -E 's/.*version[[:space:]]+([0-9]+\.[0-9]+\.[0-9]+).*/\1/' || true \
		)"; \
	fi; \
	if [ -z "$$current_version" ]; then \
		current_version="$(VERSION)"; \
	fi; \
	new_version="$$(echo "$$current_version" | awk -F. '{print $$1 "." $$2 "." $$3+1}')"; \
	echo "Detected version: $$current_version"; \
	echo "Building version: $$new_version"; \
	go fmt ./...; \
	go build -ldflags "-X main.version=$$new_version" cmd/archimedes.go; \
	mv archimedes $(binary_path)

build-custom:
	@echo "Building custom version: $(VERSION)"
	go fmt ./...
	go build -ldflags "-X main.version=$(VERSION)" cmd/archimedes.go
	mv archimedes $(binary_path)
