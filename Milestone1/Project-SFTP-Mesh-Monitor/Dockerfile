FROM python:3-alpine3.20

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY app.py my_sftp_key templates/ static/ ./

CMD ["gunicorn", "-b", "0.0.0.0:8000", "app:app"]