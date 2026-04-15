import type { User, UpdateUserRoleRequest } from '@/models/User'
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
}
