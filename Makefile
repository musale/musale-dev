.PHONY: deploy
deploy:
	@./deploy.sh

.PHONY: dev
dev:
	@hugo server -D

.PHONY: build
build:
	@rm -rf public
	@hugo -t hermit
