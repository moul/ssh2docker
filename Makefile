PACKAGES := .
COMMANDS := $(addprefix ./,$(wildcard cmd/*))
VERSION := $(shell cat .goxc.json | jq -c .PackageVersion | sed 's/"//g')
SOURCES := $(shell find . -name "*.go")

all: build


build: $(notdir $(COMMANDS))

run: build
	./ssh2docker -V --local-user=local-user

$(notdir $(COMMANDS)): $(SOURCES)
	go get -t ./...
	gofmt -w $(PACKAGES) ./cmd/$@
	go test -i $(PACKAGES) ./cmd/$@
	go build -o $@ ./cmd/$@


test:
	go get -t ./...
	go test -i $(PACKAGES) $(COMMANDS)
	go test -v $(PACKAGES) $(COMMANDS)


install:
	go install $(COMMANDS)


cover:
	find . -name profile.out -delete
	for package in $(PACKAGES); do \
	  rm -f $$package/profile.out; \
	  go test -covermode=count -coverpkg=. -coverprofile=$$package/profile.out $$package; \
	done
	echo "mode: count" > profile.out.tmp
	cat `find . -name profile.out` | grep -v mode: | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> profile.out.tmp
	mv profile.out.tmp profile.out


.PHONY: convey
convey:
	go get github.com/smartystreets/goconvey
	goconvey -cover -port=9031 -workDir="$(realpath .)" -depth=0


.PHONY: build-docker
build-docker: contrib/docker/.docker-container-built


contrib/docker/.docker-container-built: dist/latest/ssh2docker_latest_linux_386
	cp dist/latest/ssh2docker_latest_linux_386 contrib/docker/ssh2docker
	docker build -t moul/ssh2docker:latest contrib/docker
	docker run -it --rm moul/ssh2docker --version
	docker inspect --type=image --format="{{ .Id }}" moul/ssh2docker > $@.tmp
	mv $@.tmp $@


.PHONY: run-docker
run-docker: build-docker
	docker run -it --rm -p 2222:2222 moul/ssh2docker


dist/latest/ssh2docker_latest_linux_386: $(SOURCES)
	mkdir -p dist
	rm -f dist/latest
	(cd dist; ln -s $(VERSION) latest)
	goxc -bc="linux,386" xc
	cp dist/latest/ssh2docker_$(VERSION)_linux_386 dist/latest/ssh2docker_latest_linux_386


.PHONY: docker-ps
docker-ps:
	@# consider run 'make $@' inside a unix' watch
	@#   i.e:   'watch make $@'
	docker ps --filter=label=ssh2docker -a
