FROM golang:alpine

RUN apk add git bash

RUN mkdir /workdir
RUN mkdir /app

WORKDIR /app

RUN go install github.com/niklasfasching/go-org@v1.7.0

ADD org-mode-autoformat.sh /app/

ENV PATH="${PATH}:/app"
WORKDIR /workdir
