import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'
import type { AdminUser, AdminRole } from '../types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AdminUser | null>(null)
  const isAuthenticated = ref(false)
  const isChecking = ref(true)

  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')
  const isAdminOrSuperAdmin = computed(() => user.value?.role === 'super_admin' || user.value?.role === 'admin')
  const canManageAdmins = computed(() => user.value?.role === 'super_admin')

  const roleLabel = computed(() => {
    switch (user.value?.role) {
      case 'super_admin':
        return 'Super Administrator'
      case 'admin':
        return 'Administrator'
      case 'editor':
        return 'Editor Konten'
      case 'author':
        return 'Penulis Konten'
      default:
        return 'Administrator'
    }
  })

  const roleBadgeClass = computed(() => {
    switch (user.value?.role) {
      case 'super_admin':
        return 'badge-role-super'
      case 'admin':
        return 'badge-role-admin'
      case 'editor':
        return 'badge-role-editor'
      case 'author':
        return 'badge-role-author'
      default:
        return 'badge-role-admin'
    }
  })

  function hasRole(...roles: AdminRole[]): boolean {
    if (!user.value?.role) return false
    if (user.value.role === 'super_admin') return true
    return roles.includes(user.value.role)
  }

  function can(permission: string): boolean {
    if (!user.value) return false
    if (user.value.role === 'super_admin') return true
    return !!user.value.permissions?.includes(permission)
  }

  function canAccess(moduleKey: string): boolean {
    if (!user.value) return false
    if (user.value.role === 'super_admin') return true
    if (!user.value.permissions || user.value.permissions.length === 0) return false
    return user.value.permissions.includes(moduleKey)
  }

  function parseAdminData(adminData: any): AdminUser {
    return {
      id: Number(adminData.id),
      email: adminData.email,
      displayName: adminData.displayName || adminData.display_name || adminData.name || adminData.email,
      role: (adminData.role as AdminRole) || 'super_admin',
      permissions: adminData.permissions || [],
      isActive: adminData.isActive !== false && adminData.is_active !== false,
      createdAt: adminData.createdAt || adminData.created_at,
      lastLoginAt: adminData.lastLoginAt || adminData.last_login_at,
    }
  }

  async function checkSession(): Promise<boolean> {
    isChecking.value = true
    try {
      const res = await api.get<any>('/api/v1/admin/auth/me')
      const adminData = res?.admin || res?.data
      if (adminData && (adminData.email || adminData.id)) {
        user.value = parseAdminData(adminData)
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
      user.value = parseAdminData(adminData)
      isAuthenticated.value = true
    } else {
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
    isSuperAdmin,
    isAdminOrSuperAdmin,
    canManageAdmins,
    roleLabel,
    roleBadgeClass,
    hasRole,
    can,
    canAccess,
    checkSession,
    login,
    logout,
  }
})
