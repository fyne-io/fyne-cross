FROM dockercore/golang-cross:1.12.9

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        libgl1-mesa-dev \
        libegl1-mesa-dev \
        xorg-dev \
        gosu \
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
        x11proto-dev:armhf \
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
        x11proto-dev:arm64 \
        xserver-xorg-dev:arm64 \
        xtrans-dev:arm64 \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

COPY docker-entrypoint.sh /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/docker-entrypoint.sh"]