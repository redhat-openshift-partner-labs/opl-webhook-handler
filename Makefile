ifndef BIN_NAME
  override BIN_NAME = $(shell basename "$(PWD)")
endif

# Build go binary
binbuild:
	docker run --rm -v $(PWD):/usr/src/$(BIN_NAME) -w /usr/src/$(BIN_NAME) -e GOOS=linux -e GOARCH=amd64 golang:alpine go build -o build/$(BIN_NAME) -v

# Build container image
imgbuild:

# Push container image to remote repositor
imgpush:
