ARG FYNE_CROSS_VERSION

FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd

# Download FreeBSD base, extract libs/includes, pkg keys and configs
# and install fyne dependencies
# Based on: https://github.com/myfreeweb/docker-freebsd-cross/tree/fc70196f62c00a10eb61a45a4798e09214e0cdfd
ENV ABI="FreeBSD:12:amd64"
RUN mkdir /freebsd \
    && mkdir /etc/pkg/ \
	&& curl https://download.freebsd.org/ftp/releases/amd64/12.2-RELEASE/base.txz | \
		bsdtar -xf - -C /freebsd ./lib ./usr/lib ./usr/libdata ./usr/include ./usr/share/keys ./etc \
    && cp /freebsd/etc/pkg/FreeBSD.conf /etc/pkg/ \
    && ln -s /freebsd/usr/share/keys /usr/share/keys \
    && pkg -r /freebsd install --yes xorg-minimal glfw

# Set pkg variables required to compile
ENV PKG_CONFIG_SYSROOT_DIR=/freebsd
ENV PKG_CONFIG_LIBDIR=/freebsd/usr/libdata/pkgconfig:/freebsd/usr/local/libdata/pkgconfig

# Copy the clang wrappers
COPY ./docker/freebsd-amd64/resources/*-unknown-freebsd12-* /usr/local/bin/
