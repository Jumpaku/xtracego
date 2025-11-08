.DEFAULT_GOAL := help
.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?##.*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?##"}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version: ## make version VERSION=v2.0.0
	sed -e 's|^version: .*|version: $(VERSION)|g' cmd/xtracego/cli.cyamli.yaml > cmd/xtracego/cli.cyamli.yaml.new \
		&& mv cmd/xtracego/cli.cyamli.yaml.new cmd/xtracego/cli.cyamli.yaml
	make install
	xtracego version

.PHONY: install
install: ## Installs xtracego CLI tool locally.
	go generate ./...
	go install ./cmd/xtracego
