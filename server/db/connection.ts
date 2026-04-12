import Database from 'better-sqlite3'
import path from 'node:path'
import fs from 'node:fs'

const DB_PATH = process.env.DB_PATH || path.join(process.cwd(), 'data', 'badminton.db')

// Ensure data directory exists
fs.mkdirSync(path.dirname(DB_PATH), { recursive: true })

const db = new Database(DB_PATH)

// Enable WAL mode for better concurrent read performance
db.pragma('journal_mode = WAL')
db.pragma('foreign_keys = ON')

export default db
