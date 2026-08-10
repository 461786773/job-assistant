<template>
  <div class="shell">
    <header class="topbar">
      <router-link to="/" class="brand">
        <span class="brand-mark">求职助手</span>
        <span class="brand-sub">简历优化 · 面试模拟 · 薪资确认</span>
      </router-link>
      <nav class="nav">
        <template v-if="auth.user">
          <router-link to="/">工作台</router-link>
          <router-link to="/tasks/new" class="btn btn-primary">新建任务</router-link>
          <span class="nav-user">{{ auth.user.username }}</span>
          <button class="btn btn-ghost" type="button" @click="logout">退出</button>
        </template>
        <template v-else>
          <router-link to="/login">登录</router-link>
          <router-link to="/register" class="btn btn-primary">注册</router-link>
        </template>
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

const router = useRouter()
const auth = useAuthState()

function logout() {
  clearSession()
  router.push('/login')
}
</script>
