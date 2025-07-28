#!/bin/bash
# Simple TrueNAS QNAP Display Control Deployment
# This script deploys using pure Kubernetes YAML (no Helm required)

set -e

echo "=== TrueNAS QNAP Display Control - Simple Deployment ==="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root" 
   exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Is this a TrueNAS SCALE system?"
    exit 1
fi

# Create directories
echo "Creating application directories..."
mkdir -p /mnt/pool0/Multimedia
mkdir -p /media/usb

# Set permissions
chmod 755 /mnt/pool0/Multimedia /media/usb

# Build Docker image if requested
if [[ "$1" == "--build" ]]; then
    echo "Building Docker image..."
    docker build -t qnap-display-control:latest .
fi

# Deploy the application
echo "Deploying QNAP Display Control..."
kubectl apply -f truenas-deployment.yaml

# Wait for deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/qnap-display-control -n qnap-display-control

# Create systemd service for auto-start
echo "Creating systemd service for auto-start..."
cat > /etc/systemd/system/qnap-display-control.service << 'EOF'
[Unit]
Description=QNAP Display Control Service
After=k3s.service
Requires=k3s.service
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/bin/kubectl apply -f /root/display/truenas-deployment.yaml
ExecStop=/usr/local/bin/kubectl delete -f /root/display/truenas-deployment.yaml
User=root
Group=root
WorkingDirectory=/root/display

[Install]
WantedBy=multi-user.target
EOF

# Update the working directory in systemd service
sed -i "s|/root/display|${PWD}|g" /etc/systemd/system/qnap-display-control.service

# Enable and start the service
systemctl daemon-reload
systemctl enable qnap-display-control.service

echo "=== Deployment Complete ==="
echo ""
echo "Application Status:"
kubectl get pods -n qnap-display-control
echo ""
echo "To check logs:"
echo "  kubectl logs -f deployment/qnap-display-control -n qnap-display-control"
echo ""
echo "To remove:"
echo "  kubectl delete -f truenas-deployment.yaml"
echo "  systemctl disable qnap-display-control.service"
echo "  rm /etc/systemd/system/qnap-display-control.service"
