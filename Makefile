
TARGET_BRANCH?=main
# TARGET_BRANCH=feature-branch make dep-update-oss

all: dep-update-oss build

dep-update-oss:
	cd git-sensor && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd chart-sync && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd ci-runner && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd kubelink && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd kubewatch && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd lens && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd image-scanner && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss

build:
	cd chart-sync && $(MAKE)
	cd ci-runner && $(MAKE)
#	cd devtctl && $(MAKE)
	cd git-sensor && $(MAKE)
	cd kubelink && $(MAKE)
	cd kubewatch && $(MAKE)
	cd lens && $(MAKE)
	cd image-scanner && $(MAKE)
#	cd common-lib && $(MAKE)