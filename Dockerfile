FROM golang:1.26
ADD . /app
WORKDIR /app
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o go-lsvd-listmonk cmd/*.go
WORKDIR /
ENTRYPOINT ["/app/go-lsvd-listmonk"]