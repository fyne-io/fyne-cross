FROM dockercore/golang-cross:1.12.1

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        libgl1-mesa-dev \
        xorg-dev \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;
