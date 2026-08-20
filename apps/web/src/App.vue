<template>
  <div class="shell" :class="{ 'shell-welcome': !auth.user }">
    <header v-if="auth.user" class="topbar">
      <router-link to="/home" class="brand">
        <span class="brand-mark">求职助手</span>
        <span class="brand-sub">职场心理教练</span>
      </router-link>
      <nav class="nav">
        <router-link to="/home">首页</router-link>
        <router-link to="/assessments">我的</router-link>
        <router-link to="/settings">设置</router-link>
        <details class="nav-more">
          <summary>更多</summary>
          <div class="nav-more-menu">
            <router-link to="/booking">预约私教</router-link>
            <router-link to="/tasks">练习室</router-link>
          </div>
        </details>
        <span class="nav-user">{{ auth.user.username }}</span>
        <button class="btn btn-ghost btn-sm" type="button" @click="logout">换个名字</button>
      </nav>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { clearSession, useAuthState } from './auth'
import { crisisHelp, loadCopy } from './copy'
import { clearProfileCache } from './router'

const router = useRouter()
const auth = useAuthState()

onMounted(() => {
  if (!crisisHelp.value) loadCopy().catch(() => {})
})

function logout() {
  clearSession()
  clearProfileCache()
  router.push('/')
}
</script>
