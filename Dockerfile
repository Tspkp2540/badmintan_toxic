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

# Bundle server with esbuild (CJS for node without tsx)
RUN cd server && npx esbuild index.ts --bundle --platform=node --target=node20 --outfile=dist/index.cjs --format=cjs --external:better-sqlite3 --external:bcryptjs

# ===== Production Stage =====
FROM node:20-alpine AS production

WORKDIR /app

# Install only server production dependencies
COPY server/package.json server/package-lock.json* ./server/
RUN cd server && npm install --omit=dev

# Copy compiled server
COPY --from=builder /app/server/dist ./server/dist

# Copy built frontend
COPY --from=builder /app/dist ./dist

# Create data directory for SQLite
RUN mkdir -p /app/data

ENV NODE_ENV=production
ENV PORT=3000
ENV DB_PATH=/app/data/badminton.db

EXPOSE 3000

CMD ["node", "server/dist/index.cjs"]
