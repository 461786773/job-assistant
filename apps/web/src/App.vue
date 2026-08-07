<template>
  <div class="shell">
    <header class="topbar">
      <router-link to="/" class="brand">
        <span class="brand-mark">求职助手</span>
        <span class="brand-sub">简历优化 · 面试模拟 · 薪资确认</span>
      </router-link>
      <nav class="nav">
        <router-link to="/">工作台</router-link>
        <router-link to="/tasks/new" class="btn btn-primary">新建任务</router-link>
      </nav>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
</script>
