# Quick Start Guide

## 🚀 Fast Setup (Recommended)

### Option 1: Using Make (Automated)
```bash
# Clone the repository
git clone <repository-url>
cd display

# Complete setup (installs everything)
make setup

# Build and test
make build
make test

# Run the program (requires sudo for hardware access)
sudo make run
```

### Option 2: Using Installation Script
```bash
# Run the installation script
chmod +x install_rust.sh
./install_rust.sh

# Follow the prompts for automated setup
```

## 📋 Manual Setup

### Prerequisites
- Linux system (Ubuntu/Debian/RHEL/Arch)
- Internet connection
- sudo/root access

### Step 1: Install Dependencies
```bash
# Install system dependencies
make install-deps

# Or manually:
# Ubuntu/Debian:
sudo apt-get install build-essential curl wget git pkg-config

# RHEL/CentOS:
sudo yum groupinstall "Development Tools"
sudo yum install curl wget git pkgconfig
```

### Step 2: Install Bazelisk
```bash
# Automated:
make install-bazelisk

# Or manually:
curl -fsSL https://github.com/bazelbuild/bazelisk/releases/download/v1.19.0/bazelisk-linux-amd64 -o bazelisk
chmod +x bazelisk
sudo mv bazelisk /usr/local/bin/bazel
```

### Step 3: Install Rust
```bash
# Automated:
make install-rust

# Or manually:
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain 1.75.0
source ~/.cargo/env
```

### Step 4: Verify and Build
```bash
# Verify all dependencies
make verify-deps

# Build the project
make build

# Run tests
make test
```

## 🔧 Development Workflow

### Common Commands
```bash
# Show all available commands
make help

# Quick development build
make dev

# Format and lint code
make format
make lint

# Run all quality checks
make check

# Clean build artifacts
make clean
```

### Testing
```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration
```

### Building Different Variants
```bash
# Debug build (with symbols)
make build-debug

# Optimized release build
make build-opt

# Static binary for deployment
make build-static
```

## 🎯 Running the Application

### Main Program
```bash
# Standard run (with USB monitoring)
sudo make run

# Run without USB monitoring
sudo ./bin/qnap_display_control --no-usb-monitoring
```

### USB COPY Button Testing
```bash
# Test USB COPY button only
sudo make run-usb-test

# Run USB copy example
sudo make run-example
```

## 🐳 Docker Development

### Build Docker Image
```bash
# Build development environment
make docker

# Or manually:
docker build --target development -t qnap-display:dev .
docker build --target runtime -t qnap-display:latest .
```

### Run in Container
```bash
# Development environment
docker run -it --privileged -v /dev:/dev qnap-display:dev

# Runtime environment
docker run --privileged -v /dev:/dev qnap-display:latest
```

## 📦 Deployment

### Create Deployment Package
```bash
# Create release package
make package

# Deploy to local system
sudo make deploy-local

# Full release build
make release
```

### TrueNAS Deployment
```bash
# Build static binary
make build-static

# Copy to TrueNAS
scp bin/qnap_display_control user@truenas:/usr/local/bin/
ssh user@truenas "sudo chmod +x /usr/local/bin/qnap_display_control"
```

## 🔍 Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   # Solution: Run with sudo
   sudo make run
   ```

2. **Serial Port Not Found**
   ```bash
   # Check if port exists
   ls -l /dev/ttyS1
   
   # If missing, check hardware connection
   dmesg | grep ttyS
   ```

3. **Bazel/Bazelisk Not Found**
   ```bash
   # Install Bazelisk
   make install-bazelisk
   
   # Verify installation
   make verify-deps
   ```

4. **Build Failures**
   ```bash
   # Clean and rebuild
   make clean
   make build
   
   # Check system info
   make info
   ```

### Verification Commands
```bash
# Verify complete setup
make verify

# Show build information
make info

# Check workspace status
make workspace-status

# Show dependency graph
make deps-graph
```

## 📚 Additional Resources

- **Full Documentation**: `README_RUST.md`
- **Bzlmod Migration**: `BZLMOD_MIGRATION.md`
- **Makefile Help**: `make help`
- **Project Info**: `make info`

## 🆘 Getting Help

1. **Check the help**: `make help`
2. **Verify setup**: `make verify`
3. **Check logs**: Look for error messages during build/run
4. **Hardware access**: Ensure you're running with `sudo` for hardware operations
