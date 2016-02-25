build: test
	go build

generate: clean $(wildcard _assets/*)
	go generate

test: generate lint vet
	go test

clean:
	rm -f assets_gen.go
	go clean

lint:
	golint

vet:
	go vet

coverage: coverage.out
	go tool cover -html=$<

.INTERMEDIATE: coverage.out
coverage.out:
	go test -tags=integration -coverprofile=$@
