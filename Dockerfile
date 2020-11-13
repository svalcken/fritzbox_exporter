FROM golang:1.15
RUN git clone https://gitlab.com/dekarl/fritzbox_exporter.git /go/src/gitlab.com/dekarl/fritzbox_exporter
WORKDIR /go/src/gitlab.com/dekarl/fritzbox_exporter
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -extldflags "-static"' -o fritzbox-exporter .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/gitlab.com/dekarl/fritzbox_exporter/fritzbox-exporter /app/
COPY --from=0 /go/src/gitlab.com/dekarl/fritzbox_exporter/metrics.json /app/
# FIXME hack to allow local fiddling
COPY metrics.json /app/
WORKDIR /app
ENTRYPOINT ["/app/fritzbox-exporter"]
CMD ["--listen-address","0.0.0.0:9042"]
