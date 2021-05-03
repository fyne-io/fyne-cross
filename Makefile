tag := $(shell date +"%y.%m.%d")

FYNE_CROSS_VERSION := "1.1"

base:
	@docker build -f ${CURDIR}/docker/base/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base .

base-llvm: base
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/base-llvm/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm .

android: base
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/android/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android .

darwin: base-llvm
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/darwin/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-darwin .

base-freebsd: base-llvm
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/base-freebsd/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd .

freebsd: base-freebsd
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/freebsd-amd64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64 .
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/freebsd-arm64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64 .

linux: base
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-386/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386 .
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-arm/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm .
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/linux-arm64/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64 .

windows: base
	@docker build --build-arg FYNE_CROSS_VERSION=${FYNE_CROSS_VERSION} -f ${CURDIR}/docker/windows/Dockerfile -t fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows .

# build all images for release. Note do not build darwin
build-images: base android freebsd linux windows

# tag the images with the YY.MM.DD suffix
tag-images:
	# tag base images
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-$(tag)
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm-$(tag)
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd-$(tag)
	# tag android images
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android-$(tag)
	# tag linux images
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386-$(tag)
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm-$(tag)
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64-$(tag)
	# tag freebsd images
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64-$(tag)
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64 fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64-$(tag)
	# tag windows images
	@docker tag fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows-$(tag)

# push the latest images
push-latest-images:
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows

# push the tagged images
push-tag-images:
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-llvm-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base-freebsd-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-android-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-386-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-linux-arm64-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-amd64-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-freebsd-arm64-$(tag)
	@docker push fyneio/fyne-cross:${FYNE_CROSS_VERSION}-windows-$(tag)

# push all images: latest and tagged 
push-images: push-latest-images push-tag-images

embed:
	go run internal/cmd/embed/main.go
