import os
import re
import csv
import io
import logging
from flask import Flask, render_template, send_file, jsonify
from pymongo import MongoClient
import paramiko

SERVERS = {
    "sftp1": "192.168.33.11",
    "sftp2": "192.168.33.12",
    "sftp3": "192.168.33.13"
}

USERNAME = "sftpuser"
SSH_KEY_PATH = "my_sftp_key"
REMOTE_DIR = "/home/sftpuser/"
LOCAL_DIR = "downloaded_logs"

MONGO_URI = "mongodb://root:example@mongo:27017/"
DB_NAME = "logdb"
COLLECTION_NAME = "logs"

os.makedirs(LOCAL_DIR, exist_ok=True)
client = MongoClient(MONGO_URI)
db = client[DB_NAME]
collection = db[COLLECTION_NAME]

LOG_PATTERN = re.compile(r"(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) successfull connection from (\w+)")

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s'
)
logger = logging.getLogger(__name__)


def extract_date_from_filename(filename):
    match = re.search(r"(\d{4}-\d{2}-\d{2})", filename)
    if match:
        return match.group(1)
    return None


def process_log_file(filepath, target_server):
    with open(filepath, "r") as f:
        for line in f:
            match = LOG_PATTERN.match(line.strip())
            if match:
                date_str, time_str, execution_server = match.groups()
                doc = {
                    "date": date_str,
                    "time": time_str,
                    "execution_server": execution_server,
                    "target_server": target_server
                }
                if not collection.find_one(doc):
                    collection.insert_one(doc)


def is_date_already_processed(date_str, server_name):
    return collection.find_one({
        "date": date_str,
        "target_server": server_name
    }) is not None


def download_and_process_logs(server_name, server_ip):
    logger.info(f"Connecting to {server_name} ({server_ip})...")
    private_key = paramiko.RSAKey.from_private_key_file(SSH_KEY_PATH)
    transport = paramiko.Transport((server_ip, 22))
    transport.connect(username=USERNAME, pkey=private_key)
    sftp = paramiko.SFTPClient.from_transport(transport)

    for filename in sftp.listdir(REMOTE_DIR):
        if filename.endswith(".log"):
            log_date = extract_date_from_filename(filename)
            if not log_date:
                logger.warning(f"Could not extract date from {filename}, skipping.")
                continue

            if is_date_already_processed(log_date, server_name):
                logger.info(f"Skipping {filename} — already processed.")
                continue

            local_path = os.path.join(LOCAL_DIR, f"{server_name}_{filename}")
            remote_path = os.path.join(REMOTE_DIR, filename)
            logger.info(f"Downloading {remote_path}")
            sftp.get(remote_path, local_path)
            process_log_file(local_path, server_name)

    sftp.close()
    transport.close()


def collect_logs_from_all_servers():
    for server_name, server_ip in SERVERS.items():
        try:
            download_and_process_logs(server_name, server_ip)
        except Exception as e:
            print(f"Error with {server_name}: {e}")


def generate_html_table(data):
    if not data:
        return "<p>No data available</p>"

    headers = data[0].keys()
    html = '<table><thead><tr>' + ''.join(f"<th>{h}</th>" for h in headers) + "</tr></thead><tbody>"
    for row in data:
        html += "<tr>" + ''.join(f"<td>{row.get(h, '')}</td>" for h in headers) + "</tr>"
    html += "</tbody></table>"
    return html


app = Flask(__name__)


@app.route("/")
def index():
    data = list(collection.find({}, {"_id": 0}))
    table_html = generate_html_table(data)
    return render_template("index.html", table=table_html)


@app.route("/collect")
def collect():
    collect_logs_from_all_servers()
    return "<p>Logs collected successfully. <a href='/'>Go back</a></p>"


@app.route("/download")
def download_csv():
    data = list(collection.find({}, {"_id": 0}))
    if not data:
        return "No data available"
    headers = data[0].keys()
    buffer = io.StringIO()
    writer = csv.DictWriter(buffer, fieldnames=headers)
    writer.writeheader()
    writer.writerows(data)
    buffer.seek(0)
    return send_file(io.BytesIO(buffer.getvalue().encode()),
                     mimetype='text/csv',
                     as_attachment=True,
                     download_name='logs.csv')


@app.route("/graph")
def graph():
    return render_template("graph.html")


@app.route("/graph-data")
def graph_data():
    data = list(collection.find({}, {"_id": 0}))
    return jsonify(data)


if __name__ == "__main__":
    collect_logs_from_all_servers()
    app.run()
