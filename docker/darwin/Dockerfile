ARG LLVM_VERSION=12
ARG OSX_VERSION_MIN=10.12
ARG OSX_CROSS_COMMIT="035cc170338b7b252e3f13b0e3ccbf4411bffc41"

## Install latest version of llvm and clang
FROM fyneio/fyne-cross:base-latest AS darwin-base
ARG LLVM_VERSION

RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add - \
    && echo "deb http://apt.llvm.org/buster/ llvm-toolchain-buster-${LLVM_VERSION} main" | tee /etc/apt/sources.list.d/llvm.list > /dev/null \
    && apt-get update \
    && apt-get install -y -q --no-install-recommends \
        clang-${LLVM_VERSION} \
        llvm-${LLVM_VERSION} \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

ENV PATH=/usr/lib/llvm-${LLVM_VERSION}/bin:${PATH}

## Build osxcross toolchain
FROM darwin-base as osxcross
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
 && rm -rf /var/lib/apt/lists/*

COPY *.dmg /tmp/command_line_tools_for_xcode.dmg

WORKDIR "/osxcross"

RUN git clone https://github.com/tpoechtrager/osxcross.git . \
 && git checkout -q "${OSX_CROSS_COMMIT}" \
 && rm -rf ./.git

RUN ./tools/gen_sdk_package_tools_dmg.sh /tmp/command_line_tools_for_xcode.dmg

RUN mv MacOSX11*.tar.bz2 tarballs

RUN UNATTENDED=yes OSX_VERSION_MIN=${OSX_VERSION_MIN} ./build.sh


## Build darwin-latest image
FROM darwin-base

COPY --from=osxcross /osxcross/target /osxcross/target
ENV PATH=/osxcross/target/bin:$PATH
