# TrueNAS QNAP Display Control App

## 🎯 Overview

This package converts the QNAP Display Control application into a TrueNAS SCALE native app that:

- **Auto-starts** when TrueNAS boots up
- **Integrates** with TrueNAS's Kubernetes infrastructure  
- **Persists** configuration across reboots and updates
- **Provides** hardware LCD display and USB copy button support
- **Includes** TrueNAS-specific menu options (ZFS pools, datasets, services)

## 🚀 Installation Methods

### Method 1: Helm Chart (Recommended)

Full-featured installation with advanced configuration options:

```bash
# Clone repository
git clone https://github.com/robert-wenk/display.git
cd display

# Run installation script
sudo ./install-truenas.sh --build
```

### Method 2: Simple Kubernetes Deployment

Direct YAML deployment (no Helm required):

```bash
# Clone repository 
git clone https://github.com/robert-wenk/display.git
cd display

# Deploy with simple script
sudo ./deploy-simple.sh --build
```

### Method 3: Manual Deployment

For advanced users who want full control:

```bash
# Build Docker image
docker build -t qnap-display-control:latest .

# Apply Kubernetes manifests
kubectl apply -f truenas-deployment.yaml
```

## 📁 Package Structure

```
display/
├── truenas-app/              # Helm chart for TrueNAS
│   ├── app.yaml             # App metadata
│   ├── questions.yaml       # TrueNAS UI configuration
│   ├── values.yaml          # Default values
│   └── templates/           # Kubernetes templates
│       ├── deployment.yaml
│       ├── configmap.yaml
│       ├── persistentvolumeclaim.yaml
│       └── _helpers.tpl
├── truenas-deployment.yaml   # Simple Kubernetes YAML
├── install-truenas.sh       # Helm installation script
├── deploy-simple.sh         # Simple deployment script
├── uninstall-truenas.sh     # Removal script
└── Dockerfile               # Updated for TrueNAS
```

## ⚙️ Configuration

### Default Settings

| Setting | Default Value | Description |
|---------|---------------|-------------|
| **Config Storage** | `/mnt/pool0/apps/qnap-display/config` | App configuration |
| **USB Copy Storage** | `/mnt/pool0/Multimedia` | USB copy destination |
| **Serial Port** | `/dev/ttyS1` | QNAP display communication |
| **I/O Port** | `0xa05` | USB copy button |
| **Display Size** | `16x2` | Character LCD dimensions |
| **Default Text** | `"TrueNAS Ready"` | Initial display message |

### Hardware Requirements

- **Privileged Access**: Required for I/O port operations
- **Serial Port**: `/dev/ttyS1` must exist for display communication
- **I/O Port Access**: `/dev/port` required for USB copy button
- **USB Mount Point**: `/media/usb` for USB device detection

## 🎛️ Features

### Auto-Start Integration
- **Systemd Service**: Automatically starts on TrueNAS boot
- **Kubernetes Integration**: Uses TrueNAS's native container platform
- **Persistent Configuration**: Survives system updates and reboots
- **Health Monitoring**: Automatic restart on failure

### TrueNAS-Specific Enhancements
- **ZFS Integration**: Menu options for pool status and dataset management
- **Service Monitoring**: Display running TrueNAS services
- **Storage Information**: Real-time disk usage and mount points
- **Network Tools**: IP display and connectivity testing

### Hardware Features
- **16x2 LCD Display**: Real-time system information
- **Button Navigation**: ENTER/SELECT for menu navigation
- **USB Copy Button**: Hardware button for one-touch USB backup
- **LED Control**: Status and activity indicators
- **Horizontal Scrolling**: Long outputs scroll automatically

## 📋 Menu System

### Main Menu Structure
```
TrueNAS Control
├── System Info       → Hardware and OS details
├── Network           → IP address, connectivity tests
├── TrueNAS          → ZFS pools, datasets, services
├── Storage          → Disk usage information
└── Display          → Backlight controls
```

