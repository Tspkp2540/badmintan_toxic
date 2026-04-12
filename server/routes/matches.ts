import { Router } from 'express'
import { v4 as uuidv4 } from 'uuid'
import db from '../db/connection.js'
import { authMiddleware, type AuthRequest } from '../middleware/auth.js'
import { getExpForLevel, getRankByLevel, EXP_REWARDS, RANK_POINT_REWARDS } from '../utils/game.js'

const router = Router()

// All match routes need auth
router.use(authMiddleware)

function formatMatch(match: any) {
  const players = db.prepare(`
    SELECT mp.*, u.username, u.full_name, u.avatar_url, u.level, u.rank
    FROM match_players mp
    JOIN users u ON u.id = mp.user_id
    WHERE mp.match_id = ?
  `).all(match.id) as any[]

  const sets = db.prepare(`
    SELECT set_number, team_a_score, team_b_score
    FROM match_sets WHERE match_id = ?
    ORDER BY set_number
  `).all(match.id) as any[]

  let referee = null
  if (match.referee_id) {
    const ref: any = db.prepare('SELECT id, username, full_name FROM users WHERE id = ?').get(match.referee_id)
    if (ref) {
      referee = { id: ref.id, username: ref.username, fullName: ref.full_name }
    }
  }

  return {
    id: match.id,
    name: match.room_name,
    matchType: match.match_type,
    matchMode: match.match_mode,
    maxSets: match.max_sets,
    status: match.status,
    winnerTeam: match.winner_team,
    referee,
    createdBy: match.created_by,
    startedAt: match.started_at,
    endedAt: match.ended_at,
    createdAt: match.created_at,
    players: players.map((p: any) => ({
      userId: p.user_id,
      username: p.username,
      fullName: p.full_name,
      avatarUrl: p.avatar_url,
      level: p.level,
      rank: p.rank,
      team: p.team,
      expGained: p.exp_gained,
      rankPointsGained: p.rank_points_gained,
    })),
    sets: sets.map((s: any) => ({
      setNumber: s.set_number,
      teamA: s.team_a_score,
      teamB: s.team_b_score,
    })),
  }
}

// GET /api/matches - list active rooms
router.get('/', (req: AuthRequest, res) => {
  const status = req.query.status as string
  let query = 'SELECT * FROM matches'
  const params: any[] = []

  if (status) {
    query += ' WHERE status = ?'
    params.push(status)
  } else {
    query += " WHERE status IN ('waiting', 'playing', 'scoring')"
  }
  query += ' ORDER BY created_at DESC'

  const matches = db.prepare(query).all(...params) as any[]
  res.json(matches.map(formatMatch))
})

