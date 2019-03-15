FROM alpine:3.9.2 as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY drone-flowdock /bin/
ENTRYPOINT ["/bin/drone-flowdock"]
