# ===== Stage 1: Build Vue frontend =====
FROM node:20-alpine AS frontend

WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build

# ===== Stage 2: Build Go backend =====
FROM golang:1.22-alpine AS backend

WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o badminton-server .

# ===== Stage 3: Production =====
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend /app/badminton-server .
COPY --from=frontend /app/dist ./dist
RUN mkdir -p /app/data

ENV NODE_ENV=production
ENV DB_PATH=/app/data/badminton.db

EXPOSE 3000

CMD ["./badminton-server"]
