FROM golang:1.22
WORKDIR /usr/src/myapp

COPY . .
RUN apt update && apt-get upgrade && apt install protobuf-compiler -y
RUN chmod +x gen_proto.sh
# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN apt install ffmpeg -y

VOLUME 