FROM golang:1.9.2-alpine3.7 as builder
ENV GOBIN=$GOPATH/bin
COPY . /go
RUN apk add --no-cache git ;\
  go get; \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/balance-exporter \
  -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo

FROM scratch
COPY --from=builder /go/bin/balance-exporter /
EXPOSE 9913/tcp
ENTRYPOINT ["/balance-exporter"]