import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api/client'
import type { AdminUser } from '../types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AdminUser | null>(null)
  const isAuthenticated = ref(false)
  const isChecking = ref(true)

  async function checkSession(): Promise<boolean> {
    isChecking.value = true
    try {
      const res = await api.get<any>('/api/v1/admin/auth/me')
      const adminData = res?.admin || res?.data
      if (adminData && (adminData.email || adminData.id)) {
        user.value = {
          id: Number(adminData.id),
          email: adminData.email,
          displayName: adminData.displayName || adminData.name || adminData.email,
          isActive: adminData.isActive !== false,
          createdAt: adminData.createdAt,
        }
        isAuthenticated.value = true
        return true
      }
      user.value = null
      isAuthenticated.value = false
      return false
    } catch {
      user.value = null
      isAuthenticated.value = false
      return false
    } finally {
      isChecking.value = false
    }
  }

  async function login(email: string, password: string): Promise<void> {
    const res = await api.post<any>('/api/v1/admin/auth/login', {
      email,
      password,
    })
    const adminData = res?.admin || res?.data
    if (adminData && (adminData.email || adminData.id)) {
      user.value = {
        id: Number(adminData.id),
        email: adminData.email,
        displayName: adminData.displayName || adminData.name || adminData.email,
        isActive: adminData.isActive !== false,
        createdAt: adminData.createdAt,
      }
      isAuthenticated.value = true
    } else {
      // Re-check session in case cookies were established
      await checkSession()
    }
  }

  async function logout(): Promise<void> {
    try {
      await api.post('/api/v1/admin/auth/logout')
    } catch {
      // ignore
    } finally {
      user.value = null
      isAuthenticated.value = false
      window.location.href = '/cms/login'
    }
  }

  return {
    user,
    isAuthenticated,
    isChecking,
    checkSession,
    login,
    logout,
  }
})
