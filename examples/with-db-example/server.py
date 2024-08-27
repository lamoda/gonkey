import http.server
import json
import os
import random
import socketserver
from http import HTTPStatus

import psycopg2


class Handler(http.server.SimpleHTTPRequestHandler):
    def get_response(self) -> dict:
        if self.path.startswith('/info/'):
            response = self.get_info()
        elif self.path.startswith('/randint/'):
            response = self.get_rand_num()
        else:
            response = {'non-existing': True}

        return response

    def get_info(self) -> dict:
        info_id = self.path.split('/')[-1]
        return {
            'result_id': info_id,
            'query_result': postgres.get_sql_result('SELECT id, name FROM testing LIMIT 2'),
        }

    def get_rand_num(self) -> dict:
        return {'num': {'generated': str(random.randint(0, 100))}}

    def do_GET(self):
        # заголовки ответа
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(self.get_response()).encode())


class PostgresStorage:
    def __init__(self):
        params = {
            "host": os.environ['APP_POSTGRES_HOST'],
            "port": os.environ['APP_POSTGRES_PORT'],
            "user": os.environ['APP_POSTGRES_USER'],
            "password": os.environ['APP_POSTGRES_PASS'],
            "database": os.environ['APP_POSTGRES_DB'],
        }
        self.conn = psycopg2.connect(**params)
        psycopg2.extensions.register_type(psycopg2.extensions.UNICODE, self.conn)
        self.conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_AUTOCOMMIT)
        self.cursor = self.conn.cursor()

    def apply_migrations(self):
        self.cursor.execute(
            """
        CREATE TABLE IF NOT EXISTS testing (id SERIAL PRIMARY KEY, name VARCHAR(200) NOT NULL);
        """
        )
        self.conn.commit()
        self.cursor.executemany(
            "INSERT INTO testing (name) VALUES (%(name)s);",
            [{'name': 'golang'}, {'name': 'gonkey'}, {'name': 'testing'}],
        )
        self.conn.commit()

    def get_sql_result(self, sql_str):
        self.cursor.execute(sql_str)
        query_data = list(self.cursor.fetchall())
        self.conn.commit()
        return query_data


postgres = PostgresStorage()
postgres.apply_migrations()

if __name__ == '__main__':
    service = socketserver.TCPServer(('', 5000), Handler)
    service.serve_forever()
