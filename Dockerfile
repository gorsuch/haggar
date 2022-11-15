FROM alpine:3.16 as alpine
RUN apk add --update --no-cache ca-certificates

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=alpine /etc/passwd /etc/passwd
COPY bin/haggar-linux-amd64 /go/bin/haggar
USER nobody
ENTRYPOINT ["/go/bin/haggar"]