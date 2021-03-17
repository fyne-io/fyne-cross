ARG FYNE_CROSS_VERSION

## Build the gowindres CLI tool
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base AS builder

WORKDIR /app
COPY . .

RUN go build -o /go/bin/gowindres -ldflags="-w -s" ./internal/cmd/gowindres

# Build the windows-base image
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base AS windows-base

COPY --from=builder /go/bin/gowindres /usr/local/bin

RUN apt-get update \
    && apt-get install -y -q --no-install-recommends \
        gcc-mingw-w64 \
        parallel \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;
