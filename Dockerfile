FROM dockercore/golang-cross:1.12.9

RUN apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
        libgl1-mesa-dev \
        libegl1-mesa-dev \
        xorg-dev \
        gosu \
    && apt-get -qy autoremove \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*;

COPY docker-entrypoint.sh /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/docker-entrypoint.sh"]