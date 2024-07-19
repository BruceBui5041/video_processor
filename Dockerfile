FROM golang:20
WORKDIR /usr/src/myapp

COPY . .
RUN apt update && apt-get upgrade
RUN apt install ffmpeg

VOLUME 