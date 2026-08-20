<template>
  <section class="auth-page welcome-page">
    <div class="welcome-stage">
      <p class="welcome-brand reveal">求职助手</p>
      <p class="welcome-sub reveal reveal-delay-1">职场心理教练</p>
      <h1 class="welcome-title reveal reveal-delay-2">你先坐一会儿。</h1>
      <p class="welcome-lead reveal reveal-delay-3">
        这里有人听你把话说完。你可以先做职场画像或心理评估——之后的对话会据此调整；不想测，直接去聊也完全可以。
      </p>
      <form class="welcome-form reveal reveal-delay-4" @submit.prevent="submit">
        <label>
          怎么称呼你
          <input
            v-model="form.username"
            autocomplete="username"
            placeholder="例如 default"
            required
          />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="busy">
          {{ busy ? '正在开门…' : '进来坐坐' }}
        </button>
        <p class="muted privacy-line" style="margin-top: 12px; font-size: 0.86rem">
          对话与评估仅本人可见 · 非临床诊断
        </p>
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
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
      return
    }
    await router.replace(postLoginPath())
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
