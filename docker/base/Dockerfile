# docker cross 1.13.12
ARG DOCKER_CROSS_VERSION=sha256:ad1b962c4db7a04d7bad5e676589c58c9532feecdb24eccaa8090de878e4c891
# fyne stable branch
ARG FYNE_VERSION=v1.3.2

# Build the fyne command utility
FROM dockercore/golang-cross@${DOCKER_CROSS_VERSION} AS fyne
ARG FYNE_VERSION
RUN GO111MODULE=on go get -ldflags="-w -s" -v "fyne.io/fyne/cmd/fyne@${FYNE_VERSION}"

# Build the gowindres command utility
FROM dockercore/golang-cross@${DOCKER_CROSS_VERSION} AS gowindres
WORKDIR /app
COPY . .
RUN GO111MODULE=on go build -o /go/bin/gowindres -ldflags="-w -s" ./cmd/gowindres

# Build the fyne-cross base image
FROM dockercore/golang-cross@${DOCKER_CROSS_VERSION}

COPY --from=fyne /go/bin/fyne /usr/local/bin
COPY --from=gowindres /go/bin/gowindres /usr/local/bin

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        libgl1-mesa-dev \
        libegl1-mesa-dev \
        xorg-dev \
        gosu \
        zip \
        unzip \
        # headers needed by xorg-dev
        x11proto-dev \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

COPY ./docker/base/docker-entrypoint.sh /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/docker-entrypoint.sh"]
