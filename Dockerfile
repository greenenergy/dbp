FROM alpine:3.14

WORKDIR /root
RUN apk add --no-cache curl bind-tools postgresql-client

# Set the target architecture as a build argument
ARG TARGETARCH

# Copy the appropriate binary based on the architecture
COPY ./build/${TARGETARCH}/dbp /usr/bin/dbp

