// Unified API client for PidieKab Admin CMS

export interface ApiError {
  status: number
  code: string
  title: string
  detail: string
  requestId?: string
}

function getCsrfToken(): string {
  const match = document.cookie.match(new RegExp('(^|;\\s*)pidiekab_csrf=([^;]*)'))
  return match && match[2] ? decodeURIComponent(match[2]) : ''
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = endpoint.startsWith('/') ? endpoint : `/api/v1/admin/${endpoint}`
  const headers = new Headers(options.headers || {})

  headers.set('X-Requested-With', 'XMLHttpRequest')

  // Set CSRF token for state-changing requests
  const method = (options.method || 'GET').toUpperCase()
  if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(method)) {
    const csrf = getCsrfToken()
    if (csrf) {
      headers.set('X-CSRF-Token', csrf)
    }
  }

  // Set Content-Type only if body is JSON string (not FormData)
  if (options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  })

  // Handle 401 Unauthorized (session expired)
  // Never redirect on /auth/login or /auth/me or when already on login page to avoid infinite reload loop
  if (res.status === 401 && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/me')) {
    if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/cms/login')) {
      window.location.href = '/cms/login'
    }
    throw new Error('Sesi Anda telah kedaluwarsa. Silakan login kembali.')
  }

  // If no content returned
  if (res.status === 204) {
    return {} as T
  }

  const contentType = res.headers.get('content-type') || ''
  const isJson = contentType.includes('application/json')
  const data = isJson ? await res.json() : await res.text()

  if (!res.ok) {
    const error: ApiError = {
      status: res.status,
      code: data?.code || 'UNKNOWN_ERROR',
      title: data?.title || 'Terjadi Kesalahan',
      detail: data?.detail || data?.message || (typeof data === 'string' ? data : 'Permintaan gagal.'),
      requestId: res.headers.get('X-Request-Id') || data?.request_id,
    }
    throw error
  }

  return data as T
}

export const api = {
  get: <T>(endpoint: string) => request<T>(endpoint, { method: 'GET' }),
  post: <T>(endpoint: string, body?: any) =>
    request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),
  put: <T>(endpoint: string, body?: any) =>
    request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),
  patch: <T>(endpoint: string, body?: any) =>
    request<T>(endpoint, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    }),
  del: <T>(endpoint: string) => request<T>(endpoint, { method: 'DELETE' }),
  delete: <T>(endpoint: string) => request<T>(endpoint, { method: 'DELETE' }),
  upload: <T>(endpoint: string, formData: FormData) =>
    request<T>(endpoint, {
      method: 'POST',
      body: formData,
    }),
}
