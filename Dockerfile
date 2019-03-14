FROM alpine:3.9.2
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY drone-flowdock /bin/
ENTRYPOINT ["/bin/drone-flowdock"]
