ARG GO_VERSION=1.16.7
# fyne stable branch
ARG FYNE_VERSION=v2.0.4

FROM golang:${GO_VERSION}-buster AS tools
ARG FYNE_VERSION

# Install the fyne CLI tool
RUN go get -ldflags="-w -s" -v "fyne.io/fyne/v2/cmd/fyne@${FYNE_VERSION}"

# Build the fyne-cross base image
FROM golang:${GO_VERSION}-buster AS base

COPY --from=tools /go/bin/fyne /usr/local/bin

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        gosu \
        zip \
        unzip \
        # fyne deps
        libgl1-mesa-dev \
        libegl1-mesa-dev \
        libgles2-mesa-dev \
        xorg-dev \
        # headers needed by xorg-dev
        x11proto-dev \
        # deps to support wayland
        libxkbcommon-dev \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

COPY ./docker/base/docker-entrypoint.sh /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/docker-entrypoint.sh"]
