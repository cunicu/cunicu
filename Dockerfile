FROM golang:1.18-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN apk add \
    git \
    make \
    protoc

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

ENV CGO_ENABLED=0

COPY . .
RUN make

FROM alpine:3.16

COPY --from=builder /app/wice /

CMD [ "/wice" ]
