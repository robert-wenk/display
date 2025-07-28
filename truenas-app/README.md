# QNAP Display Control - TrueNAS App

This directory contains the TrueNAS SCALE application configuration for the QNAP Display Control system. This allows the application to be installed as a native TrueNAS app that automatically starts when the system boots.

## 🚀 Quick Installation

### Prerequisites

- TrueNAS SCALE system
- QNAP hardware with display and USB copy button support
- Root access to TrueNAS system
- Helm 3.x installed (usually included with TrueNAS SCALE)

### Installation Steps

1. **Clone or copy the application to your TrueNAS system:**
   ```bash
   git clone https://github.com/robert-wenk/display.git
   cd display
   ```

2. **Make installation script executable:**
   ```bash
   chmod +x install-truenas.sh uninstall-truenas.sh
   ```

3. **Run the installation script:**
   ```bash
   sudo ./install-truenas.sh --build
   ```

   The `--build` flag will build the Docker image locally. Omit it if you have a pre-built image.

4. **Verify installation:**
   ```bash
   kubectl get pods -n ix-qnap-display-control
   kubectl logs -f deployment/qnap-display-control -n ix-qnap-display-control
   ```

## 📋 Configuration

The app uses the following default configuration that can be customized during installation:

### Storage Locations
- **Config Directory**: `/mnt/pool0/apps/qnap-display/config`
- **USB Copy Destination**: `/mnt/pool0/Multimedia`

### Hardware Access
- **Serial Port**: `/dev/ttyS1` (QNAP display communication)
- **I/O Port**: `0xa05` (USB copy button)
- **USB Mount**: `/media/usb` (USB devices for copying)

### Security Settings
- **Privileged Mode**: Enabled (required for hardware I/O access)
- **Run as Root**: Enabled (required for hardware access)

## 🔧 Customization

### Modify Default Values

Edit `truenas-app/values.yaml` to customize:

```yaml
# Example customizations
config:
  display:
    defaultText: "My Custom Text"
  logging:
    level: "debug"

persistence:
  multimedia:
    hostPath: "/mnt/tank/USB-Copies"

hardwareAccess:
  serialPort:
    device: "/dev/ttyS0"  # Different serial port
```

### Advanced Configuration

Modify `truenas-app/templates/configmap.yaml` to customize the menu system, add new commands, or change USB copy behavior.

## 🎛️ Features

### Auto-Start on Boot
- Systemd service automatically starts the app when TrueNAS boots
- Integrates with TrueNAS's Kubernetes infrastructure
- Survives system reboots and updates

### Hardware Integration
- **Display Control**: 16x2 character LCD with menu system
- **Button Support**: ENTER/SELECT buttons for navigation
- **USB Copy**: Hardware button triggers USB device copying
- **LED Control**: Status and disk activity LEDs

### Menu System
- **TrueNAS Integration**: Specific menu items for ZFS pools, datasets, services
- **Network Tools**: IP display, ping tests
- **System Information**: Hardware info, storage status
- **Scrolling Output**: Long command outputs scroll horizontally

## 📱 Usage

Once installed and running:

1. **Navigation**: Use ENTER/SELECT buttons on QNAP panel
2. **Menu System**: Navigate through TrueNAS-specific options
3. **USB Copy**: Press USB copy button to backup USB devices
4. **Status Display**: View system information on LCD display

### Menu Options
- **System Info**: Hardware and OS information
- **Network**: IP address, connectivity tests
- **TrueNAS**: ZFS pools, datasets, services status
- **Storage**: Disk usage and mount points
- **Display**: Backlight control

## 🔍 Monitoring

### Check Application Status
```bash
# Pod status
kubectl get pods -n ix-qnap-display-control

# Application logs
kubectl logs -f deployment/qnap-display-control -n ix-qnap-display-control

# Service status
systemctl status qnap-display-control
```

### Debug Hardware Issues
```bash
# Check serial port
ls -la /dev/ttyS*

# Check I/O port access
sudo xxd -s 0xa05 -l 1 /dev/port

# Test USB mount detection
ls -la /media/usb/
```

## 🗑️ Uninstallation

To remove the application:

```bash
sudo ./uninstall-truenas.sh
```

This will:
- Stop and disable the systemd service
- Remove the Helm deployment
- Delete the Kubernetes namespace
- Optionally remove configuration files

## 🔧 Troubleshooting

### Common Issues

**App won't start:**
- Check hardware permissions: `ls -la /dev/ttyS1 /dev/port`
- Verify privileged mode is enabled
- Check pod logs for specific errors

**USB copy not working:**
- Verify `/dev/port` exists and is accessible
- Check I/O port address (0xa05) is correct for your hardware
- Ensure USB devices are mounted at `/media/usb`

**Display not responding:**
- Verify serial port device `/dev/ttyS1` exists
- Check baud rate setting (default: 1200)
- Test with different serial devices if available

**Permission errors:**
- Ensure privileged mode is enabled
- Verify running as root user
- Check hardware device permissions

### Getting Help

Check logs for detailed error information:
```bash
kubectl logs deployment/qnap-display-control -n ix-qnap-display-control --tail=100
```

## 📝 Notes

- This app requires privileged access for hardware I/O operations
- TrueNAS SCALE's Kubernetes infrastructure provides automatic restarts
- Configuration persists across app updates
- USB copy operations create timestamped directories
- The app integrates with TrueNAS's existing storage pools

## 🔄 Updates

To update the application:
1. Pull latest changes
2. Rebuild Docker image: `docker build -t qnap-display-control:latest .`
3. Upgrade Helm chart: `helm upgrade qnap-display-control ./truenas-app -n ix-qnap-display-control`

The systemd service will automatically restart the app if needed.
