## -*- docker-image-name: "scaleway/ssh2docker:latest" -*-
FROM scaleway/docker:1.9.0


# Prepare rootfs for image-builder
RUN /usr/local/sbin/builder-enter


# Install Golang (see https://github.com/scaleway-community/scaleway-golang/blob/master/Dockerfile)
RUN apt-get update -qq \
 && apt-get install -y -q --no-install-recommends \
       curl gcc ca-certificates libc6-dev git mercurial \
 && apt-get clean
ENV GOARCH=arm GOOS=linux GOLANG_VERSION=1.5.1 GOROOT=/usr/local/go GOPATH=/go
ENV PATH=${PATH}:${GOROOT}/bin:${GOPATH}/bin
RUN echo "Installing Golang 1.4" \
 && cd /tmp \
 && curl -O https://storage.googleapis.com/golang/go1.4.2.src.tar.gz \
 && echo '460caac03379f746c473814a65223397e9c9a2f6 go1.4.2.src.tar.gz' | sha1sum -c \
 && tar -C /usr/local -xzf go1.4.2.src.tar.gz \
 && rm -f go1.4.2.src.tar.gz \
 && mv /usr/local/go /usr/local/go1.4.2 \
 && cd /usr/local/go1.4.2/src \
 && ./make.bash \
 && echo "Installing Golang 1.5.1 Using go1.4.2" \
 && cd /tmp \
 && curl -O https://storage.googleapis.com/golang/go1.5.1.src.tar.gz \
 && echo '0df564746d105f4180c2b576a1553ebca9d9a124 go1.5.1.src.tar.gz' | sha1sum -c \
 && tar -C /usr/local -xzf go1.5.1.src.tar.gz \
 && rm -f /tmp/go1.5.1.src.tar.gz \
 && cd /usr/local/go/src \
 && GOROOT_BOOTSTRAP=/usr/local/go1.4.2 ./make.bash --no-clean \
 && rm -rf /usr/local/go1.4.2
RUN echo "Configure environment" \
 && mkdir -p /go/src /go/bin \
 && chmod -R 777 /go \
 && echo export GOROOT=${GOROOT} > /etc/profile.d/golang.sh \
 && echo export GOPATH=${GOPATH} >> /etc/profile.d/golang.sh \
 && echo export PATH=\${PATH}:\${GOROOT}/bin:\${GOPATH}/bin >> /etc/profile.d/golang.sh


# Install ssh2docker
RUN go get github.com/moul/ssh2docker/...


# Change default ssh port to something else
RUN echo "FIXME"


# Clean rootfs from image-builder
RUN /usr/local/sbin/builder-leave
