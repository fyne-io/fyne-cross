tag := $(shell date +"%y.%m.%d")
RUNNER := $(shell 2>/dev/null 1>&2 docker version && echo "docker" || echo "podman")
REGISTRY := docker.io
FYNE_CROSS_VERSION := "1.2"


base:
	@$(RUNNER) build -f ${CURDIR}/docker/base/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base .

base-llvm: base
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/base-llvm/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm .

android: base
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/android/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android .

darwin: base-llvm
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/darwin/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-darwin .

base-freebsd: base-llvm
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/base-freebsd/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd .

freebsd: base-freebsd
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/freebsd-amd64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64 .
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/freebsd-arm64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64 .

linux: base
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-386/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386 .
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-arm/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm .
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-arm64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64 .

windows: base
	@$(RUNNER) build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/windows/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows .

# build all images for release. Note do not build darwin
build-images: base android freebsd linux windows
ifeq ($(RUNNER),podman)
	$(MAKE) podman-tag
endif

# tag the images with the YY.MM.DD suffix
tag-images:
	# tag base images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-$(tag)
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm-$(tag)
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd-$(tag)
	# tag android images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android-$(tag)
	# tag linux images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386-$(tag)
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm-$(tag)
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64-$(tag)
	# tag freebsd images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64-$(tag)
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64-$(tag)
	# tag windows images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows-$(tag)

podman-tag:
	# tag base images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd
	# tag android images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android
	# tag linux images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386 $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64 $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64
	# tag freebsd images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64 $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64 $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64
	# tag windows images
	@$(RUNNER) tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows $(REGISTRY)/fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows


# push the latest images
push-latest-images:
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows

# push the tagged images
push-tag-images:
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64-$(tag)
	@$(RUNNER) push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows-$(tag)

# push all images: latest and tagged 
push-images: push-latest-images push-tag-images

embed:
	go run internal/cmd/embed/main.go