// POST /api/matches - create room
router.post('/', (req: AuthRequest, res) => {
  const { name, matchType, matchMode, maxSets } = req.body

  if (!name || !matchType || !matchMode) {
    res.status(400).json({ message: 'กรุณากรอกข้อมูลให้ครบ' })
    return
  }

  if (!['singles', 'doubles'].includes(matchType)) {
    res.status(400).json({ message: 'matchType ต้องเป็น singles หรือ doubles' })
    return
  }

  if (!['casual', 'ranked'].includes(matchMode)) {
    res.status(400).json({ message: 'matchMode ต้องเป็น casual หรือ ranked' })
    return
  }

  const id = uuidv4()
  db.prepare(`
    INSERT INTO matches (id, room_name, match_type, match_mode, max_sets, created_by)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(id, name, matchType, matchMode, maxSets || 3, req.userId!)

  // Creator joins team A
  db.prepare(`
    INSERT INTO match_players (match_id, user_id, team)
    VALUES (?, ?, 'A')
  `).run(id, req.userId!)

  const match = db.prepare('SELECT * FROM matches WHERE id = ?').get(id)
  res.status(201).json(formatMatch(match))
})

// POST /api/matches/:id/join
router.post('/:id/join', (req: AuthRequest, res) => {
  const { team } = req.body
  if (!team || !['A', 'B'].includes(team)) {
    res.status(400).json({ message: 'กรุณาเลือกทีม A หรือ B' })
    return
  }

  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status !== 'waiting') {
    res.status(400).json({ message: 'ห้องไม่อยู่ในสถานะรอผู้เล่น' })
    return
  }

  // Check if already in
  const existing = db.prepare('SELECT id FROM match_players WHERE match_id = ? AND user_id = ?').get(match.id, req.userId!)
  if (existing) {
    res.status(400).json({ message: 'คุณอยู่ในห้องนี้แล้ว' })
    return
  }

  if (match.referee_id === req.userId) {
    res.status(400).json({ message: 'คุณเป็นกรรมการอยู่แล้ว' })
    return
  }

  // Check team capacity
  const maxPerTeam = match.match_type === 'singles' ? 1 : 2
  const teamCount: any = db.prepare('SELECT COUNT(*) as c FROM match_players WHERE match_id = ? AND team = ?').get(match.id, team)
  if (teamCount.c >= maxPerTeam) {
    res.status(400).json({ message: 'ทีมนี้เต็มแล้ว' })
    return
  }

  db.prepare('INSERT INTO match_players (match_id, user_id, team) VALUES (?, ?, ?)').run(match.id, req.userId!, team)

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// POST /api/matches/:id/referee
router.post('/:id/referee', (req: AuthRequest, res) => {
  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status !== 'waiting') {
    res.status(400).json({ message: 'ห้องไม่อยู่ในสถานะรอผู้เล่น' })
    return
  }
  if (match.referee_id) {
    res.status(400).json({ message: 'มีกรรมการแล้ว' })
    return
  }

  const existing = db.prepare('SELECT id FROM match_players WHERE match_id = ? AND user_id = ?').get(match.id, req.userId!)
  if (existing) {
    res.status(400).json({ message: 'คุณเป็นผู้เล่นอยู่แล้ว' })
    return
  }

  db.prepare('UPDATE matches SET referee_id = ? WHERE id = ?').run(req.userId!, match.id)

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// POST /api/matches/:id/start
router.post('/:id/start', (req: AuthRequest, res) => {
  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status !== 'waiting') {
    res.status(400).json({ message: 'ห้องไม่อยู่ในสถานะรอผู้เล่น' })
    return
  }

  const teamA: any = db.prepare("SELECT COUNT(*) as c FROM match_players WHERE match_id = ? AND team = 'A'").get(match.id)
  const teamB: any = db.prepare("SELECT COUNT(*) as c FROM match_players WHERE match_id = ? AND team = 'B'").get(match.id)

  if (teamA.c === 0 || teamB.c === 0) {
    res.status(400).json({ message: 'ต้องมีผู้เล่นอย่างน้อยฝั่งละ 1 คน' })
    return
  }

  if (match.match_type === 'doubles' && (teamA.c < 2 || teamB.c < 2)) {
    res.status(400).json({ message: 'เกมคู่ต้องมีผู้เล่นฝั่งละ 2 คน' })
    return
  }

  db.prepare("UPDATE matches SET status = 'playing', started_at = datetime('now') WHERE id = ?").run(match.id)

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// POST /api/matches/:id/end
router.post('/:id/end', (req: AuthRequest, res) => {
  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status !== 'playing') {
    res.status(400).json({ message: 'เกมไม่ได้กำลังแข่ง' })
    return
  }

  db.prepare("UPDATE matches SET status = 'scoring', ended_at = datetime('now') WHERE id = ?").run(match.id)

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// POST /api/matches/:id/score
router.post('/:id/score', (req: AuthRequest, res) => {
  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status !== 'scoring') {
    res.status(400).json({ message: 'เกมไม่อยู่ในสถานะกรอกคะแนน' })
    return
  }

  const { sets } = req.body
  if (!Array.isArray(sets) || sets.length === 0) {
    res.status(400).json({ message: 'กรุณากรอกคะแนนอย่างน้อย 1 เซ็ต' })
    return
  }

  // Validate scores
  for (const s of sets) {
    if (typeof s.setNumber !== 'number' || typeof s.teamA !== 'number' || typeof s.teamB !== 'number') {
      res.status(400).json({ message: 'ข้อมูลคะแนนไม่ถูกต้อง' })
      return
    }
    if (s.teamA < 0 || s.teamB < 0 || s.teamA > 30 || s.teamB > 30) {
      res.status(400).json({ message: 'คะแนนต้องอยู่ระหว่าง 0-30' })
      return
    }
  }

  const insertSet = db.prepare('INSERT INTO match_sets (match_id, set_number, team_a_score, team_b_score) VALUES (?, ?, ?, ?)')
  const transaction = db.transaction(() => {
    // Insert sets
    for (const s of sets) {
      insertSet.run(match.id, s.setNumber, s.teamA, s.teamB)
    }

    // Determine winner
    let teamAWins = 0
    let teamBWins = 0
    for (const s of sets) {
      if (s.teamA > s.teamB) teamAWins++
      else if (s.teamB > s.teamA) teamBWins++
    }
    const winner = teamAWins > teamBWins ? 'A' : teamBWins > teamAWins ? 'B' : 'draw'

    db.prepare("UPDATE matches SET status = 'finished', winner_team = ? WHERE id = ?").run(winner, match.id)

    // Get players
    const players = db.prepare('SELECT * FROM match_players WHERE match_id = ?').all(match.id) as any[]
    const rewards = EXP_REWARDS[match.match_mode as keyof typeof EXP_REWARDS]

    const updatePlayer = db.prepare('UPDATE match_players SET exp_gained = ?, rank_points_gained = ? WHERE id = ?')
    const updateUser = db.prepare(`
      UPDATE users SET
        exp = exp + ?,
        wins = wins + ?,
        losses = losses + ?,
        draws = draws + ?,
        total_matches = total_matches + 1,
        rank_points = MAX(0, rank_points + ?),
        points = points + ?,
        updated_at = datetime('now')
      WHERE id = ?
    `)

    for (const player of players) {
      const isWinner = player.team === winner
      const isDraw = winner === 'draw'

      let expGained: number
      let rpGained = 0
      let winInc = 0, lossInc = 0, drawInc = 0

      if (isDraw) {
        expGained = rewards.draw
        drawInc = 1
        if (match.match_mode === 'ranked') rpGained = RANK_POINT_REWARDS.draw
      } else if (isWinner) {
        expGained = rewards.win
        winInc = 1
        if (match.match_mode === 'ranked') rpGained = RANK_POINT_REWARDS.win
      } else {
        expGained = rewards.lose
        lossInc = 1
        if (match.match_mode === 'ranked') rpGained = RANK_POINT_REWARDS.lose
      }

      updatePlayer.run(expGained, rpGained, player.id)
      updateUser.run(expGained, winInc, lossInc, drawInc, rpGained, expGained, player.user_id)

      // Level up check
      const user: any = db.prepare('SELECT * FROM users WHERE id = ?').get(player.user_id)
      let { level, exp, exp_to_next_level } = user
      while (exp >= exp_to_next_level) {
        exp -= exp_to_next_level
        level++
        exp_to_next_level = getExpForLevel(level)
      }
      const newRank = getRankByLevel(level)
      const totalMatches = user.total_matches
      const winRate = totalMatches > 0 ? ((user.wins / totalMatches) * 100) : 0

      db.prepare(`
        UPDATE users SET level = ?, exp = ?, exp_to_next_level = ?, rank = ?, win_rate = ?
        WHERE id = ?
      `).run(level, exp, exp_to_next_level, newRank, Math.round(winRate * 10) / 10, player.user_id)
    }
  })

  transaction()

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// POST /api/matches/:id/leave
router.post('/:id/leave', (req: AuthRequest, res) => {
  const match: any = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  if (match.status === 'playing') {
    res.status(400).json({ message: 'ไม่สามารถออกขณะกำลังแข่ง' })
    return
  }

  if (match.referee_id === req.userId) {
    db.prepare('UPDATE matches SET referee_id = NULL WHERE id = ?').run(match.id)
  } else {
    db.prepare('DELETE FROM match_players WHERE match_id = ? AND user_id = ?').run(match.id, req.userId!)
  }

  // If no players left, cancel the match
  const remaining: any = db.prepare('SELECT COUNT(*) as c FROM match_players WHERE match_id = ?').get(match.id)
  if (remaining.c === 0) {
    db.prepare("UPDATE matches SET status = 'cancelled' WHERE id = ?").run(match.id)
  }

  const updated = db.prepare('SELECT * FROM matches WHERE id = ?').get(match.id)
  res.json(formatMatch(updated))
})

// GET /api/matches/:id
router.get('/:id', (req: AuthRequest, res) => {
  const match = db.prepare('SELECT * FROM matches WHERE id = ?').get(req.params.id)
  if (!match) {
    res.status(404).json({ message: 'ไม่พบห้อง' })
    return
  }
  res.json(formatMatch(match))
})

export default router
