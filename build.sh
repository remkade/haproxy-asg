#!/bin/bash

export GOOS=linux
export GOARCH=amd64

if [ ! -f ./haproxy-asg ]; then
	go get -v ...
	go build
fi

fpm -s dir \
	-t deb \
	-n haproxy-asg \
	-a amd64 \
	--maintainer kyle.leaders@gmail.com \
	--url 'https://github.com/remkade/haproxy-asg' \
	--license 'GPLv3' \
	--depends 'haproxy = 1.8' \
	--deb-systemd haproxy-asg.service \
	--version $(awk '/Version / { gsub("\"", "", $4); print $4 }' main.go) \
	./haproxy-asg=/usr/local/bin/ \
	./haproxy.cfg.tmpl=/etc/haproxy-asg/
