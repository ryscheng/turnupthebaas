lint:
	golint ./... | grep -v ^vendor || echo "golint: done"

.PHONY: lint
