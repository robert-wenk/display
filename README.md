# QNAP Display Control

A high-performance Go application for controlling LCD displays and monitoring USB copy buttons on QNAP devices, specifically designed for TrueNAS deployment.

## 🚀 Features

- **LCD Display Control**: Write text and progress indicators to HD44780-compatible displays
- **USB Copy Button Monitoring**: Real-time hardware button detection via I/O port polling
- **Interactive Menu System**: Configurable menu navigation with ENTER/SELECT buttons
- **Serial Communication**: Robust serial port management for display communication
- **Configuration Management**: JSON-based configuration with sensible defaults
- **Command Execution**: Execute system commands from the menu interface
- **Daemon Mode**: Background service operation for continuous monitoring
- **Cross-Platform Build**: Bazel and Go build system support
- **Comprehensive Testing**: Unit tests, integration tests, and benchmarks
- **Low Resource Usage**: Optimized 3.7MB binary with minimal memory footprint

## 📋 Requirements

- **Hardware**: QNAP devices with LCD display and USB copy button
- **Operating System**: Linux (TrueNAS compatible)
- **Privileges**: Root access for I/O port access
- **Dependencies**: None (statically linked binary)
- **Configuration**: None required (works out-of-the-box with `/dev/ttyS1`)

## 🔧 Installation

### Pre-built Binary

Download the latest release binary:

```bash
wget https://github.com/robert-wenk/display/releases/latest/download/qnap-display-control-linux-amd64.tar.gz
tar -xzf qnap-display-control-linux-amd64.tar.gz
sudo mv qnap-display-control /usr/local/bin/
sudo chmod +x /usr/local/bin/qnap-display-control
```

### Build from Source

#### Option 1: Bazel Build (Recommended)

```bash
# Clone repository
git clone https://github.com/robert-wenk/display.git
cd display

# Build with Bazel
make build

# Install
sudo cp bazel-bin/qnap-display-control_/qnap-display-control /usr/local/bin/
```

#### Option 2: Go Build

```bash
# Ensure Go 1.21.5+ is installed
make build-go

# Install
sudo cp bin/qnap-display-control /usr/local/bin/
```

## 🎯 Usage

### Command Line Interface

```bash
# Basic usage with default settings (no config.json required)
sudo qnap-display-control

# Run as daemon with verbose logging
sudo qnap-display-control --daemon --verbose

# Custom configuration file (optional)
sudo qnap-display-control --config /etc/qnap-display/config.json

# Custom serial port and baud rate
sudo qnap-display-control --port /dev/ttyUSB0 --baud 1200
```

### Available Flags

```
Flags:
  -b, --baud int        Serial port baud rate (default 1200)
  -c, --config string   Configuration file path (optional, default "/etc/qnap-display/config.json")
  -d, --daemon          Run as daemon
  -h, --help            help for qnap-display-control
  -p, --port string     Serial port device (default "/dev/ttyS1")
  -v, --verbose         Enable verbose logging
```

### Configuration File (Optional)

The application works without any configuration file using sensible defaults. To customize settings including the menu system, create `/etc/qnap-display/config.json`:

