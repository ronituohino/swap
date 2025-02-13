FROM alpine/psql:17.2

WORKDIR /app

COPY idf-calculator.sql /app/

COPY <<EOF /app/entrypoint.sh
#!/bin/sh
export PGPASSWORD=\${POSTGRES_PASSWORD}
psql "postgresql://\${POSTGRES_USER}:\${PGPASSWORD}@\${POSTGRES_HOST}:\${POSTGRES_PORT}/\${POSTGRES_DB}" -f /app/idf-calculator.sql
EOF

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]