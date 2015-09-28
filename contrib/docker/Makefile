all: ssh2docker


.PHONY: ssh2docker
ssh2docker:
	cd ../..; goxc -bc="linux,386" xc
	cp ../../dist/*/ssh2docker_*_linux_386 ssh2docker
	docker build -t moul/ssh2docker:latest .
	docker run -it --rm moul/ssh2docker --version
