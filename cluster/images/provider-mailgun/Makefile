# ====================================================================================
# Setup Project

PLATFORMS := linux_amd64 linux_arm64
include ../../../build/makelib/common.mk

# ====================================================================================
# Options
IMAGE ?= $(BUILD_REGISTRY)/provider-mailgun:$(VERSION)
TARGETOS := $(word 1, $(subst _, ,$(PLATFORM)))
TARGETARCH := $(word 2, $(subst _, ,$(PLATFORM)))
include ../../../build/makelib/imagelight.mk

# ====================================================================================
# Image Build
img.build: ../../../_output/bin/$(PLATFORM)/provider
	@$(INFO) docker build $(IMAGE)
	@$(MAKE) BUILD_ARGS="--load" img.build.shared
	@$(OK) docker build $(IMAGE)

../../../_output/bin/$(PLATFORM)/provider:
	@$(MAKE) -C ../../.. go.build PLATFORM=$(PLATFORM)

img.build.shared:
	@cp Dockerfile $(IMAGE_TEMP_DIR)
	@cp -r $(OUTPUT_DIR)/bin/ $(IMAGE_TEMP_DIR)/
	@docker buildx build $(BUILD_ARGS) --platform $(IMAGE_PLATFORM) --build-arg TARGETOS=$(TARGETOS) --build-arg TARGETARCH=$(TARGETARCH) -t $(IMAGE) $(IMAGE_TEMP_DIR)

# ====================================================================================
# Image Publish
img.publish:
	@$(INFO) Skipping image publish for $(IMAGE)
	@echo Publish is deferred to xpkg machinery
	@$(OK) Image publish skipped for $(IMAGE)
