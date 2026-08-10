.PHONY: build install git

GO := /usr/local/go/bin/go

build:
	$(GO) build -o rita .

install: build
	sudo ln -sf $(CURDIR)/rita /usr/local/bin/rita

git:
	@read -p "commit message: " msg; \
	git add -A; \
	git commit -m "$$msg"; \
	git push
