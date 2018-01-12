FROM golang:1.9.2-alpine3.7 as builder
ENV GOBIN=$GOPATH/bin
COPY . /go
RUN apk add --no-cache git ;\
  go get; \
  go build -o bin/balance-exporter -a -tags netgo -installsuffix netgo

FROM scratch
COPY --from=builder /go/bin/balance-exporter /
EXPOSE 9913/tcp
ENTRYPOINT ["/balance_exporter"]