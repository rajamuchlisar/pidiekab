import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'info' | 'warning'
  title: string
  message: string
  requestId?: string
}

export const useNotificationStore = defineStore('notification', () => {
  const toasts = ref<Toast[]>([])

  function show(type: Toast['type'], title: string, message: string, requestId?: string) {
    const id = Math.random().toString(36).substring(2, 9)
    const toast: Toast = { id, type, title, message, requestId }
    toasts.value.push(toast)

    setTimeout(() => {
      remove(id)
    }, 4500)
  }

  function success(title: string, message: string) {
    show('success', title, message)
  }

  function error(title: string, message: string, requestId?: string) {
    show('error', title, message, requestId)
  }

  function remove(id: string) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  return {
    toasts,
    show,
    success,
    error,
    remove,
  }
})
