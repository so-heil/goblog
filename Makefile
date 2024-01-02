tidy:
	go mod tidy

install-dependencies:
	npm install
	go mod tidy

build-templ:
	templ generate

build-tailwind:
	npm run build

build-webserver:
	go build -o ./tmp/main cmd/website/main.go

run-webserver:
	go run cmd/website/main.go serve

run-static:
	go run cmd/website/main.go static

build: build-templ build-tailwind build-webserver

start: build run-webserver

static: build-tailwind run-static

dev:
	air -c .air.toml

test:
	go test -v ./...
