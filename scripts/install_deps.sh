#!/bin/bash

# Internet Monitor - System Dependencies Installation Script
# Installs required system-level packages for network monitoring

set -e  # Exit on any error

echo "🚀 Installing Internet Monitor System Dependencies..."

# Update package list
echo "📦 Updating package list..."
sudo apt-get update

# Install basic network utilities
echo "🌐 Installing network utilities..."
sudo apt-get install -y \
    net-tools \
    dnsutils \
    wireless-tools \
    sqlite3 \
    curl \
    wget \
    iputils-ping

# Install Ookla Speedtest CLI
echo "⚡ Installing Ookla Speedtest CLI..."
if ! command -v speedtest &> /dev/null; then
    # Add Ookla repository and install speedtest
    sudo apt-get install -y gnupg1 apt-transport-https dirmngr
    export INSTALL_KEY=379CE192D401AB61
    sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys $INSTALL_KEY
    echo "deb https://ookla.bintray.com/debian generic main" | sudo tee /etc/apt/sources.list.d/speedtest.list
    sudo apt-get update
    sudo apt-get install -y speedtest
    
    # Accept Ookla license automatically on first run
    echo "📝 Accepting Ookla Speedtest license..."
    speedtest --accept-license --accept-gdpr > /dev/null 2>&1 || true
else
    echo "✅ Speedtest CLI already installed"
fi

# Verify installations
echo "🔍 Verifying installations..."

echo -n "  - ping: "
if command -v ping &> /dev/null; then
    echo "✅ OK ($(ping -V 2>&1 | head -n1))"
else
    echo "❌ MISSING"
fi

echo -n "  - dig: "
if command -v dig &> /dev/null; then
    echo "✅ OK ($(dig -v 2>&1 | head -n1))"
else
    echo "❌ MISSING"
fi

echo -n "  - iwconfig: "
if command -v iwconfig &> /dev/null; then
    echo "✅ OK"
else
    echo "❌ MISSING"
fi

echo -n "  - speedtest: "
if command -v speedtest &> /dev/null; then
    echo "✅ OK ($(speedtest --version))"
else
    echo "❌ MISSING"
fi

echo -n "  - sqlite3: "
if command -v sqlite3 &> /dev/null; then
    echo "✅ OK ($(sqlite3 --version | cut -d' ' -f1))"
else
    echo "❌ MISSING"
fi

echo "🎉 System dependencies installation completed!"
echo ""
echo "Next steps:"
echo "1. Activate virtual environment: source .venv/bin/activate"
echo "2. Install Python dependencies: pip install -r requirements.txt"
echo "3. Initialize database: python -c 'from src.database import init_db; init_db()'"