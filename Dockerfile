# Multi-stage Dockerfile for QNAP Display Control - TrueNAS Edition
# ==================================================================

# Stage 1: Build environment
FROM ubuntu:22.04 AS builder

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV BAZELISK_VERSION=1.19.0

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    pkg-config \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Bazelisk
RUN curl -fsSL https://github.com/bazelbuild/bazelisk/releases/download/v${BAZELISK_VERSION}/bazelisk-linux-amd64 \
    -o /usr/local/bin/bazel && \
    chmod +x /usr/local/bin/bazel

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | \
    sh -s -- -y --default-toolchain ${RUST_VERSION}
ENV PATH="/root/.cargo/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Copy project files
COPY . .

# Build the project
RUN make build-static

# Stage 2: Runtime environment (minimal)
FROM ubuntu:22.04 AS runtime

# Install minimal runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -r qnap && useradd -r -g qnap qnap

# Copy binaries from builder stage
COPY --from=builder /workspace/bin/* /usr/local/bin/
COPY --from=builder /workspace/README_RUST.md /usr/share/doc/qnap-display/

# Set proper permissions
RUN chmod +x /usr/local/bin/qnap_display_control && \
    chmod +x /usr/local/bin/usb_copy_example

# Create volume for device access
VOLUME ["/dev"]

# Expose metadata
LABEL maintainer="QNAP Display Control" \
      version="0.1.0" \
      description="QNAP TS-670 Pro Display Control (Rust)" \
      org.opencontainers.image.source="https://github.com/robert-wenk/display"

# Default command
CMD ["/usr/local/bin/qnap_display_control", "--help"]

# Stage 3: Development environment
FROM builder AS development

# Install additional development tools
RUN cargo install cargo-watch cargo-audit

# Install debugging tools
RUN apt-get update && apt-get install -y \
    gdb \
    strace \
    ltrace \
    && rm -rf /var/lib/apt/lists/*

# Set up development environment
WORKDIR /workspace
CMD ["/bin/bash"]
