FROM --platform=${TARGETPLATFORM} busybox:latest
COPY bang /usr/local/bin/bang
ENTRYPOINT ["/usr/local/bin/bang"]
