#!/bin/bash
# QNAP Display Control Installation Script (Go Version)

set -e

echo "QNAP Display Control - Installation Script"
echo "=========================================="

# Check if running on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "Error: This script is designed for Linux systems (QNAP/TrueNAS)"
    exit 1
fi

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root for I/O port access"
   exit 1
fi

echo "✓ Running on Linux as root"

# Check for Go (if building from source)
if command -v go &> /dev/null; then
    echo "✓ Go found: $(go version)"
    GO_AVAILABLE=true
else
    echo "⚠ Go not found - will skip source build option"
    GO_AVAILABLE=false
fi

# Check for Bazel (if building from source)
if command -v bazel &> /dev/null; then
    echo "✓ Bazel found: $(bazel version | head -1)"
    BAZEL_AVAILABLE=true
else
    echo "⚠ Bazel not found - will skip Bazel build option"
    BAZEL_AVAILABLE=false
fi

# Installation options
echo ""
echo "Installation Options:"
echo "1. Install pre-built binary (recommended)"
echo "2. Build from source with Bazel"
echo "3. Build from source with Go"
echo "4. Exit"

read -p "Choose an option (1-4): " choice

case $choice in
    1)
        echo "Installing pre-built binary..."
        
        # Check if binary exists in current directory
        if [[ -f "bin/qnap-display-control" ]]; then
            echo "✓ Found binary in bin/ directory"
            cp bin/qnap-display-control /usr/local/bin/
        elif [[ -f "bazel-bin/qnap-display-control_/qnap-display-control" ]]; then
            echo "✓ Found Bazel-built binary"
            cp bazel-bin/qnap-display-control_/qnap-display-control /usr/local/bin/
        else
            echo "Error: No pre-built binary found. Please build first or choose option 2/3."
            exit 1
        fi
        ;;
    2)
        if [[ "$BAZEL_AVAILABLE" != "true" ]]; then
            echo "Error: Bazel is required for this option"
            exit 1
        fi
        
        echo "Building with Bazel..."
        make build
        cp bazel-bin/qnap-display-control_/qnap-display-control /usr/local/bin/
        ;;
    3)
        if [[ "$GO_AVAILABLE" != "true" ]]; then
            echo "Error: Go is required for this option"
            exit 1
        fi
        
        echo "Building with Go..."
        make build-go
        cp bin/qnap-display-control /usr/local/bin/
        ;;
    4)
        echo "Installation cancelled"
        exit 0
        ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

# Set proper permissions
chmod +x /usr/local/bin/qnap-display-control

echo "✓ Binary installed to /usr/local/bin/"

# Check for serial port
if [[ -e "/dev/ttyS1" ]]; then
    echo "✓ Serial port /dev/ttyS1 found"
elif [[ -e "/dev/ttyUSB0" ]]; then
    echo "✓ Serial port /dev/ttyUSB0 found"
else
    echo "⚠ Warning: No expected serial ports found (/dev/ttyS1 or /dev/ttyUSB0)"
    echo "  This is normal if not running on the actual QNAP device"
fi

# Create configuration directory
mkdir -p /etc/qnap-display

# Create default configuration if it doesn't exist
if [[ ! -f "/etc/qnap-display/config.json" ]]; then
    read -p "Create optional configuration file? The application works without it (y/n): " create_config
    
    if [[ "$create_config" =~ ^[Yy]$ ]]; then
        echo "Creating default configuration..."
        cat > /etc/qnap-display/config.json << 'EOF'
{
  "serial_port": "/dev/ttyS1",
  "baud_rate": 1200,
  "usb_copy_button": {
    "enabled": true,
    "io_port": "0xa05",
    "poll_interval_ms": 100,
    "debounce_ms": 50
  },
  "display": {
    "enabled": true,
    "backlight": true,
    "contrast": 128,
    "default_message": {
      "line1": "QNAP Ready",
      "line2": "Press USB Copy"
    }
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
EOF
        echo "✓ Default configuration created at /etc/qnap-display/config.json"
    else
        echo "✓ Skipping configuration file creation - application will use defaults"
    fi
fi

# Offer to create systemd service
read -p "Create systemd service for automatic startup? (y/n): " create_service

if [[ "$create_service" =~ ^[Yy]$ ]]; then
    echo "Creating systemd service..."
    cat > /etc/systemd/system/qnap-display.service << 'EOF'
[Unit]
Description=QNAP Display Controller
After=network.target
Wants=network.target

[Service]
Type=forking
ExecStart=/usr/local/bin/qnap-display-control --daemon --config /etc/qnap-display/config.json
Restart=always
RestartSec=5
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable qnap-display.service
    echo "✓ Systemd service created and enabled"
    
    read -p "Start the service now? (y/n): " start_service
    if [[ "$start_service" =~ ^[Yy]$ ]]; then
        systemctl start qnap-display.service
        echo "✓ Service started"
        systemctl status qnap-display.service --no-pager
    fi
fi

echo ""
echo "Installation completed successfully!"
echo ""
echo "Usage:"
echo "  sudo qnap-display-control --help"
echo "  sudo qnap-display-control --daemon --verbose"
echo ""
echo "Configuration file: /etc/qnap-display/config.json"
echo "Service control: systemctl {start|stop|status} qnap-display.service"
