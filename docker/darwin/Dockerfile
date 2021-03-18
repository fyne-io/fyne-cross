ARG LLVM_VERSION=12
ARG OSX_VERSION_MIN=10.12
ARG OSX_CROSS_COMMIT="035cc170338b7b252e3f13b0e3ccbf4411bffc41"
ARG FYNE_CROSS_VERSION=1.1

## Build osxcross toolchain
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm as osxcross
ARG OSX_CROSS_COMMIT
ARG OSX_VERSION_MIN

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends \
    bzip2 \
    cmake \ 
    cpio \
    patch \
    libbz2-dev \
    libssl-dev \
    zlib1g-dev \
    liblzma-dev \
    libxml2-dev \
    uuid-dev \
 && rm -rf /var/lib/apt/lists/*

COPY *.dmg /tmp/command_line_tools_for_xcode.dmg

WORKDIR "/osxcross"

RUN curl -L https://github.com/tpoechtrager/osxcross/archive/${OSX_CROSS_COMMIT}.tar.gz | tar -zx --strip-components=1

RUN ./tools/gen_sdk_package_tools_dmg.sh /tmp/command_line_tools_for_xcode.dmg

RUN mv MacOSX11*.tar.bz2 tarballs

RUN UNATTENDED=yes OSX_VERSION_MIN=${OSX_VERSION_MIN} ./build.sh


## Build darwin-latest image
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm

COPY --from=osxcross /osxcross/target /osxcross/target
ENV PATH=/osxcross/target/bin:$PATH
