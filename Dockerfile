FROM alpine

COPY build/large-uplink.amd64 /usr/bin/large-uplink

EXPOSE 8080
CMD ["/usr/bin/large-uplink"]