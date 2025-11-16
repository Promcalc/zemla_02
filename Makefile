##@ Versioning
.PHONY: version
version: ## Show current version information
	@echo "Current version information:"
	@echo "Git tag: $(shell git describe --tags --exact-match 2>/dev/null || echo 'none')"
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Go version: $(shell go version | cut -d ' ' -f 3)"
	@echo "Build date: $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')"

.PHONY: version-tag
version-tag: ## Create a version tag (usage: make version-tag VERSION=x.y.z)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required. Usage: make version-tag VERSION=x.y.z"; \
		exit 1; \
	fi
	@echo "Creating version tag v$(VERSION)..."
	git tag -a "v$(VERSION)" -m "Release version $(VERSION)"
	git push origin "v$(VERSION)"
	@echo "Version tag v$(VERSION) created successfully!"

.PHONY: version-list
version-list: ## List all version tags
	@echo "Available version tags:"
	git tag -l 'v*' --sort=-version:refname

.PHONY: build-with-version
build-with-version: ## Build binaries with version information
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo 'dev') && \
	COMMIT_SHA=$$(git rev-parse HEAD) && \
	BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
	echo "Building with version: $$VERSION, commit: $$COMMIT_SHA, date: $$BUILD_DATE" && \
	CGO_ENABLED=1 go build -ldflags="-w -s \
	-X 'main.version=$$VERSION' \
	-X 'main.commit=$$(echo $$COMMIT_SHA | cut -c1-7)' \
	-X 'main.buildDate=$$BUILD_DATE' \
	-X 'main.goVersion=$$(go version | cut -d ' ' -f 3)'" \
	-o bin/collector ./cmd/collector && \
	CGO_ENABLED=1 go build -ldflags="-w -s \
	-X 'main.version=$$VERSION' \
	-X 'main.commit=$$(echo $$COMMIT_SHA | cut -c1-7)' \
	-X 'main.buildDate=$$BUILD_DATE' \
	-X 'main.goVersion=$$(go version | cut -d ' ' -f 3)'" \
	-o bin/web ./cmd/web

##@ Docker (обновленные цели)
.PHONY: docker-build-version
docker-build-version: ## Build Docker images with version tags
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo 'latest') && \
	SHORT_SHA=$$(git rev-parse --short HEAD) && \
	COMMIT_SHA=$$(git rev-parse HEAD) && \
	BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
	echo "Building Docker images with version: $$VERSION, short SHA: $$SHORT_SHA" && \
	docker build -t lot-collector:$$VERSION -t lot-collector:$$SHORT_SHA --build-arg APP=collector --build-arg VERSION=$$VERSION --build-arg COMMIT_SHA=$$COMMIT_SHA --build-arg BUILD_DATE=$$BUILD_DATE -f Dockerfile . && \
	docker build -t lot-web:$$VERSION -t lot-web:$$SHORT_SHA --build-arg APP=web --build-arg VERSION=$$VERSION --build-arg COMMIT_SHA=$$COMMIT_SHA --build-arg BUILD_DATE=$$BUILD_DATE -f Dockerfile .