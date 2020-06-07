FROM lucor/fyne-cross:base-latest

ENV JAVA_HOME /usr/local/android_jdk8
ENV ANDROID_HOME /usr/local/android_sdk
ENV ANDROID_SDK_TOOLS_VERSION 4333796
ENV ANDROID_SDK_TOOLS_SHA256SUM 92ffee5a1d98d856634e8b71132e8a95d96c83a63fde1099be3d86df3106def9
ENV ANDROID_SDK_BUILD_TOOLS_VERSION 29.0.2
ENV ANDROID_SDK_BUILD_TOOLS_BIN ${ANDROID_HOME}/build-tools/${ANDROID_SDK_BUILD_TOOLS_VERSION}
ENV ANDROID_SDK_PLATFORM 29
ENV ANDROID_NDK_BIN ${ANDROID_HOME}/ndk-bundle/toolchains/llvm/prebuilt/linux-x86_64/bin

ENV PATH=${PATH}:${JAVA_HOME}/bin:${ANDROID_NDK_BIN}:${ANDROID_SDK_BUILD_TOOLS_BIN}

# Install Java
RUN wget -O jdk8.tgz "https://android.googlesource.com/platform/prebuilts/jdk/jdk8/+archive/refs/heads/master/linux-x86.tar.gz"; \
	mkdir -p ${JAVA_HOME}; \
	tar zxvf jdk8.tgz -C ${JAVA_HOME}; \
	rm jdk8.tgz

# Install SDK
RUN wget -O sdk.zip "https://dl.google.com/android/repository/sdk-tools-linux-${ANDROID_SDK_TOOLS_VERSION}.zip"; \
	echo "${ANDROID_SDK_TOOLS_SHA256SUM} *sdk.zip" | sha256sum -c -; \
	unzip -d ${ANDROID_HOME} sdk.zip; \
	rm sdk.zip;

# Install tools, platforms and ndk
RUN yes | ${ANDROID_HOME}/tools/bin/sdkmanager \
	"build-tools;${ANDROID_SDK_BUILD_TOOLS_VERSION}" \
	"ndk-bundle" \
	"platforms;android-${ANDROID_SDK_PLATFORM}" \
	"platform-tools"
