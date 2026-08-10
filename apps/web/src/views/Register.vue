<template>
  <section class="auth-page">
    <div class="auth-card">
      <h1>注册</h1>
      <p class="muted">创建账号后，任务仅对本账号可见。</p>
      <form class="form" @submit.prevent="submit">
        <label>
          用户名
          <input
            v-model="form.username"
            autocomplete="username"
            placeholder="2–32 位字母、数字或下划线"
            required
          />
        </label>
        <label>
          密码
          <input
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            placeholder="至少 6 位"
            required
          />
        </label>
        <label>
          确认密码
          <input v-model="form.confirm" type="password" autocomplete="new-password" required />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="busy">
          {{ busy ? '创建中…' : '注册并登录' }}
        </button>
      </form>
      <p class="auth-switch muted">
        已有账号？
        <router-link :to="{ path: '/login', query: $route.query }">登录</router-link>
      </p>
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
const form = reactive({ username: '', password: '', confirm: '' })

async function submit() {
  error.value = ''
  if (form.password !== form.confirm) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (form.password.length < 6) {
    error.value = '密码至少 6 位'
    return
  }
  busy.value = true
  try {
    const data = await api.register({
      username: form.username.trim(),
      password: form.password,
    })
    setSession(data.token, data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.replace(redirect || '/')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
