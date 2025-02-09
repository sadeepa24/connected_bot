#!/bin/bash

install_dependencies() {
  echo "Installing missing dependencies..."
  sudo apt update

  for cmd in jq curl unzip tar; do
    if ! command -v "$cmd" &> /dev/null; then
      echo "$cmd not found, installing..."
      sudo apt install -y "$cmd"
    fi
  done
}

for cmd in jq curl unzip tar; do
  if ! command -v "$cmd" &> /dev/null; then
    install_dependencies
    break
  fi
done

# Define paths and filenames
CONNECTED_BOT_DIR="/usr/local/etc/connected_bot"
SERVICE_FILE="/etc/systemd/system/connected-bot.service"
ZIP_FILE="connected-bot.zip"
CONFIG_JSON="config.json"
USERMSG_JSON="usermsg.json"
LATEST_RELEASE_URL="https://github.com/sadeepa24/connected_bot/releases/latest/download/connected_bot_Linux_x86_64.tar.gz"
TAR_FILE="connected_bot_Linux_x86_64.tar.gz"
TEMP_DIR="/tmp/connected_bot_extracted"

# Ensure the target directory exists
if [ ! -d "$CONNECTED_BOT_DIR" ]; then
  echo "Creating target directory: $CONNECTED_BOT_DIR"
  sudo mkdir -p "$CONNECTED_BOT_DIR"
fi

# Download the latest release of connected_bot
echo "Downloading the latest connected_bot release from GitHub..."
sudo curl -L -o "$TAR_FILE" "$LATEST_RELEASE_URL"

# Create a temporary directory and extract the archive
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
echo "Extracting $TAR_FILE..."
sudo tar -xzf "$TAR_FILE" -C "$TEMP_DIR"

# Find the actual binary
BOT_BINARY=$(find "$TEMP_DIR" -type f -name "connected_bot_linux" | head -n 1)
if [ -z "$BOT_BINARY" ]; then
  echo "Error: connected_bot binary not found in extracted files!"
  exit 1
fi

# Move the binary to the target directory
sudo mv "$BOT_BINARY" "$CONNECTED_BOT_DIR/connected_bot"
sudo chmod +x "$CONNECTED_BOT_DIR/connected_bot"

# Cleanup temporary files
rm -rf "$TEMP_DIR"
rm -f "$TAR_FILE"

# Check if the zip file exists in the current directory
if [ -f "$ZIP_FILE" ]; then
  echo "Found $ZIP_FILE, unzipping..."
  sudo unzip -o "$ZIP_FILE" -d "$CONNECTED_BOT_DIR"

  # Check for the presence of config.json and usermsg.json
  if [ ! -f "$CONNECTED_BOT_DIR/$CONFIG_JSON" ] || [ ! -f "$CONNECTED_BOT_DIR/$USERMSG_JSON" ]; then
    echo "Error: config.json or usermsg.json not found!"
    exit 1
  fi

  # Check the paths in config.json and verify if the files exist
  if [ -f "$CONNECTED_BOT_DIR/$CONFIG_JSON" ]; then
    SBOX_PATH=$(jq -r '.sbox_path' "$CONNECTED_BOT_DIR/$CONFIG_JSON")
    TEMPLATES_PATH=$(jq -r '.templates_path' "$CONNECTED_BOT_DIR/$CONFIG_JSON")
    
    if [ ! -f "$SBOX_PATH" ]; then
      echo "Error: sbox_path file not found!"
      exit 1
    fi

    if [ ! -f "$TEMPLATES_PATH" ]; then
      echo "Error: templates_path file not found!"
      exit 1
    fi
  fi
else
  echo "Warning: $ZIP_FILE not found, skipping configuration files extraction."
fi

# Check if systemd service exists, if not, add the service
if [ ! -f "$SERVICE_FILE" ]; then
  echo "Creating systemd service for connected-bot..."
  sudo bash -c "cat << EOF > $SERVICE_FILE
[Unit]
Description=Connected Bot Service
After=network.target

[Service]
ExecStart=$CONNECTED_BOT_DIR/connected_bot
WorkingDirectory=$CONNECTED_BOT_DIR
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF"

  # Reload systemd and enable the service
  sudo systemctl daemon-reload
  sudo systemctl enable connected-bot.service
fi

# Start the service
echo "Starting connected_bot service..."
sudo systemctl start connected-bot.service

# Check if the service started successfully
sleep 2
if ! sudo systemctl is-active --quiet connected-bot.service; then
  echo "Error: connected_bot service failed to start. See logs below:"
  sudo systemctl status connected-bot.service --no-pager
  exit 1
fi

echo "connected_bot setup completed successfully!"
