import { Router } from 'express'
import bcrypt from 'bcryptjs'
import { v4 as uuidv4 } from 'uuid'
import db from '../db/connection.js'
import { signToken, authMiddleware, type AuthRequest } from '../middleware/auth.js'
import { getExpForLevel, getRankByLevel } from '../utils/game.js'

const router = Router()

function formatUser(row: any) {
  return {
    id: row.id,
    username: row.username,
    email: row.email,
    fullName: row.full_name,
    avatarUrl: row.avatar_url,
    level: row.level,
    exp: row.exp,
    expToNextLevel: row.exp_to_next_level,
    rank: row.rank,
    wins: row.wins,
    losses: row.losses,
    totalMatches: row.total_matches,
    winRate: row.win_rate,
    points: row.points,
    rankPoints: row.rank_points,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

// POST /api/auth/register
router.post('/register', async (req, res) => {
  try {
    const { username, email, password, fullName } = req.body

    if (!username || !email || !password || !fullName) {
      res.status(400).json({ message: 'กรุณากรอกข้อมูลให้ครบ' })
      return
    }

    if (password.length < 6) {
      res.status(400).json({ message: 'รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร' })
      return
    }

    const existing = db.prepare('SELECT id FROM users WHERE email = ? OR username = ?').get(email, username)
    if (existing) {
      res.status(409).json({ message: 'อีเมลหรือชื่อผู้ใช้นี้ถูกใช้แล้ว' })
      return
    }

    const id = uuidv4()
    const passwordHash = await bcrypt.hash(password, 12)
    const expToNextLevel = getExpForLevel(1)

    db.prepare(`
      INSERT INTO users (id, username, email, password_hash, full_name, exp_to_next_level)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(id, username, email, passwordHash, fullName, expToNextLevel)

    const user = db.prepare('SELECT * FROM users WHERE id = ?').get(id)
    const token = signToken({ userId: id, email })

    res.status(201).json({ token, user: formatUser(user) })
  } catch (err: any) {
    console.error('Register error:', err)
    res.status(500).json({ message: 'เกิดข้อผิดพลาด' })
  }
})

// POST /api/auth/login
router.post('/login', async (req, res) => {
  try {
    const { email, password } = req.body

    if (!email || !password) {
      res.status(400).json({ message: 'กรุณากรอกอีเมลและรหัสผ่าน' })
      return
    }

    const user: any = db.prepare('SELECT * FROM users WHERE email = ?').get(email)
    if (!user) {
      res.status(401).json({ message: 'อีเมลหรือรหัสผ่านไม่ถูกต้อง' })
      return
    }

    const valid = await bcrypt.compare(password, user.password_hash)
    if (!valid) {
      res.status(401).json({ message: 'อีเมลหรือรหัสผ่านไม่ถูกต้อง' })
      return
    }

    const token = signToken({ userId: user.id, email: user.email })
    res.json({ token, user: formatUser(user) })
  } catch (err: any) {
    console.error('Login error:', err)
    res.status(500).json({ message: 'เกิดข้อผิดพลาด' })
  }
})

// GET /api/auth/profile
router.get('/profile', authMiddleware, (req: AuthRequest, res) => {
  const user = db.prepare('SELECT * FROM users WHERE id = ?').get(req.userId!)
  if (!user) {
    res.status(404).json({ message: 'ไม่พบผู้ใช้' })
    return
  }
  res.json(formatUser(user))
})

// PUT /api/auth/profile
router.put('/profile', authMiddleware, (req: AuthRequest, res) => {
  const { username, fullName, avatarUrl } = req.body
  const updates: string[] = []
  const values: any[] = []

  if (username) {
    const existing: any = db.prepare('SELECT id FROM users WHERE username = ? AND id != ?').get(username, req.userId!)
    if (existing) {
      res.status(409).json({ message: 'ชื่อผู้ใช้นี้ถูกใช้แล้ว' })
      return
    }
    updates.push('username = ?')
    values.push(username)
  }
  if (fullName) {
    updates.push('full_name = ?')
    values.push(fullName)
  }
  if (avatarUrl !== undefined) {
    updates.push('avatar_url = ?')
    values.push(avatarUrl)
  }

  if (updates.length === 0) {
    res.status(400).json({ message: 'ไม่มีข้อมูลที่จะอัปเดต' })
    return
  }

  updates.push("updated_at = datetime('now')")
  values.push(req.userId!)

  db.prepare(`UPDATE users SET ${updates.join(', ')} WHERE id = ?`).run(...values)

  const user = db.prepare('SELECT * FROM users WHERE id = ?').get(req.userId!)
  res.json(formatUser(user))
})

export default router
