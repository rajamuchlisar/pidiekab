<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useNotificationStore } from '../stores/notification'

const router = useRouter()
const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')

async function handleLogin() {
  errorMessage.value = ''
  if (!email.value || !password.value) {
    errorMessage.value = 'Email dan kata sandi wajib diisi.'
    return
  }

  isLoading.value = true
  try {
    await authStore.login(email.value, password.value)
    if (authStore.isAuthenticated) {
      notificationStore.success('Login Berhasil', `Selamat datang kembali, ${authStore.user?.displayName}!`)
      await router.push('/dashboard')
    } else {
      errorMessage.value = 'Email atau kata sandi tidak valid.'
    }
  } catch (err: any) {
    errorMessage.value = err.detail || 'Email atau kata sandi tidak valid.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="card-header">
        <div class="logo-circle">PK</div>
        <h2 class="title">PidieKab Admin CMS</h2>
        <p class="subtitle">Platform Manajemen Konten & Informasi Daerah</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div v-if="errorMessage" class="error-alert">
          {{ errorMessage }}
        </div>

        <div class="form-group">
          <label for="email">Alamat Email</label>
          <input
            id="email"
            name="email"
            v-model="email"
            type="email"
            inputmode="email"
            autocapitalize="none"
            autocorrect="off"
            spellcheck="false"
            required
            placeholder="admin@pidiekab.web.id"
            autocomplete="username email"
            class="input-control"
          />
        </div>

        <div class="form-group">
          <label for="password">Kata Sandi</label>
          <div class="password-wrapper">
            <input
              id="password"
              name="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="••••••••"
              autocomplete="current-password"
              autocapitalize="none"
              autocorrect="off"
              spellcheck="false"
              class="input-control"
            />
            <button
              type="button"
              class="btn-toggle-pw"
              @click="showPassword = !showPassword"
              tabindex="-1"
            >
              {{ showPassword ? 'Sembunyikan' : 'Lihat' }}
            </button>
          </div>
        </div>

        <button type="submit" class="btn-submit" :disabled="isLoading">
          <span v-if="isLoading">Memverifikasi...</span>
          <span v-else>Masuk ke Panel Admin</span>
        </button>
      </form>

      <div class="card-footer">
        <p>Akses terbatas hanya untuk administrator internal PidieKab.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #064e3b 0%, #022c22 100%);
  padding: 1rem;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: #ffffff;
  border-radius: 14px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.25), 0 10px 10px -5px rgba(0, 0, 0, 0.1);
  padding: 2.25rem 2rem;
  margin: auto;
}

@media (max-width: 480px) {
  .login-card {
    padding: 1.75rem 1.25rem;
    border-radius: 12px;
  }
}

.card-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo-circle {
  width: 52px;
  height: 52px;
  background: #10b981;
  color: #064e3b;
  font-weight: 800;
  font-size: 1.4rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  margin: 0 auto 1rem auto;
}

.title {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.35rem 0;
}

.subtitle {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.error-alert {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
  padding: 0.65rem 0.85rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  user-select: none;
}

.input-control {
  width: 100%;
  -webkit-appearance: none;
  appearance: none;
  padding: 0.75rem 0.95rem;
  border: 1.5px solid #cbd5e1;
  border-radius: 8px;
  font-size: 16px; /* CRITICAL: Minimum 16px prevents mobile browser auto-zoom and enables touch keyboard trigger */
  line-height: 1.5;
  color: #0f172a;
  background-color: #ffffff;
  transition: border-color 0.15s, box-shadow 0.15s;
  cursor: text;
  -webkit-user-select: text;
  user-select: text;
  touch-action: manipulation;
}

.input-control:focus {
  outline: none;
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15);
}

.password-wrapper {
  position: relative;
  display: flex;
  width: 100%;
}

.password-wrapper .input-control {
  width: 100%;
  padding-right: 5rem;
}

.btn-toggle-pw {
  position: absolute;
  right: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  font-size: 0.8rem;
  color: #059669;
  font-weight: 600;
  cursor: pointer;
  padding: 0.35rem 0.6rem;
  touch-action: manipulation;
  z-index: 2;
}

.btn-submit {
  margin-top: 0.5rem;
  background: #059669;
  color: #ffffff;
  border: none;
  padding: 0.85rem;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s ease;
  touch-action: manipulation;
}

.btn-submit:hover:not(:disabled) {
  background: #047857;
}

.btn-submit:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.card-footer {
  margin-top: 1.75rem;
  text-align: center;
  border-top: 1px solid #f1f5f9;
  padding-top: 1rem;
}

.card-footer p {
  font-size: 0.75rem;
  color: #94a3b8;
  margin: 0;
}
</style>
