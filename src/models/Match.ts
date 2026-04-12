import type { User } from './User'

export type MatchType = 'singles' | 'doubles'
export type MatchMode = 'casual' | 'ranked'
export type RoomStatus = 'waiting' | 'playing' | 'scoring' | 'finished'
export type TeamSide = 'A' | 'B'

export interface RoomPlayer {
  user: Pick<User, 'id' | 'username' | 'fullName' | 'avatarUrl' | 'level' | 'rank'>
  team: TeamSide
  ready: boolean
  joinedAt: number
}

export interface SetScore {
  setNumber: number
  teamA: number
  teamB: number
}

export interface MatchResult {
  sets: SetScore[]
  winner: TeamSide | 'draw'
  expGained: Record<string, number>       // userId -> exp
  rankPointsGained?: Record<string, number> // userId -> rank points (ranked only)
}

export interface CourtRoom {
  id: string
  name: string
  matchType: MatchType
  matchMode: MatchMode
  status: RoomStatus
  maxSets: number
  players: RoomPlayer[]
  referee?: Pick<User, 'id' | 'username' | 'fullName'>
  createdBy: string
  createdAt: number
  startedAt?: number
  endedAt?: number
  timeLimit: number         // in ms, default 20 min
  extendedTime: number      // extra time added in ms
  result?: MatchResult
}

export const MATCH_TIME_LIMIT = 20 * 60 * 1000 // 20 minutes

export const EXP_REWARDS = {
  casual: { win: 30, lose: 10, draw: 15 },
  ranked: { win: 50, lose: 15, draw: 25 },
}

export const RANK_POINT_REWARDS = {
  win: 25,
  lose: -10,
  draw: 5,
}
