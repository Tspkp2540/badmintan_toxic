# ===== Build Stage =====
FROM node:20-alpine AS builder

WORKDIR /app

# Install frontend dependencies
COPY package.json package-lock.json* ./
RUN npm install

# Install server dependencies
COPY server/package.json server/package-lock.json* ./server/
RUN cd server && npm install

# Copy source
COPY . .

# Build Vue frontend
RUN npm run build

# ===== Production Stage =====
FROM node:20-alpine AS production

WORKDIR /app

# Install only server production dependencies
COPY server/package.json server/package-lock.json* ./server/
RUN cd server && npm install --omit=dev

# Copy server source
COPY server/ ./server/

# Copy built frontend
COPY --from=builder /app/dist ./dist

# Create data directory for SQLite
RUN mkdir -p /app/data

ENV NODE_ENV=production
ENV PORT=3000
ENV DB_PATH=/app/data/badminton.db

EXPOSE 3000

# Run server with tsx
CMD ["npx", "-y", "tsx", "server/index.ts"]
