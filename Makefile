build-templ:
	templ generate

build-tailwind:
	npm run build

build-webserver:
	go build -o ./tmp/main cmd/webserver/main.go

build: build-templ build-tailwind build-webserver

dev:
	air -c .air.toml