FROM alpine/psql:17.2

WORKDIR /app

COPY idf-calculator.sql /app/

COPY ./run.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]