# fyne-cross linux 386 image
FROM fyneio/fyne-cross:base-latest
RUN dpkg --add-architecture i386 \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \     
        gccgo-i686-linux-gnu \
        libgl1-mesa-dev:i386 \
        libegl1-mesa-dev:i386 \
        libgles2-mesa-dev:i386 \
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
        # deps to support wayland
        libxkbcommon-dev:i386 \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;
