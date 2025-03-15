# build/Dockerfile
FROM alpine:3.14

WORKDIR /root

RUN apk add --no-cache curl bind-tools postgresql-client

# Set the target architecture as a build argument
ARG TARGETARCH

# Print the architecture during build (for debugging)
RUN echo "Building for architecture: ${TARGETARCH}"

# Copy the binary from the architecture-specific folder
# Since Dockerfile is in build/, use relative path from there
COPY ${TARGETARCH}/dbp /usr/bin/dbp

# Set executable permissions
RUN chmod +x /usr/bin/dbp

ENTRYPOINT ["/usr/bin/dbp"]

