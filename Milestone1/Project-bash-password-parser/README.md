# Log-Analizer Script

## Description
This Bash script extracts and evaluates passwords from a specified log file. It identifies password fields within the file, checks their strength using a regex pattern, and categorizes them as either **STRONG PASSWORD** or **WEAK PASSWORD** with color-coded output.

## Features
- Extracts service names and counts occurrences.
- Searches for passwords from the log file.
- Uses regex to validate password strength.
- Prints results in **green** (strong passwords) or **red** (weak passwords).


## Usage
```bash
./script.sh <logfile>
```
Example:
```bash
./script.sh server.log
```

## Password Strength Criteria
A password is considered **strong** if it:
- Has at least **10 characters**
- Contains **at least one uppercase letter**
- Contains **at least one lowercase letter**
- Contains **at least one number**
- Contains **at least one special character**

## Output Example
```
Starting counting services
      1 WordPress
...
Searching for passwords.
Checking passwords strength
STRONG PASSWORD      SecurePass@123
WEAK PASSWORD        password123
