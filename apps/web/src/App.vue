<template>
  <div class="shell">
    <header class="topbar">
      <router-link :to="auth.user ? '/home' : '/'" class="brand">
        <span class="brand-mark">求职助手</span>
        <span class="brand-sub">职场心理教练</span>
      </router-link>
      <nav class="nav" v-if="auth.user">
        <router-link to="/home">此刻</router-link>
        <router-link to="/assessments">记录</router-link>
        <router-link to="/wellbeing">状态</router-link>
        <router-link to="/booking">预约私教</router-link>
        <router-link to="/tasks">练习室</router-link>
        <router-link to="/settings">说明</router-link>
        <span class="nav-user">{{ auth.user.username }}</span>
        <button class="btn btn-ghost" type="button" @click="logout">换个名字</button>
      </nav>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { clearSession, useAuthState } from './auth'
import { clearProfileCache } from './router'

const router = useRouter()
const auth = useAuthState()

function logout() {
  clearSession()
  clearProfileCache()
  router.push('/')
}
</script>
