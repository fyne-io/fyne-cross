ARG FYNE_CROSS_VERSION
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm as pkg

# Build pkg for linux
# Based on https://github.com/freebsd/pkg/blob/release-1.12/.cirrus.yml#L19-L29
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
        build-essential \
        zlib1g-dev \
        libbz2-dev \
        m4 \
        libexpat1-dev \
        liblz4-dev \
        libxml2-dev \
        libzstd-dev \
        bsdtar \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

RUN mkdir -p /pkg-src \
    && mkdir -p /pkg/etc \
    && curl -L https://github.com/freebsd/pkg/archive/1.12.0.tar.gz | bsdtar -xf - -C /pkg-src \
    && cd /pkg-src/pkg-* \ 
    && ./scripts/install_deps.sh \
    && ./configure --with-libarchive.pc --prefix=/pkg \
    && make -j4 || make V=1 \
    && make install \
    && rm -rf /pkg-src

FROM pkg as base-freebsd

COPY --from=pkg /pkg /pkg

ENV PATH=/pkg/sbin:${PATH}
