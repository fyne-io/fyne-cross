tag := $(shell date +"%y.%m.%d")

build-images:
	@docker build -f ${CURDIR}/docker/base/Dockerfile -t fyneio/fyne-cross:base-latest .
	@docker tag fyneio/fyne-cross:base-latest fyneio/fyne-cross:base-$(tag)
	@docker build -f ${CURDIR}/docker/linux-386/Dockerfile -t fyneio/fyne-cross:linux-386-latest .
	@docker tag fyneio/fyne-cross:linux-386-latest fyneio/fyne-cross:linux-386-$(tag)
	@docker build -f ${CURDIR}/docker/linux-arm/Dockerfile -t fyneio/fyne-cross:linux-arm-latest .
	@docker tag fyneio/fyne-cross:linux-arm-latest fyneio/fyne-cross:linux-arm-$(tag)
	@docker build -f ${CURDIR}/docker/linux-arm64/Dockerfile -t fyneio/fyne-cross:linux-arm64-latest .
	@docker tag fyneio/fyne-cross:linux-arm64-latest fyneio/fyne-cross:linux-arm64-$(tag)
	@docker build -f ${CURDIR}/docker/android/Dockerfile -t fyneio/fyne-cross:android-latest .
	@docker tag fyneio/fyne-cross:android-latest fyneio/fyne-cross:android-$(tag)
	@docker build -f ${CURDIR}/docker/freebsd/Dockerfile -t fyneio/fyne-cross:freebsd-latest .
	@docker tag fyneio/fyne-cross:freebsd-latest fyneio/fyne-cross:freebsd-$(tag)

push-images:
	$(eval TAG := $(date +"%y.%m.%d"))
	@docker push fyneio/fyne-cross:base-latest
	@docker push fyneio/fyne-cross:base-$(tag)
	@docker push fyneio/fyne-cross:linux-386-latest
	@docker push fyneio/fyne-cross:linux-386-$(tag)
	@docker push fyneio/fyne-cross:linux-arm-latest
	@docker push fyneio/fyne-cross:linux-arm-$(tag)
	@docker push fyneio/fyne-cross:linux-arm64-latest
	@docker push fyneio/fyne-cross:linux-arm64-$(tag)
	@docker push fyneio/fyne-cross:android-latest
	@docker push fyneio/fyne-cross:android-$(tag)
	@docker push fyneio/fyne-cross:freebsd-latest
	@docker push fyneio/fyne-cross:freebsd-$(tag)