### Command Examples
- **System Info**: `uname -a` → Shows kernel and hardware info
- **Network IP**: `hostname -I` → Displays current IP addresses
- **ZFS Pools**: `zpool status` → Shows pool health and status
- **Datasets**: `zfs list` → Lists all ZFS datasets
- **Services**: `systemctl list-units --type=service --state=running`
- **Storage**: `df -h` → Shows filesystem usage

## 🔧 Customization

### Modify Storage Paths

Edit installation script or values.yaml:
```bash
# Custom multimedia storage location
MULTIMEDIA_DIR="/mnt/tank/USB-Backups"

# Custom config location  
CONFIG_DIR="/mnt/pool0/ix-applications/qnap-display"
```

### Add Custom Menu Items

Edit `truenas-app/templates/configmap.yaml`:
```json
"custom_command": {
  "title": "My Command",
  "description": "Custom functionality",
  "type": "command",
  "command": "echo 'Custom output'"
}
```

### Hardware Configuration

Modify hardware settings in values.yaml:
```yaml
hardwareAccess:
  serialPort:
    device: "/dev/ttyS0"    # Different serial port
  ioPortAccess:
    port: "0xa10"           # Different I/O port
  usbAccess:
    mountPath: "/mnt/usb"   # Different USB mount
```

## 🔍 Monitoring and Troubleshooting

### Check Application Status
```bash
# Pod status
kubectl get pods -n qnap-display-control

# Application logs
kubectl logs -f deployment/qnap-display-control -n qnap-display-control

# Systemd service status
systemctl status qnap-display-control
```

### Common Issues

**App Won't Start:**
```bash
# Check hardware devices
ls -la /dev/ttyS1 /dev/port

# Verify privileged mode
kubectl describe pod -n qnap-display-control

# Check image availability
docker images | grep qnap-display-control
```

**USB Copy Not Working:**
```bash
# Test I/O port access
sudo xxd -s 0xa05 -l 1 /dev/port

# Check USB mount point
ls -la /media/usb/

# Verify button detection in logs
kubectl logs deployment/qnap-display-control -n qnap-display-control | grep -i usb
```

**Display Not Responding:**
```bash
# Test serial port
ls -la /dev/ttyS*

# Check baud rate configuration
kubectl get configmap qnap-display-config -n qnap-display-control -o yaml
```

## 🗑️ Uninstallation

### Helm Installation
```bash
sudo ./uninstall-truenas.sh
```

### Simple Deployment
```bash
kubectl delete -f truenas-deployment.yaml
systemctl disable qnap-display-control.service
rm /etc/systemd/system/qnap-display-control.service
```

## 🔄 Updates

### Update Application
```bash
# Pull latest changes
git pull

# Rebuild image
docker build -t qnap-display-control:latest .

# Update deployment
kubectl rollout restart deployment/qnap-display-control -n qnap-display-control
```

### Upgrade Configuration
```bash
# For Helm installations
helm upgrade qnap-display-control ./truenas-app -n ix-qnap-display-control

# For simple deployments
kubectl apply -f truenas-deployment.yaml
```

## 💡 Benefits

✅ **Native TrueNAS Integration** - Runs as a proper TrueNAS SCALE app  
✅ **Auto-Start on Boot** - Systemd service ensures availability  
✅ **Persistent Storage** - Configuration survives reboots and updates  
✅ **Hardware Access** - Full I/O port and serial communication support  
✅ **TrueNAS-Specific Features** - ZFS and service monitoring menus  
✅ **Easy Installation** - One-command deployment with scripts  
✅ **Horizontal Scrolling** - Long outputs display properly on 16x2 LCD  
✅ **USB Copy Integration** - Hardware button triggers automated backups  
✅ **Health Monitoring** - Kubernetes ensures app restarts on failure  
✅ **Resource Management** - Proper CPU and memory limits  

This TrueNAS app package transforms the QNAP Display Control system into a fully integrated TrueNAS SCALE application with automatic startup, persistent configuration, and enhanced TrueNAS-specific functionality.
