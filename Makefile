tag := $(shell date +"%y.%m.%d")

base:
	@docker build -f ${CURDIR}/docker/base/Dockerfile -t fyneio/fyne-cross:base-latest .

android: base
	@docker build -f ${CURDIR}/docker/android/Dockerfile -t fyneio/fyne-cross:android-latest .

darwin: base
	@docker build -f ${CURDIR}/docker/darwin/Dockerfile -t fyneio/fyne-cross:darwin-latest .

freebsd: base
	@docker build -f ${CURDIR}/docker/freebsd/Dockerfile -t fyneio/fyne-cross:freebsd-latest .

linux: base
	@docker build -f ${CURDIR}/docker/linux-386/Dockerfile -t fyneio/fyne-cross:linux-386-latest .
	@docker build -f ${CURDIR}/docker/linux-arm/Dockerfile -t fyneio/fyne-cross:linux-arm-latest .
	@docker build -f ${CURDIR}/docker/linux-arm64/Dockerfile -t fyneio/fyne-cross:linux-arm64-latest .

windows: base
	@docker build -f ${CURDIR}/docker/windows/Dockerfile -t fyneio/fyne-cross:windows-latest .

# build all images for release. Note do not build darwin
build-images: base android freebsd linux windows
	@docker tag fyneio/fyne-cross:freebsd-latest fyneio/fyne-cross:freebsd-$(tag)

push-images:
	$(eval TAG := $(date +"%y.%m.%d"))
	@docker push fyneio/fyne-cross:base-latest
	@docker tag fyneio/fyne-cross:base-latest fyneio/fyne-cross:base-$(tag)
	@docker push fyneio/fyne-cross:base-$(tag)
	@docker push fyneio/fyne-cross:android-latest
	@docker tag fyneio/fyne-cross:android-latest fyneio/fyne-cross:android-$(tag)
	@docker push fyneio/fyne-cross:android-$(tag)
	@docker push fyneio/fyne-cross:linux-386-latest
	@docker tag fyneio/fyne-cross:linux-386-latest fyneio/fyne-cross:linux-386-$(tag)
	@docker push fyneio/fyne-cross:linux-386-$(tag)
	@docker push fyneio/fyne-cross:linux-arm-latest
	@docker tag fyneio/fyne-cross:linux-arm-latest fyneio/fyne-cross:linux-arm-$(tag)
	@docker push fyneio/fyne-cross:linux-arm-$(tag)
	@docker push fyneio/fyne-cross:linux-arm64-latest
	@docker tag fyneio/fyne-cross:linux-arm64-latest fyneio/fyne-cross:linux-arm64-$(tag)
	@docker push fyneio/fyne-cross:linux-arm64-$(tag)
	@docker push fyneio/fyne-cross:freebsd-latest
	@docker tag fyneio/fyne-cross:freebsd-latest fyneio/fyne-cross:freebsd-$(tag)
	@docker push fyneio/fyne-cross:freebsd-$(tag)
	@docker push fyneio/fyne-cross:windows-latest
	@docker tag fyneio/fyne-cross:windows-latest fyneio/fyne-cross:windows-$(tag)
	@docker push fyneio/fyne-cross:windows-$(tag)
