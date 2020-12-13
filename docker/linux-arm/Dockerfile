# fyne-cross linux arm image
FROM fyneio/fyne-cross:base-latest
RUN dpkg --add-architecture armhf \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \     
        gccgo-arm-linux-gnueabihf \
        libgl1-mesa-dev:armhf \
        libegl1-mesa-dev:armhf \
        libgles2-mesa-dev:armhf \
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
        # deps to support wayland
        libxkbcommon-dev:armhf \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;
