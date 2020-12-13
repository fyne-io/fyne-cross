# fyne-cross linux arm64 image
FROM fyneio/fyne-cross:base-latest
RUN dpkg --add-architecture arm64 \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        gccgo-aarch64-linux-gnu \
        libgl1-mesa-dev:arm64 \
        libegl1-mesa-dev:arm64 \
        libgles2-mesa-dev:arm64 \
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
        # deps to support wayland
        libxkbcommon-dev:arm64 \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;
