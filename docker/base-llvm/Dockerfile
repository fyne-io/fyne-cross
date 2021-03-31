ARG LLVM_VERSION=12
ARG FYNE_CROSS_VERSION

FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base
ARG LLVM_VERSION

RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add - \
    && echo "deb http://apt.llvm.org/buster/ llvm-toolchain-buster-${LLVM_VERSION} main" | tee /etc/apt/sources.list.d/llvm.list > /dev/null \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        clang-${LLVM_VERSION} \
        llvm-${LLVM_VERSION} \
        lld-${LLVM_VERSION}  \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

ENV PATH=/usr/lib/llvm-${LLVM_VERSION}/bin:${PATH}
