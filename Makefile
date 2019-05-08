.DEFAULT_GOAL := help

dev: ## run the dev server
	 nodemon --signal SIGINT -e go,tpl -w . -w templates -x "go run ./... || exit 1"

docker: ## build the docker version
	docker build -t halkeye/discord-twitch-streamers .

release: ## bump the version and push a release
	bumpversion patch && git push origin master --tags

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
