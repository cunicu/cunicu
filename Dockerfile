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

COPY . .
RUN make

FROM alpine:3.15

COPY --from=builder /app/wice /

CMD [ "/wice" ]
