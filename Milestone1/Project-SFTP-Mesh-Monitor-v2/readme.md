# SFTP Mesh Monitor

**SFTP Mesh Monitor** is a Go-based dashboard for collecting, storing, visualizing, and analyzing logs from multiple SFTP servers in a unified web interface.

The project uses:
- Go (with concurrency via goroutines for x500 speed improvement)
- Docker + Docker Compose
- Vagrant (for local server simulation)
- MongoDB (log storage)
- Graphs and filters via HTML/CSS + JavaScript
- Video background UI

---

## Features

- Collect logs from remote SFTP servers over SSH/SFTP
- Parse and store logs into MongoDB
- Web UI for viewing logs with filters and sorting:
  - by date
  - execution server
  - target server
- Download .csv for more advanced analysis
- Graph page with multi-line charts for visual log analysis
- Light/dark alternating table rows
- Animated background with video
- **Improved performance**: x500 speed boost using Go's concurrency
- **Reduced image size**: Over 10x smaller due to Go's compiled nature

---

## Prerequisites

Ensure you have installed:

- [Vagrant](https://www.vagrantup.com/)
- [VirtualBox](https://www.virtualbox.org/) (or other Vagrant-compatible provider)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

---

## Runbook (Step-by-Step)

### 1. Clone the repository

```bash
git clone git@github.com:bohdan28/DevOps-Projects.git

cd DevOps-Projects/Milestone1/Project-SFTP-Mesh-Monitor-v2
```
### 2. Create SSH key for SFTP access
```bash
ssh-keygen -t rsa -b 4096 -f my_sftp_key
```
This generates:

- my_sftp_key (private)

- my_sftp_key.pub (public) — to be injected into Vagrant servers

### 3. Start SFTP servers via Vagrant
```bash
vagrant up
```
This will provision 3 SFTP servers (e.g. sftp1, sftp2, sftp3) with shared log directories.

Reports of security audit are saved in project folder to ensure VMs are robust.

### 4. Configure Environment Variables

You may customize the following settings in `docker-compose.yml`:

- `SERVERS`: List of SFTP servers to monitor.
- `SSH_KEY_PATH`: Path to your SSH private key.
- `USERNAME`: Username for SFTP access.
- `REMOTE_DIR`: Directory on the remote SFTP server to monitor.
- `LOCAL_DIR`: Local directory for temporary file storage.
- `MONGO_URI`: MongoDB connection URI.
- `DB_NAME`: Name of the MongoDB database.
- `COLLECTION_NAME`: Name of the MongoDB collection for logs.
- `MONGO_INITDB_ROOT_USERNAME`: MongoDB root username.
- `MONGO_INITDB_ROOT_PASSWORD`: MongoDB root password.

Make sure to update these values as needed before proceeding.

### 5. Launch the application stack
```bash
docker-compose up --build -d
```
This starts:

- "Go app on port 80"

- "MongoDB database"

### 6. Open the web browser
Visit: [localhost_page](http://localhost:80)

line-------------------------------------------------------------------

## Project Structure
```graphql
.
├── app.go                   # Main Go application
├── templates/               # HTML templates with UI styling
├── Dockerfile               # App image
├── docker-compose.yml       # Compose definition
├── Vagrantfile              # Defines local SFTP mesh
├── script.sh                # Bash script for ssh logging
├── my_sftp_key              # Your SSH private key
├── my_sftp_key.pub          # Your SSH public key-print
├── README.md                # You're here!
```
