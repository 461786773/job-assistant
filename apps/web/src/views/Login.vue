<template>
  <section class="auth-page">
    <div class="auth-card">
      <h1>登录</h1>
      <p class="muted">登录后只能看到自己的求职任务。</p>
      <form class="form" @submit.prevent="submit">
        <label>
          用户名
          <input v-model="form.username" autocomplete="username" placeholder="字母/数字/下划线" required />
        </label>
        <label>
          密码
          <input v-model="form.password" type="password" autocomplete="current-password" required />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="busy">
          {{ busy ? '登录中…' : '登录' }}
        </button>
      </form>
      <p class="auth-switch muted">
        还没有账号？
        <router-link :to="{ path: '/register', query: $route.query }">注册</router-link>
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
const form = reactive({ username: '', password: '' })

async function submit() {
  error.value = ''
  busy.value = true
  try {
    const data = await api.login({
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
