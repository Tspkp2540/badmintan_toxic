import type { User, UpdateUserRoleRequest, UpdateSkillLevelRequest } from '@/models/User'
import apiClient from './api'

export const userService = {
  async getUsers(): Promise<User[]> {
    const response = await apiClient.get<User[]>('/users')
    return response.data
  },

  async updateUserRole(data: UpdateUserRoleRequest): Promise<User> {
    const response = await apiClient.put<User>('/users/role', data)
    return response.data
  },

  async updateSkillLevel(data: UpdateSkillLevelRequest): Promise<any> {
    const response = await apiClient.put('/matches/skill-level', data)
    return response.data
  },
}
