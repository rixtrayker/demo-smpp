#!/bin/bash
# Script to create a proxy for capturing SMPP communication (port 2775) using tcpdump

# Capture interface (replace with the actual interface used for SMPP traffic)
capture_interface="any"

# Function to generate a unique filename with timestamp and process ID
generate_filename() {
  local timestamp=$(date +"%Y-%m-%d-%H:%M:%S")
  local random_id=$(echo $RANDOM)
  echo "proxy-dump-${timestamp}-${random_id}.pcap"
}

# Capture packets with tcpdump
filename=$(generate_filename)
if command -v tcpdump >/dev/null 2>&1; then
  tcpdump -i "$capture_interface" \
         -w "$filename" \
         -s 0 -nn -tttt -A 'port 2775' \
  && echo "Packets captured successfully to $filename" || echo "Error: Failed to capture packets."
else
  echo "Error: tcpdump is not installed."
fi

# old command example
# tcpdump -i any -w proxy-dump-$(date +"%Y-%m-%d-%H:%M:%S").pcap -s 0 -nn -tttt -A 'port 2775's