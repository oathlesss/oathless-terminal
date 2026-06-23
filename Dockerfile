# Stage 1: Build Vue frontend
FROM node:22-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend with embedded frontend
FROM golang:1.24-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/dist internal/handler/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /oathless-terminal ./cmd/server/

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN adduser -D -h /home/app app
USER app
WORKDIR /home/app
COPY --from=backend /oathless-terminal .
EXPOSE 8080
ENTRYPOINT ["./oathless-terminal"]
