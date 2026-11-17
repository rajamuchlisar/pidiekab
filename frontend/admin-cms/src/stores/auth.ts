import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface AdminUser {
  id: string
  email: string
  displayName: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AdminUser | null>(null)
  const isAuthenticated = ref(false)

  function setUser(newUser: AdminUser | null) {
    user.value = newUser
    isAuthenticated.value = newUser !== null
  }

  function clearAuth() {
    user.value = null
    isAuthenticated.value = false
  }

  return {
    user,
    isAuthenticated,
    setUser,
    clearAuth,
  }
})
