#!/bin/bash

inventory_file="inventory.ini"
static_ip_playbook="bootstrap-raspies.yml"
k3s_playbook="k3s-install.yml"
ssh_password="your_ssh_password"

# Run the playbook to configure static IP, disable password logins, and create authorized SSH key
ansible-playbook -i "$inventory_file" "$static_ip_playbook" --extra-vars "ansible_ssh_pass=$ssh_password"

# Run the playbook to install and configure k3s
ansible-playbook -i "$inventory_file" "$k3s_playbook" --extra-vars "ansible_ssh_pass=$ssh_password"
