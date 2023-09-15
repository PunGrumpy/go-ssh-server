# Author: PunGrumpy
# For running a server (SSH server)

#
# Build stage
#
FROM golang:1.21.1 AS build
WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/server

#
# Runtime stage
#
FROM debian:stable-slim
WORKDIR /app
COPY --from=build /app/server /app/server
COPY --from=build /app/server_key.pem /app/server_key.pem
COPY --from=build /app/server_key.pub /app/server_key.pub
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get install --no-install-recommends -y openssh-server=9.4 && \
    mkdir /var/run/sshd && \
    echo 'root:root' | chpasswd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    chmod +x /app/server && \
    useradd -ms /bin/bash sshserver && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
EXPOSE 2023
ENTRYPOINT ["/app/server"]
