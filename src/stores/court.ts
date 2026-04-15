import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { matchService } from '@/services/matchService'
import { courtService } from '@/services/courtService'

export interface MatchPlayer {
  userId: string
  username: string
  fullName: string
  avatarUrl?: string
  level: number
  rank: string
  skillLevel: string
  skillStars: number
  team: 'A' | 'B'
  expGained: number
  rankPointsGained: number
}

export interface MatchSet {
  setNumber: number
  teamA: number
  teamB: number
}

export interface MatchRoom {
  id: string
  name: string
  matchType: 'singles' | 'doubles'
  matchMode: 'casual' | 'ranked' | 'skill_test'
  maxSets: number
  status: 'waiting' | 'playing' | 'scoring' | 'finished' | 'cancelled'
  winnerTeam: 'A' | 'B' | 'draw' | null
  referee: { id: string; username: string; fullName: string } | null
  createdBy: string
  startedAt: string | null
  endedAt: string | null
  createdAt: string
  players: MatchPlayer[]
  sets: MatchSet[]
}

export const useCourtStore = defineStore('court', () => {
  const rooms = ref<MatchRoom[]>([])
  const activeRoomId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeRoom = computed(() =>
    rooms.value.find((r) => r.id === activeRoomId.value) ?? null,
  )

  const waitingRooms = computed(() =>
    rooms.value.filter((r) => r.status === 'waiting'),
  )

  const playingRooms = computed(() =>
    rooms.value.filter((r) => r.status === 'playing'),
  )

  async function fetchRooms() {
    loading.value = true
    error.value = null
    try {
      rooms.value = await matchService.getMatches()
    } catch (e: any) {
      error.value = e.response?.data?.message || 'โหลดห้องไม่สำเร็จ'
    } finally {
      loading.value = false
    }
  }

  async function fetchRoomsByCourt(courtId: string) {
    loading.value = true
    error.value = null
    try {
      rooms.value = await courtService.getCourtMatches(courtId)
    } catch (e: any) {
      error.value = e.response?.data?.message || 'โหลดห้องไม่สำเร็จ'
    } finally {
      loading.value = false
    }
  }

  async function fetchRoom(id: string) {
    try {
      const room = await matchService.getMatch(id)
      const idx = rooms.value.findIndex((r) => r.id === id)
      if (idx >= 0) {
        rooms.value[idx] = room
      } else {
        rooms.value.push(room)
      }
      return room
    } catch (e: any) {
      error.value = e.response?.data?.message || 'โหลดห้องไม่สำเร็จ'
    }
  }

  async function createRoom(options: {
    courtId: string
    name: string
    matchType: string
    matchMode: string
    maxSets: number
  }) {
    loading.value = true
    error.value = null
    try {
      const room = await matchService.createMatch(options)
      rooms.value.unshift(room)
      activeRoomId.value = room.id
      return room
    } catch (e: any) {
      error.value = e.response?.data?.message || 'สร้างห้องไม่สำเร็จ'
    } finally {
      loading.value = false
    }
  }

  async function joinRoom(roomId: string, team: string) {
    error.value = null
    try {
      const room = await matchService.joinMatch(roomId, team)
      updateRoom(room)
      activeRoomId.value = roomId
      return true
    } catch (e: any) {
      error.value = e.response?.data?.message || 'เข้าร่วมไม่สำเร็จ'
      return false
    }
  }

  async function joinAsReferee(roomId: string) {
    error.value = null
    try {
      const room = await matchService.joinAsReferee(roomId)
      updateRoom(room)
      activeRoomId.value = roomId
      return true
    } catch (e: any) {
      error.value = e.response?.data?.message || 'เป็นกรรมการไม่สำเร็จ'
      return false
    }
  }

  async function startGame(roomId: string) {
    error.value = null
    try {
      const room = await matchService.startMatch(roomId)
      updateRoom(room)
      return true
    } catch (e: any) {
      error.value = e.response?.data?.message || 'เริ่มเกมไม่สำเร็จ'
      return false
    }
  }

  async function endGame(roomId: string) {
    error.value = null
    try {
      const room = await matchService.endMatch(roomId)
      updateRoom(room)
    } catch (e: any) {
      error.value = e.response?.data?.message || 'จบเกมไม่สำเร็จ'
    }
  }

  async function submitScores(roomId: string, sets: MatchSet[]) {
    error.value = null
    try {
      const room = await matchService.submitScores(roomId, sets)
      updateRoom(room)
    } catch (e: any) {
      error.value = e.response?.data?.message || 'ส่งคะแนนไม่สำเร็จ'
    }
  }

  async function leaveRoom(roomId: string) {
    error.value = null
    try {
      await matchService.leaveMatch(roomId)
      if (activeRoomId.value === roomId) {
        activeRoomId.value = null
      }
      await fetchRooms()
    } catch (e: any) {
      error.value = e.response?.data?.message || 'ออกจากห้องไม่สำเร็จ'
    }
  }

  function updateRoom(room: MatchRoom) {
    const idx = rooms.value.findIndex((r) => r.id === room.id)
    if (idx >= 0) {
      rooms.value[idx] = room
    } else {
      rooms.value.unshift(room)
    }
  }

  function setActiveRoom(roomId: string | null) {
    activeRoomId.value = roomId
  }

  return {
    rooms,
    activeRoomId,
    activeRoom,
    waitingRooms,
    playingRooms,
    loading,
    error,
    fetchRooms,
    fetchRoomsByCourt,
    fetchRoom,
    createRoom,
    joinRoom,
    joinAsReferee,
    startGame,
    endGame,
    submitScores,
    leaveRoom,
    setActiveRoom,
  }
})
