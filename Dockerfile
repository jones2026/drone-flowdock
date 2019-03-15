FROM scratch
COPY drone-flowdock /bin/
ENTRYPOINT ["/bin/drone-flowdock"]
