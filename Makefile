.PHONY: deploy
deploy:
	@./deploy.sh

.PHONY: dev
dev:
	@hugo server -D

.PHONY: dev-wsl
dev-wsl:
	@hugo server -D --bind 172.23.115.90 --baseURL=http://172.23.115.90 --disableFastRender

.PHONY: build
build:
	@hugo -t terminal

.PHONY: update_subs
update_subs:
	@git submodule update --rebase --remote