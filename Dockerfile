FROM golang:20
WORKDIR /usr/src/myapp

COPY . .
RUN apt update && apt-get upgrade -y
RUN apt install ffmpeg -y

VOLUME 