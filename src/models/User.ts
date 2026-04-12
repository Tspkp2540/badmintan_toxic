export interface User {
  id: string
  username: string
  email: string
  fullName: string
  avatarUrl?: string
  level: number
  exp: number
  expToNextLevel: number
  rank: PlayerRank
  wins: number
  losses: number
  totalMatches: number
  winRate: number
  points: number
  createdAt: string
  updatedAt: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  fullName: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface UpdateProfileRequest {
  username?: string
  fullName?: string
  avatarUrl?: string
}

export const PlayerRank = {
  BRONZE: 'Bronze',
  SILVER: 'Silver',
  GOLD: 'Gold',
  PLATINUM: 'Platinum',
  DIAMOND: 'Diamond',
  MASTER: 'Master',
  GRANDMASTER: 'Grandmaster',
} as const

export type PlayerRank = (typeof PlayerRank)[keyof typeof PlayerRank]

export const RANK_THRESHOLDS: Record<PlayerRank, number> = {
  [PlayerRank.BRONZE]: 0,
  [PlayerRank.SILVER]: 5,
  [PlayerRank.GOLD]: 10,
  [PlayerRank.PLATINUM]: 20,
  [PlayerRank.DIAMOND]: 35,
  [PlayerRank.MASTER]: 50,
  [PlayerRank.GRANDMASTER]: 75,
}

export function getRankByLevel(level: number): PlayerRank {
  if (level >= 75) return PlayerRank.GRANDMASTER
  if (level >= 50) return PlayerRank.MASTER
  if (level >= 35) return PlayerRank.DIAMOND
  if (level >= 20) return PlayerRank.PLATINUM
  if (level >= 10) return PlayerRank.GOLD
  if (level >= 5) return PlayerRank.SILVER
  return PlayerRank.BRONZE
}

export function getExpForLevel(level: number): number {
  return Math.floor(100 * Math.pow(1.2, level - 1))
}
