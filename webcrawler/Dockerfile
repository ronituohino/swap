FROM ghcr.io/astral-sh/uv:python3.13-alpine

WORKDIR /app

ENV UV_COMPILE_BYTECODE=1

COPY . .

RUN uv sync --frozen --no-dev

CMD ["uv", "run", "scrapy", "crawl", "infinite"]