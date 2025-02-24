#!/bin/bash

# Install curl if not already installed
sudo apt-get update
sudo apt-get install -y curl

# Add Tailscale GPG key
curl -fsSL https://pkgs.tailscale.com/stable/raspbian/bullseye.noarmor.gpg | sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg > /dev/null

# Add the Tailscale repository
curl -fsSL https://pkgs.tailscale.com/stable/raspbian/bullseye.tailscale-keyring.list | sudo tee /etc/apt/sources.list.d/tailscale.list

# Update package list and install Tailscale
sudo apt-get update
sudo apt-get install -y tailscale

# Enable IP forwarding
echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.d/99-tailscale.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.d/99-tailscale.conf
sudo sysctl -p /etc/sysctl.d/99-tailscale.conf

# Set up Tailscale with routes and exit node
sudo tailscale up --advertise-routes=192.168.1.0/24 --advertise-exit-node

echo "Tailscale setup is complete!"
