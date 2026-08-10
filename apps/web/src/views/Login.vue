<template>
  <section class="auth-page">
    <div class="auth-card">
      <h1>输入用户名</h1>
      <p class="muted">初期版本无需密码。输入用户名进入教练工作台；不同用户名数据隔离。</p>
      <form class="form" @submit.prevent="submit">
        <label>
          用户名
          <input
            v-model="form.username"
            autocomplete="username"
            placeholder="默认 default"
            required
          />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="busy">
          {{ busy ? '进入中…' : '进入教练工作台' }}
        </button>
      </form>
    </div>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { setSession } from '../auth'

const router = useRouter()
const route = useRoute()
const busy = ref(false)
const error = ref('')
const form = reactive({ username: 'default' })

async function submit() {
  error.value = ''
  busy.value = true
  try {
    const data = await api.login({
      username: form.username.trim() || 'default',
    })
    setSession(data.token, data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/home'
    router.replace(redirect || '/home')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
