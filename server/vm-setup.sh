#!/bin/bash
# Helper script to run commands on the Ubuntu VM via SSH
# Usage: ssh -p 2222 root@127.0.0.1 < vm-exec.sh

# Fix SSH key authentication
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Add the public key (will be replaced by actual key)
cat >> ~/.ssh/authorized_keys << 'PUBKEY'
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCxQjO/u9FEYUfb9rkHAVBwY4HRRNd2WHc43OeGt5UqkNjxpsiVcp74khGmYVgFDRZsddZ+lKFWadSRbYRwYr5RvZRhLgrcfFpcFSX8tj8jaSVo9h3dDKsQCqnYF4yetq6RfP4RthJ7V5VqHnECiWME1vLlkUdn2jpYdZlLcPXzmesAR6Fcxn4L90uAUEF/6avPRL8BvDtMh3tJlGmPhZxSA+AkTRUreUWSpvAqZmp38fXgqS+CedrLQRxTukZ2icNjZfKqVRgjVttmY3hKjOOcrgrOv2I16qkPJcPXFkZkyPGx1uiKhFaD/TP0lXDMGcTsxrqTW/25FRMOVlSUiYqCmxj9EUQRPtBi2OXHzM3KFoPiJZX1oQQmzETTBcNBCrx2P3j3bztoycRODJZgO7BIg8v8N0gm0a4WrxH+aJ6qIdYkrIECaJcp8xaW3lfw/lp3jYFbXhJaTk06EeobMhhIq2UOeqPbc0wNSSlewme6iZVbaPJzV85etBoI91RlIss207H4mT6xD/Dz6kUkZ5jkUi5c4kzFw32a1IxTq/OyxkL7CHbPCED63dPrbPmZBfCA30fNmTAFfyWEOYVMICSFMUGZRepC5LWVPu0YbraRCKWU1w2BLxZD8aLCqxaoHUujaRVs3rM8Grm58ZRelAvMh3ZARy1nVr50Cpjz0paIKQ== your_email@example.com
PUBKEY

chmod 600 ~/.ssh/authorized_keys

# Ensure sshd allows pubkey auth
sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config
systemctl restart sshd

echo "SSH key added successfully"

# Check Docker
docker --version
docker compose version
