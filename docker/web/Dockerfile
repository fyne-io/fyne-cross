ARG FYNE_CROSS_VERSION

ARG GOPHERJS_GO_VERSION=1.17.9

# fyne-cross image for web
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base
ARG GOPHERJS_GO_VERSION

# Install the GopherJS CLI tool
RUN go install -ldflags="-w -s" -v "github.com/gopherjs/gopherjs@latest"; \
    mv /go/bin/gopherjs /usr/local/bin/gopherjs; \
    # Install the Go version compatible with the GopherJS
    go install golang.org/dl/go${GOPHERJS_GO_VERSION}@latest; \ 
    go${GOPHERJS_GO_VERSION} download; \
    rm /root/sdk/go${GOPHERJS_GO_VERSION}/*.tar.gz; \
    mv /root/sdk/go${GOPHERJS_GO_VERSION} /usr/local; \
    rm -rf /go/pkg

ENV GOPHERJS_GOROOT="/usr/local/go${GOPHERJS_GO_VERSION}"
