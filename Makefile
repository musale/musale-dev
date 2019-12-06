.PHONY: deploy
deploy:
	@./deploy.sh

.PHONY: dev
dev:
	@hugo server -D

.PHONY: build
build:
	@hugo -t hoyt

.PHONY: update_subs
update_subs:
	@git submodule update --rebase --remote