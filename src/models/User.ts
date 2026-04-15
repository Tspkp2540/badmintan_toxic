export const UserRole = {
  ADMIN: 'admin',
  LEADER: 'leader',
  VICE_LEADER: 'vice_leader',
  PLAYER: 'player',
  GUEST: 'guest',
} as const

export type UserRole = (typeof UserRole)[keyof typeof UserRole]

export const UserRoleLabels: Record<string, string> = {
  [UserRole.ADMIN]: 'ผู้ดูแลระบบ',
  [UserRole.LEADER]: 'หัวหน้าก๊วน',
  [UserRole.VICE_LEADER]: 'รองหัวหน้าก๊วน',
  [UserRole.PLAYER]: 'ผู้เล่น',
  [UserRole.GUEST]: 'ผู้เยี่ยมชม',
}

export const SkillLevel = {
  BG1: 'BG1',
  BG2: 'BG2',
  S: 'S',
  N: 'N',
  'P-': 'P-',
  P: 'P',
  'P+': 'P+',
} as const

export type SkillLevel = (typeof SkillLevel)[keyof typeof SkillLevel]

export const SkillLevelLabels: Record<string, string> = {
  BG1: 'มือใหม่เริ่มหัด',
  BG2: 'มือหน้าบ้าน',
  S: 'มือกลาง S',
  N: 'มือกลาง N',
  'P-': 'ใกล้เคียงโค้ช',
  P: 'มือโค้ชทั่วไป',
  'P+': 'ฟอร์มนักกีฬา',
}

export const SkillLevelOrder: SkillLevel[] = ['BG1', 'BG2', 'S', 'N', 'P-', 'P', 'P+']

export interface User {
  id: string
  username: string
  email: string
  fullName: string
  role: UserRole
  avatarUrl?: string
  skillLevel: SkillLevel
  skillStars: number
  level: number
  exp: number
  expToNextLevel: number
  rank: PlayerRank
  wins: number
  losses: number
  totalMatches: number
  winRate: number
  points: number
  rankPoints: number
  promoWins: number
  promoLosses: number
  createdAt: string
  updatedAt: string
}

export interface LoginRequest {
  username: string
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

export interface UpdateUserRoleRequest {
  userId: string
  role: UserRole
}

export interface UpdateSkillLevelRequest {
  userId: string
  skillLevel: SkillLevel
  skillStars: number
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

export const TierStarsRequired: Record<string, number> = {
  BG1: 3,
  BG2: 3,
  S: 4,
  N: 4,
  'P-': 5,
  P: 5,
  'P+': 0,
}
