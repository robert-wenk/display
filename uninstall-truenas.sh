#!/bin/bash
# TrueNAS QNAP Display Control Uninstallation Script

set -e

APP_NAME="qnap-display-control"
NAMESPACE="ix-${APP_NAME}"
CONFIG_DIR="/mnt/pool0/apps/${APP_NAME}"

echo "=== TrueNAS QNAP Display Control Uninstallation ==="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root" 
   exit 1
fi

# Stop and disable systemd service
echo "Stopping and disabling systemd service..."
systemctl stop ${APP_NAME}.service 2>/dev/null || true
systemctl disable ${APP_NAME}.service 2>/dev/null || true
rm -f /etc/systemd/system/${APP_NAME}.service
systemctl daemon-reload

# Uninstall Helm chart
echo "Uninstalling Helm chart..."
helm uninstall ${APP_NAME} -n ${NAMESPACE} 2>/dev/null || true

# Delete namespace
echo "Deleting Kubernetes namespace..."
k3s kubectl delete namespace ${NAMESPACE} 2>/dev/null || true

# Optionally remove config directory
read -p "Remove configuration directory ${CONFIG_DIR}? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing configuration directory..."
    rm -rf "${CONFIG_DIR}"
else
    echo "Configuration directory preserved"
fi

echo "=== Uninstallation Complete ==="
