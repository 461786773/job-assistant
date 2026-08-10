<template>
  <section class="auth-page">
    <div class="auth-card">
      <h1>链接进入</h1>
      <p class="muted">输入用户名进入应用（受控访问）。新用户将先完成心理评测与诉求确认。</p>
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
          {{ busy ? '进入中…' : '进入' }}
        </button>
      </form>
    </div>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, postLoginPath } from '../api'
import { setSession } from '../auth'
import { clearProfileCache } from '../router'

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
    clearProfileCache()
    const me = await api.me()
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect && me.hasInitialAssessment && me.primaryNeed) {
      router.replace(redirect)
      return
    }
    router.replace(postLoginPath(me))
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
