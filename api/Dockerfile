### Build step
FROM golang:1.23.1-alpine3.19 AS builder

WORKDIR /app

# Copy code
COPY . .

# Download Go modules
RUN go mod download

# Build binary
RUN go build -o ./bin/todo-api ./main.go



### Run step 
FROM scratch AS runner

WORKDIR /usr/src/app

COPY --from=builder /app/bin/search-api ./search-api

ENV GIN_MODE=release

CMD ["./search-api"]