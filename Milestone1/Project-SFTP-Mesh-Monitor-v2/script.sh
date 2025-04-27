#!/bin/bash

# Get the hostname of the current machine
MY_HOSTNAME=$(hostname)

# Define all servers
declare -A SERVERS
SERVERS["sftp1"]="192.168.33.11"
SERVERS["sftp2"]="192.168.33.12"
SERVERS["sftp3"]="192.168.33.13"

# Log message
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
DAY=$(date +"%Y-%m-%d")
MESSAGE="$TIMESTAMP successfull connection from $MY_HOSTNAME"

# Log file on remote servers
REMOTE_LOG_FILE="/home/sftpuser/${DAY}.log"

# Send log to all neighbors
for HOST in "${!SERVERS[@]}"; do
    if [[ "$HOST" != "$MY_HOSTNAME" ]]; then
        IP="${SERVERS[$HOST]}"
        ssh -o StrictHostKeyChecking=no -i /home/sftpuser/.ssh/id_rsa sftpuser@$IP "echo '$MESSAGE' >> $REMOTE_LOG_FILE"
    fi
done
