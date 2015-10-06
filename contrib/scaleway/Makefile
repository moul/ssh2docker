NAME =			ssh2docker
VERSION =		latest
VERSION_ALIASES =
TITLE =			ssh2docker
DESCRIPTION =		SSH proxy to Docker containers
SOURCE_URL =		https://github.com/moul/ssh2docker
VENDOR_URL =		https://github.com/moul/ssh2docker

IMAGE_VOLUME_SIZE =     50G
IMAGE_BOOTSCRIPT =      docker
IMAGE_NAME =            ssh2docker


## Image tools  (https://github.com/scaleway/image-tools)
all:	docker-rules.mk
docker-rules.mk:
	wget -qO - http://j.mp/scw-builder | bash
-include docker-rules.mk
