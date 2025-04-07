#!/bin/bash

LOGFILE=$1

if [[ -z "$LOGFILE" ]]; then
    echo "Log file not provided!"
    exit 1
fi

echo "Starting counting services"

grep -oP "'name'\s*=>\s*'\K[^']+" "$LOGFILE" | sort | uniq -c | sort -nr

echo "Searching for passwords."

PASSWORDS=(
    $(grep -oP 'id="admin_pass" size="30" value="\K[^"]+' "$LOGFILE" | sort | uniq)
    $(grep -oP "'softdbpass'\s*=>\s*'\K[^']+" "$LOGFILE" | sort | uniq)
    $(grep -oP "'admin_pass'\s*=>\s*'\K[^']+" "$LOGFILE" | sort | uniq)
)

echo ''

check_passwords() {
    local regex='^(?=.*\d)(?=.*[A-Z])(?=.*[a-z])(?=.*[^\w\d\s:])([^\s]){10,}$'
    for password in "${PASSWORDS[@]}"; do
        if echo "$password" | grep -Pq "$regex"; then
            echo -e "\e[32mSTRONG PASSWORD      $password\e[0m"
        else
            echo -e "\e[31mWEAK PASSWORD        $password\e[0m"
        fi
    done
}

echo "Checking passwords strength"
check_passwords

echo "Script execution completed."
