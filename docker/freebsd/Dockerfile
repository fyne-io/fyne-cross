FROM lucor/fyne-cross:base-latest

# Install pkg
# Based on https://github.com/freebsd/pkg/blob/release-1.12/.cirrus.yml#L19-L29
# TODO: split run time and build time deps to save space
RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        libsqlite3-dev \ 
        libbsd-dev \
        libarchive-dev \
        libssl-dev \
        liblzma-dev \
        liblua5.2-dev \
        nettle-dev \
        liblzo2-dev \
        libattr1-dev \
        libacl1-dev \
        wget \
        build-essential \
        zlib1g-dev \
        libbz2-dev \
        m4 \
        libexpat1-dev \
        liblz4-dev \
        libxml2-dev \
        libzstd-dev \
        bsdtar \
    && mkdir /pkg \
    && curl -L https://github.com/freebsd/pkg/archive/1.12.0.tar.gz | bsdtar -xf - -C /pkg \
    && cd /pkg/pkg-* \ 
    && ./scripts/install_deps.sh \
    && ./configure --with-libarchive.pc \
    && make -j4 || make V=1 \
    && make install \
    && rm -rf /pkg \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

# Download FreeBSD base, extract libs/includes, pkg keys and configs
# and install fyne dependencies
# Based on: https://github.com/myfreeweb/docker-freebsd-cross/tree/fc70196f62c00a10eb61a45a4798e09214e0cdfd
ENV ABI="FreeBSD:11:amd64"
RUN mkdir /freebsd \
    && mkdir /etc/pkg/ \
	&& curl https://download.freebsd.org/ftp/releases/amd64/11.2-RELEASE/base.txz | \
		bsdtar -xf - -C /freebsd ./lib ./usr/lib ./usr/libdata ./usr/include ./usr/share/keys ./etc \
    && cp /freebsd/etc/pkg/FreeBSD.conf /etc/pkg/ \
    && ln -s /freebsd/usr/share/keys /usr/share/keys \
    && pkg -r /freebsd install --yes xorg-minimal glfw

# Set pkg variables required to compile
ENV PKG_CONFIG_SYSROOT_DIR=/freebsd
ENV PKG_CONFIG_LIBDIR=/freebsd/usr/libdata/pkgconfig:/freebsd/usr/local/libdata/pkgconfig

# Copy the clang wrappers
COPY ./docker/freebsd/resources/x86_64-unknown-freebsd11-* /usr/local/bin/

RUN pkg -r /freebsd install --yes xorg
