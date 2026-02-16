export PATH := $(PATH):/usr/local/go/bin:$(HOME)/go/bin

.PHONY: build run test generate clean docker docker-run

build: generate
	go build -o sftrails .

run: build
	./sftrails

test: generate
	go test ./... -v

generate:
	templ generate

clean:
	rm -f sftrails *.db
	find . -name "*_templ.go" -delete

docker:
	podman build -t sftrails .

docker-run: docker
	podman run --rm -p 8080:8080 -v sftrails-data:/data -e DB_PATH=/data/sftrails.db sftrails
