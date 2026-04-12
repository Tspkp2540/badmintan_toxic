import { Router } from 'express'
import db from '../db/connection.js'

const router = Router()

// GET /api/rankings
router.get('/', (req, res) => {
  const page = Math.max(1, parseInt(req.query.page as string) || 1)
  const limit = Math.min(100, Math.max(1, parseInt(req.query.limit as string) || 20))
  const offset = (page - 1) * limit

  const totalRow: any = db.prepare('SELECT COUNT(*) as count FROM users WHERE total_matches > 0').get()
  const totalPlayers = totalRow.count

  const rows = db.prepare(`
    SELECT id, username, full_name, avatar_url, level, rank, wins, losses, win_rate, total_matches, points, rank_points
    FROM users
    WHERE total_matches > 0
    ORDER BY rank_points DESC, wins DESC
    LIMIT ? OFFSET ?
  `).all(limit, offset) as any[]

  const rankings = rows.map((row, i) => ({
    position: offset + i + 1,
    userId: row.id,
    username: row.username,
    fullName: row.full_name,
    avatarUrl: row.avatar_url,
    level: row.level,
    rank: row.rank,
    wins: row.wins,
    losses: row.losses,
    winRate: row.win_rate,
    totalMatches: row.total_matches,
    points: row.points,
  }))

  res.json({
    rankings,
    totalPlayers,
    currentPage: page,
    totalPages: Math.ceil(totalPlayers / limit),
  })
})

// GET /api/rankings/:userId
router.get('/:userId', (req, res) => {
  const user: any = db.prepare(`
    SELECT id, username, full_name, avatar_url, level, rank, wins, losses, win_rate, total_matches, points, rank_points
    FROM users WHERE id = ?
  `).get(req.params.userId)

  if (!user) {
    res.status(404).json({ message: 'ไม่พบผู้ใช้' })
    return
  }

  // Calculate position
  const posRow: any = db.prepare(`
    SELECT COUNT(*) + 1 as position FROM users
    WHERE total_matches > 0 AND (rank_points > ? OR (rank_points = ? AND wins > ?))
  `).get(user.rank_points, user.rank_points, user.wins)

  res.json({
    position: posRow.position,
    userId: user.id,
    username: user.username,
    fullName: user.full_name,
    avatarUrl: user.avatar_url,
    level: user.level,
    rank: user.rank,
    wins: user.wins,
    losses: user.losses,
    winRate: user.win_rate,
    totalMatches: user.total_matches,
    points: user.points,
  })
})

export default router
