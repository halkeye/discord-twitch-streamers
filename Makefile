dev:
	 nodemon --signal SIGINT -e go,tpl -w . -w templates -x "go run ./... || exit 1"
