SHELL := /bin/bash
cur-dir := $(shell pwd)
base  := $(shell basename $(cur-dir))

default:
	go get -d -v
	go build -v

docker:
	docker run --rm -v "$(cur-dir)":/usr/src/$(base) -w /usr/src/$(base) golang:1.6 bash -c make
