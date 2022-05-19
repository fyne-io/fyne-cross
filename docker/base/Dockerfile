ARG GO_VERSION=1.18.2
# fyne stable branch
ARG FYNE_VERSION=v2.1.4

ARG FIXUID_VERSION=0.5.1

FROM golang:${GO_VERSION}-buster AS tools
ARG FYNE_VERSION

# Install the fyne CLI tool
RUN go install -ldflags="-w -s" -v "fyne.io/fyne/v2/cmd/fyne@${FYNE_VERSION}"

# Build the fyne-cross base image
FROM golang:${GO_VERSION}-buster AS base
ARG FIXUID_VERSION

COPY --from=tools /go/bin/fyne /usr/local/bin

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        gosu \
        zip \
        unzip \
        # Switch from gz to xz compression for unix packages fyne-io/fyne#1919
        xz-utils \
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

# Install fixuid see #41
RUN addgroup --gid 1000 docker; \
    adduser --uid 1000 --ingroup docker --home /home/docker --shell /bin/sh --disabled-password --gecos "" docker; \
    curl -SsL https://github.com/boxboat/fixuid/releases/download/v${FIXUID_VERSION}/fixuid-${FIXUID_VERSION}-linux-amd64.tar.gz | tar -C /usr/local/bin -xzf -; \
    chown root:root /usr/local/bin/fixuid; \
    chmod 4755 /usr/local/bin/fixuid; \
    mkdir -p /etc/fixuid; \
    printf "user: docker\ngroup: docker\n" > /etc/fixuid/config.yml
