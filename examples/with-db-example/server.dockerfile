FROM python:3.7.9

RUN pip install -U psycopg2-binary --no-cache-dir
COPY server.py /app/server.py
