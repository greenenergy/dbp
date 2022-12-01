FROM postgres:15rc2
#docker pull postgres:alpine3.16
RUN apt-get update

COPY initdb.sh /docker-entrypoint-initdb.d/initdb.sh
