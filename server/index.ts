import express from 'express'
import cors from 'cors'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { migrate } from './db/migrate.js'
import authRoutes from './routes/auth.js'
import rankingRoutes from './routes/rankings.js'
import matchRoutes from './routes/matches.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const PORT = parseInt(process.env.PORT || '3000', 10)
const isProduction = process.env.NODE_ENV === 'production'

// Run migrations
migrate()

const app = express()

// Middleware
app.use(cors())
app.use(express.json())

// API Routes
app.use('/api/auth', authRoutes)
app.use('/api/rankings', rankingRoutes)
app.use('/api/matches', matchRoutes)

// Health check
app.get('/api/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() })
})

// Serve Vue frontend in production
if (isProduction) {
  const clientDist = path.join(__dirname, '..', 'dist')
  app.use(express.static(clientDist))

  // SPA fallback
  app.get('*', (_req, res) => {
    res.sendFile(path.join(clientDist, 'index.html'))
  })
}

app.listen(PORT, '0.0.0.0', () => {
  console.log(`🏸 Badminton Hub server running on port ${PORT}`)
  console.log(`   Mode: ${isProduction ? 'production' : 'development'}`)
})
