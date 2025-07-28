#!/bin/bash
# TrueNAS QNAP Display Control Installation Script
# This script installs the QNAP Display Control app on TrueNAS SCALE

set -e

# Configuration
APP_NAME="qnap-display-control"
NAMESPACE="ix-${APP_NAME}"
CHART_DIR="./truenas-app"
CONFIG_DIR="/mnt/pool0/apps/${APP_NAME}/config"
MULTIMEDIA_DIR="/mnt/pool0/Multimedia"

echo "=== TrueNAS QNAP Display Control Installation ==="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root" 
   exit 1
fi

# Check if TrueNAS SCALE
if ! command -v k3s &> /dev/null; then
    echo "Error: k3s not found. This script is designed for TrueNAS SCALE."
    exit 1
fi

# Create directories
echo "Creating application directories..."
mkdir -p "${CONFIG_DIR}"
mkdir -p "${MULTIMEDIA_DIR}"

# Set permissions
chown -R root:root "${CONFIG_DIR}"
chmod 755 "${CONFIG_DIR}"

# Check for required hardware devices
echo "Checking hardware devices..."
if [[ ! -e "/dev/ttyS1" ]]; then
    echo "Warning: Serial device /dev/ttyS1 not found"
    echo "This may be normal if running on non-QNAP hardware"
fi

if [[ ! -e "/dev/port" ]]; then
    echo "Warning: /dev/port not found"
    echo "I/O port access may not be available"
fi

# Create Kubernetes namespace
echo "Creating Kubernetes namespace..."
k3s kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | k3s kubectl apply -f -

# Build Docker image if needed
if [[ "$1" == "--build" ]]; then
    echo "Building Docker image..."
    docker build -t "${APP_NAME}:latest" .
fi

# Install Helm chart
echo "Installing Helm chart..."
helm upgrade --install "${APP_NAME}" "${CHART_DIR}" \
    --namespace "${NAMESPACE}" \
    --set persistence.config.hostPath="${CONFIG_DIR}" \
    --set persistence.multimedia.hostPath="${MULTIMEDIA_DIR}" \
    --set hardwareAccess.serialPort.enabled=true \
    --set hardwareAccess.ioPortAccess.enabled=true \
    --set hardwareAccess.usbAccess.enabled=true \
    --set securityContext.privileged=true \
    --set securityContext.runAsRoot=true \
    --create-namespace \
    --wait

# Create systemd service for auto-start
echo "Creating systemd service for auto-start..."
cat > /etc/systemd/system/${APP_NAME}.service << EOF
[Unit]
Description=QNAP Display Control Service
After=k3s.service
Requires=k3s.service
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/bin/helm upgrade --install ${APP_NAME} ${PWD}/${CHART_DIR} --namespace ${NAMESPACE} --reuse-values
ExecStop=/usr/local/bin/helm uninstall ${APP_NAME} --namespace ${NAMESPACE}
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
systemctl daemon-reload
systemctl enable ${APP_NAME}.service

echo "=== Installation Complete ==="
echo "Application: ${APP_NAME}"
echo "Namespace: ${NAMESPACE}"
echo "Config Directory: ${CONFIG_DIR}"
echo "Multimedia Directory: ${MULTIMEDIA_DIR}"
echo ""
echo "To check status:"
echo "  kubectl get pods -n ${NAMESPACE}"
echo "  kubectl logs -f deployment/${APP_NAME} -n ${NAMESPACE}"
echo ""
echo "To uninstall:"
echo "  helm uninstall ${APP_NAME} -n ${NAMESPACE}"
echo "  systemctl disable ${APP_NAME}.service"
echo "  rm /etc/systemd/system/${APP_NAME}.service"
