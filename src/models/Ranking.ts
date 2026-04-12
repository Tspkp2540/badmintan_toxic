export interface RankingEntry {
  position: number
  userId: string
  username: string
  fullName: string
  avatarUrl?: string
  level: number
  rank: string
  wins: number
  losses: number
  winRate: number
  totalMatches: number
  points: number
}

export interface LeaderboardResponse {
  rankings: RankingEntry[]
  totalPlayers: number
  currentPage: number
  totalPages: number
}
