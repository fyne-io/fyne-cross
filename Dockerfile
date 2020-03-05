# docker cross 1.13.8
ARG DOCKER_CROSS_VERSION=sha256:c7b8b09c766cf483682e53e42f16e3cde33c50cdb24585b6f0af10108bdf1b94 
# fyne develop branch
ARG FYNE_VERSION=4962907a273370d8c75173a9220aa2f64bbca162

# Build the fyne command utility
FROM dockercore/golang-cross@${DOCKER_CROSS_VERSION} AS fyne
ARG FYNE_VERSION
RUN GO111MODULE=on go get -v "fyne.io/fyne/cmd/fyne@${FYNE_VERSION}"

# Build the gowindres command utility
FROM dockercore/golang-cross@${DOCKER_CROSS_VERSION} AS gowindres
WORKDIR /app
COPY . .
RUN GO111MODULE=on go build -o /go/bin/gowindres -ldflags '-w -s' ./cmd/gowindres

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
        # headers needed by xorg-dev
        x11proto-dev \ 
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

RUN dpkg --add-architecture armhf \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \     
        gccgo-arm-linux-gnueabihf \
        libgl1-mesa-dev:armhf \
        libegl1-mesa-dev:armhf \
        libdmx-dev:armhf \
        libfontenc-dev:armhf \
        libfs-dev:armhf \
        libice-dev:armhf \
        libsm-dev:armhf \
        libx11-dev:armhf \
        libxau-dev:armhf \
        libxaw7-dev:armhf \
        libxcomposite-dev:armhf \
        libxcursor-dev:armhf \
        libxdamage-dev:armhf \
        libxdmcp-dev:armhf \
        libxext-dev:armhf \
        libxfixes-dev:armhf \
        libxfont-dev:armhf \
        libxft-dev:armhf \
        libxi-dev:armhf \
        libxinerama-dev:armhf \
        libxkbfile-dev:armhf \
        libxmu-dev:armhf \
        libxmuu-dev:armhf \
        libxpm-dev:armhf \
        libxrandr-dev:armhf \
        libxrender-dev:armhf \
        libxres-dev:armhf \
        libxss-dev:armhf \
        libxt-dev:armhf \
        libxtst-dev:armhf \
        libxv-dev:armhf \
        libxvmc-dev:armhf \
        libxxf86dga-dev:armhf \
        libxxf86vm-dev:armhf \
        xserver-xorg-dev:armhf \
        xtrans-dev:armhf \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

RUN dpkg --add-architecture arm64 \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        gccgo-aarch64-linux-gnu \
        libgl1-mesa-dev:arm64 \
        libegl1-mesa-dev:arm64 \
        libdmx-dev:arm64 \
        libfontenc-dev:arm64 \
        libfs-dev:arm64 \
        libice-dev:arm64 \
        libsm-dev:arm64 \
        libx11-dev:arm64 \
        libxau-dev:arm64 \
        libxaw7-dev:arm64 \
        libxcomposite-dev:arm64 \
        libxcursor-dev:arm64 \
        libxdamage-dev:arm64 \
        libxdmcp-dev:arm64 \
        libxext-dev:arm64 \
        libxfixes-dev:arm64 \
        libxfont-dev:arm64 \
        libxft-dev:arm64 \
        libxi-dev:arm64 \
        libxinerama-dev:arm64 \
        libxkbfile-dev:arm64 \
        libxmu-dev:arm64 \
        libxmuu-dev:arm64 \
        libxpm-dev:arm64 \
        libxrandr-dev:arm64 \
        libxrender-dev:arm64 \
        libxres-dev:arm64 \
        libxss-dev:arm64 \
        libxt-dev:arm64 \
        libxtst-dev:arm64 \
        libxv-dev:arm64 \
        libxvmc-dev:arm64 \
        libxxf86dga-dev:arm64 \
        libxxf86vm-dev:arm64 \
        xserver-xorg-dev:arm64 \
        xtrans-dev:arm64 \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

RUN dpkg --add-architecture i386 \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \     
        gccgo-i686-linux-gnu \
        libgl1-mesa-dev:i386 \
        libegl1-mesa-dev:i386 \
        libdmx-dev:i386 \
        libfontenc-dev:i386 \
        libfs-dev:i386 \
        libice-dev:i386 \
        libsm-dev:i386 \
        libx11-dev:i386 \
        libxau-dev:i386 \
        libxaw7-dev:i386 \
        libxcomposite-dev:i386 \
        libxcursor-dev:i386 \
        libxdamage-dev:i386 \
        libxdmcp-dev:i386 \
        libxext-dev:i386 \
        libxfixes-dev:i386 \
        libxfont-dev:i386 \
        libxft-dev:i386 \
        libxi-dev:i386 \
        libxinerama-dev:i386 \
        libxkbfile-dev:i386 \
        libxmu-dev:i386 \
        libxmuu-dev:i386 \
        libxpm-dev:i386 \
        libxrandr-dev:i386 \
        libxrender-dev:i386 \
        libxres-dev:i386 \
        libxss-dev:i386 \
        libxt-dev:i386 \
        libxtst-dev:i386 \
        libxv-dev:i386 \
        libxvmc-dev:i386 \
        libxxf86dga-dev:i386 \
        libxxf86vm-dev:i386 \
        xserver-xorg-dev:i386 \
        xtrans-dev:i386 \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

COPY docker-entrypoint.sh /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/docker-entrypoint.sh"]