```json
{
  "serial_port": {
    "device": "/dev/ttyS1",
    "baud_rate": 1200,
    "timeout_ms": 1000
  },
  "usb_copy": {
    "io_port": "0xa05",
    "poll_interval_ms": 50,
    "enabled": true
  },
  "display": {
    "width": 16,
    "height": 2,
    "backlight_pin": -1,
    "contrast": 128,
    "default_text": "QNAP Ready"
  },
  "menu": {
    "enabled": true,
    "button_delay_ms": 200,
    "main_menu": {
      "title": "Main Menu",
      "description": "QNAP Control",
      "type": "submenu",
      "items": {
        "system": {
          "title": "System Info",
          "description": "Show system information",
          "type": "command",
          "command": "uname -a"
        },
        "network": {
          "title": "Network",
          "description": "Network configuration",
          "type": "submenu",
          "items": {
            "ip": {
              "title": "Show IP",
              "description": "Display IP address",
              "type": "command",
              "command": "hostname -I"
            }
          }
        }
      }
    }
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

### Menu System

The application features a comprehensive menu system that can be navigated using the LCD panel buttons:

#### Button Controls
- **SELECT Button**: Navigate through menu options (cycles through available items)
- **ENTER Button**: Select current option (execute command or enter submenu)

#### Menu Navigation
1. **Main Menu**: Shows on startup with menu description on line 1, current selection on line 2
2. **Submenus**: Navigate into submenus for organized command groups
3. **Back Navigation**: Automatically adds "Back" option in submenus
4. **Command Execution**: Shows "Executing..." then displays command results

#### Menu Configuration
- **Menu Items**: Can be either `"submenu"` or `"command"` type
- **Commands**: Shell commands executed when selected
- **Hierarchy**: Unlimited nesting of submenus
- **Customizable**: Fully configurable via JSON

See `config_example.json` for a comprehensive menu configuration example.

## 🔧 Development

### Project Structure

```
cmd/                    # CLI application entry point
├── main.go            # Main application
internal/              # Internal packages
├── config/            # Configuration management
├── controller/        # Display controller logic
├── monitor/           # USB button monitoring
├── hardware/          # I/O port access
├── serial/            # Serial communication
└── error/             # Error handling
test/                  # Test suites
├── integration/       # Integration tests
└── benchmark/         # Performance benchmarks
```

### Building and Testing

```bash
# Install dependencies
make verify-deps

# Run tests
make test

# Run linting
make lint

# Format code
make format

# Build optimized binary
make build-opt

# Build static binary for deployment
make build-static

# Create deployment package
make package
```

## 🔌 Hardware Details

### USB Copy Button Detection

- **I/O Port Address**: `0xa05` (register)
- **Detection Method**: Direct I/O port access using `ioperm()` and `inb()` syscalls
- **Button Bit**: Bit 2 in the port value (active low)
- **Polling Interval**: 100ms (configurable)
- **Debouncing**: 50ms hardware debounce protection

### LCD Display Communication

- **Protocol**: HD44780-compatible command set
- **Serial Interface**: `/dev/ttyS1` (configurable)
- **Default Baud Rate**: 1200 (configurable)
- **Display Size**: 2 lines × 16 characters
- **Features**: Text positioning, progress bars, backlight control

## 🚀 TrueNAS Deployment

### SystemD Service

Create `/etc/systemd/system/qnap-display.service`:

```ini
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
```

Enable and start:

```bash
sudo systemctl enable qnap-display.service
sudo systemctl start qnap-display.service
sudo systemctl status qnap-display.service
```

## 🐛 Troubleshooting

### Permission Issues

```bash
# Ensure I/O port access
sudo dmesg | grep -i "permission denied"

# Check device permissions
ls -la /dev/ttyUSB0

# Add user to dialout group (if running as non-root)
sudo usermod -a -G dialout $USER
```

### Serial Port Issues

```bash
# List available serial ports
ls -la /dev/tty*

# Test serial port
sudo minicom -D /dev/ttyUSB0 -b 9600

# Check for conflicts
sudo lsof /dev/ttyUSB0
```

### I/O Port Access

```bash
# Check if ioperm is available
dmesg | grep ioperm

# Verify port access in kernel logs
sudo dmesg | grep 0xa05

# Check process capabilities
sudo getcap /usr/local/bin/qnap-display-control
```

## 📊 Performance

| Metric | Value |
|--------|-------|
| Binary Size (Bazel) | 3.7 MB |
| Binary Size (Go) | 5.5 MB |
| Memory Usage | < 5 MB |
| CPU Usage | < 0.1% |
| Startup Time | < 100ms |
| Button Response | < 50ms |

## 📄 License

MIT License - See LICENSE file for details.

## 🏆 Credits

- **Hardware Research**: Based on QNAP hardware documentation and community research
- **USB Copy Button**: Implementation based on I/O port analysis
- **Display Protocol**: HD44780 compatibility layer for QNAP displays
- **Original Python Version**: Foundation research for hardware interaction
