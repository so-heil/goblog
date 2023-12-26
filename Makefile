install-dependencies:
	go install github.com/a-h/templ/cmd/templ@latest
	npm install
	go mod tidy

build-templ:
	templ generate

build-tailwind:
	npm run build

build-webserver:
	go build -o ./tmp/main cmd/website/main.go

build-static:
	go run cmd/website/main.go static

test:
	go test -v business/articles/articles_test.go

build: build-templ build-tailwind build-webserver

static: build-templ build-tailwind build-static

dev:
	air -c .air.toml