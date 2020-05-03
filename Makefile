tag := $(shell date +"%y.%m.%d")

build-images:
	@docker build -f ${CURDIR}/docker/base/Dockerfile -t lucor/fyne-cross:base-latest .
	@docker tag lucor/fyne-cross:base-latest lucor/fyne-cross:base-$(tag)
	@docker build -f ${CURDIR}/docker/linux-386/Dockerfile -t lucor/fyne-cross:linux-386-latest .
	@docker tag lucor/fyne-cross:linux-386-latest lucor/fyne-cross:linux-386-$(tag)
	@docker build -f ${CURDIR}/docker/linux-arm/Dockerfile -t lucor/fyne-cross:linux-arm-latest .
	@docker tag lucor/fyne-cross:linux-arm-latest lucor/fyne-cross:linux-arm-$(tag)
	@docker build -f ${CURDIR}/docker/linux-arm64/Dockerfile -t lucor/fyne-cross:linux-arm64-latest .
	@docker tag lucor/fyne-cross:linux-arm64-latest lucor/fyne-cross:linux-arm64-$(tag)
	@docker build -f ${CURDIR}/docker/android/Dockerfile -t lucor/fyne-cross:android-latest .
	@docker tag lucor/fyne-cross:android-latest lucor/fyne-cross:android-$(tag)

push-images:
	$(eval TAG := $(date +"%y.%m.%d"))
	@docker push lucor/fyne-cross:base-latest
	@docker push lucor/fyne-cross:base-$(tag)
	@docker push lucor/fyne-cross:linux-386-latest
	@docker push lucor/fyne-cross:linux-386-$(tag)
	@docker push lucor/fyne-cross:linux-arm-latest
	@docker push lucor/fyne-cross:linux-arm-$(tag)
	@docker push lucor/fyne-cross:linux-arm64-latest
	@docker push lucor/fyne-cross:linux-arm64-$(tag)
	@docker push lucor/fyne-cross:android-latest
	@docker push lucor/fyne-cross:android-$(tag)
