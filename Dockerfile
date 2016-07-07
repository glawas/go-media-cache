FROM ubuntu:16.04

MAINTAINER Gilberto Lawas <glawas@dvidshub.net>

ADD . /go/src/app

RUN DEBIAN_FRONTEND=noninteractive apt-get update --fix-missing && apt-get -y upgrade
RUN apt-get install -y \
apt-utils \
golang-1.6 \
git

RUN echo "US/Eastern" | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata

RUN dpkg -i /go/src/app/vips_8.3.1-1_amd64.deb
RUN apt-get build-dep -y vips

ENV GOPATH /go

RUN ln -s /usr/local/lib/libvips.so.42 /usr/lib/libvips.so.42
RUN go get github.com/daddye/vips
RUN go install app
RUN cp /go/src/app/config.json go/bin/.

ENTRYPOINT /go/bin/app

EXPOSE 80
