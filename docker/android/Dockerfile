ARG FYNE_CROSS_VERSION
FROM fyneio/fyne-cross:${FYNE_CROSS_VERSION}-base

ENV JAVA_HOME /usr/local/android_jdk8
ENV ANDROID_HOME /usr/local/android_sdk
ENV COMMANDLINETOOLS_VERSION 7583922
ENV COMMANDLINETOOLS_SHA256SUM 124f2d5115eee365df6cf3228ffbca6fc3911d16f8025bebd5b1c6e2fcfa7faf
ENV ANDROID_SDK_BUILD_TOOLS_VERSION 30.0.3
ENV ANDROID_SDK_BUILD_TOOLS_BIN ${ANDROID_HOME}/build-tools/${ANDROID_SDK_BUILD_TOOLS_VERSION}
ENV ANDROID_SDK_PLATFORM 30
ENV ANDROID_NDK_BIN ${ANDROID_HOME}/ndk-bundle/toolchains/llvm/prebuilt/linux-x86_64/bin

ENV PATH=${PATH}:${JAVA_HOME}/bin:${ANDROID_NDK_BIN}:${ANDROID_SDK_BUILD_TOOLS_BIN}

# Install Java
RUN wget -O jdk8.tgz "https://android.googlesource.com/platform/prebuilts/jdk/jdk8/+archive/refs/heads/master/linux-x86.tar.gz"; \
	mkdir -p ${JAVA_HOME}; \
	tar zxvf jdk8.tgz -C ${JAVA_HOME}; \
	rm jdk8.tgz

# Install command line tools
RUN wget -O sdk.zip "https://dl.google.com/android/repository/commandlinetools-linux-${COMMANDLINETOOLS_VERSION}_latest.zip"; \
	echo "${COMMANDLINETOOLS_SHA256SUM} *sdk.zip" | sha256sum -c -; \
	unzip -d /tmp sdk.zip; \
	mkdir -p ${ANDROID_HOME}/cmdline-tools; \
	mv /tmp/cmdline-tools ${ANDROID_HOME}/cmdline-tools/latest; \
	rm sdk.zip;

# Install tools, platforms and ndk
RUN yes | ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager \
	"build-tools;${ANDROID_SDK_BUILD_TOOLS_VERSION}" \
	"ndk-bundle" \
	"platforms;android-${ANDROID_SDK_PLATFORM}" \
	"platform-tools"
