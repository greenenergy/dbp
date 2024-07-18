FROM alpine:3.14

WORKDIR /root
RUN apk add --no-cache curl bind-tools postgresql-client
COPY ./dbp /usr/bin/

