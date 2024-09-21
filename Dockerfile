FROM alpine:3.14

WORKDIR /root
RUN apk add --no-cache curl bind-tools postgresql-client

# Set the target architecture as a build argument
ARG TARGETARCH

# Print the architecture to check what is being passed during the build
RUN echo "Building for architecture: ${TARGETARCH}"

# Copy the appropriate binary based on the architecture
COPY ./build/${TARGETARCH}/dbp /usr/bin/dbp

RUN chmod +x /usr/bin/dbp

ENTRYPOINT ["/usr/bin/dbp"]
