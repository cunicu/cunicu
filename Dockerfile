FROM golang:1.16-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY wice/ ./wice/
RUN go build -o build/wice ./wice

FROM scratch

COPY --from=builder /app/build/wice /

CMD [ "/wice" ]
