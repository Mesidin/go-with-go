.PHONY: run test dist app install-desktop fmt

run:
	go run ./cmd/go-with-go

test:
	go test ./...

fmt:
	gofmt -w .

dist:
	./scripts/dist.sh

app:
	./scripts/macos-app.sh

install-desktop:
	./scripts/install-desktop.sh
