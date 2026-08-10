.PHONY: build git

GO := /usr/local/go/bin/go

build:
	$(GO) build -o rita .

git:
	@read -p "commit message: " msg; \
	git add -A; \
	git commit -m "$$msg"; \
	git push
