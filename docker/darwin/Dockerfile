ARG LLVM_VERSION=12
ARG OSX_VERSION_MIN=10.12
ARG OSX_CROSS_COMMIT="8a716a43a72dab1db9630d7824ee0af3730cb8f9"
ARG FYNE_CROSS_VERSION=1.2

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

ARG SDK_VERSION
RUN echo "Available SDKs:" && ls -1 MacOSX*.tar.* && \
    if [ -z "$SDK_VERSION" ] ;\
     then ls -1 MacOSX*.tar.* | sort -Vr | head -1 | xargs -i mv {} tarballs ;\
     else mv MacOSX*.tar.* tarballs ; \
    fi

RUN UNATTENDED=yes SDK_VERSION=${SDK_VERSION} OSX_VERSION_MIN=${OSX_VERSION_MIN} ./build.sh


## Build darwin-latest image
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm

COPY --from=osxcross /osxcross/target /osxcross/target
ENV PATH=/osxcross/target/bin:$PATH
