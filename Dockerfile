FROM alpine:3.14

WORKDIR /root
RUN apk add --no-cache curl bind-tools
ADD dbp /usr/bin/
#ENTRYPOINT ["/usr/bin/dbp", "apply"]

