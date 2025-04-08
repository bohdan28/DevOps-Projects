# 📡 SFTP Mesh Monitor

**SFTP Mesh Monitor** is a Flask-based dashboard for collecting, storing, visualizing, and analyzing logs from multiple SFTP servers in a unified web interface.

The project uses:
- 🐍 Python + Flask
- 🐋 Docker + Docker Compose
- 🧰 Vagrant (for local server simulation)
- 🧮 MongoDB (log storage)
- 📊 Graphs and filters via HTML/CSS + JavaScript
- 🎥 Video background UI

---

## 🚀 Features

- Collect logs from remote SFTP servers over SSH/SFTP
- Parse and store logs into MongoDB
- Web UI for viewing logs with filters and sorting:
  - by date
  - execution server
  - target server
- Download .csv for more advanced analizis
- Graph page with multi-line charts for visual log analysis
- Light/dark alternating table rows
- Animated background with video

---

## 🛠️ Prerequisites

Ensure you have installed:

- [Vagrant](https://www.vagrantup.com/)
- [VirtualBox](https://www.virtualbox.org/) (or other Vagrant-compatible provider)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🧭 Runbook (Step-by-Step)

### 🔁 1. Clone the repository

```bash
git clone git@github.com:bohdan28/DevOps-Projects.git

cd DevOps-Projects/Milestone1/Project-SFTP-Mesh-Monitor
```
### 🔑 2. Create SSH key for SFTP access
```bash
ssh-keygen -t rsa -b 4096 -f my_sftp_key
```
This generates:

- my_sftp_key (private)

- my_sftp_key.pub (public) — to be injected into Vagrant servers

### 📦 3. Start SFTP servers via Vagrant
```bash
vagrant up
```
This will provision 3 SFTP servers (e.g. sftp1, sftp2, sftp3) with shared log directories.

Reports of security audit are saved in project folder to ensure VMs are robust.

### 🐳 4. Launch the application stack
```bash
docker-compose up --build -d
```
This starts:

- "Flask app via Gunicorn on port 80"

- "MongoDB database"

### 🌐 5. Open the web browser
Visit: [localhost_page](http://localhost:80)

line-------------------------------------------------------------------

## 📂 Project Structure
```graphql
.
├── app.py                   # Main Flask app
├── templates/               # HTML templates
├── static/
│   ├── styles.css           # UI styling
├── Dockerfile               # App image
├── docker-compose.yml       # Compose definition
├── Vagrantfile              # Defines local SFTP mesh
├── script.sh                # Bash script for ssh logging
├── my_sftp_key              # Your SSH private key
├── my_sftp_key.pub          # Your SSH public key-print
├── README.md                # You're here!
```
