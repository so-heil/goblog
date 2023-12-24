build-templ:
	templ generate

build-tailwind:
	npm run build

build-webserver:
	go build -o ./tmp/main cmd/website/main.go

test:
	go test -v business/articles/articles_test.go

build: build-templ build-tailwind build-webserver

dev:
	air -c .air.toml