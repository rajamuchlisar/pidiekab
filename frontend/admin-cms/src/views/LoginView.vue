<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref('')

async function handleLogin() {
  if (!email.value || !password.value) {
    errorMessage.value = 'Email dan password wajib diisi.'
    return
  }
  // Phase 0 placeholder
  authStore.setUser({
    id: '1',
    email: email.value,
    displayName: 'Administrator',
  })
  await router.push('/dashboard')
}
</script>

<template>
  <main class="login-container">
    <div class="login-card">
      <h1>PidieKab Admin CMS</h1>
      <p>Masuk ke portal administrasi internal.</p>

      <div v-if="errorMessage" class="error-banner">
        {{ errorMessage }}
      </div>

      <form @submit.prevent="handleLogin">
        <label for="email">Email</label>
        <input id="email" v-model="email" type="email" required autocomplete="email" />

        <label for="password">Password</label>
        <input id="password" v-model="password" type="password" required autocomplete="current-password" />

        <button type="submit">Masuk</button>
      </form>
    </div>
  </main>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: #f3f4f6;
}
.login-card {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}
.error-banner {
  background-color: #fee2e2;
  color: #991b1b;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
}
form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
}
button {
  padding: 0.75rem;
  background-color: #047857;
  color: #ffffff;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
}
</style>
