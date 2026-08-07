<template>
  <section>
    <div class="hero">
      <h1>工作台</h1>
      <p>按 JD 建任务，跑通简历优化、面试模拟、薪资确认。先从一份真实投递开始。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-else-if="loading" class="muted">加载中…</p>

    <div v-else-if="!tasks.length" class="panel empty">
      <h2>还没有求职任务</h2>
      <p>上传简历、粘贴 JD，创建第一条过关任务。</p>
      <div class="row" style="justify-content: center; margin-top: 18px">
        <router-link class="btn btn-primary" to="/tasks/new">新建任务</router-link>
      </div>
    </div>

    <div v-else class="grid">
      <router-link
        v-for="t in tasks"
        :key="t.id"
        :to="`/tasks/${t.id}`"
        class="task-card"
      >
        <div>
          <h3>{{ t.title || '未命名任务' }}</h3>
          <div class="meta">
            <div>{{ [t.company, t.targetRole].filter(Boolean).join(' · ') || '未填公司/岗位' }}</div>
            <div>更新于 {{ formatTime(t.updatedAt) }}</div>
          </div>
        </div>
        <span class="badge">{{ STATUS_LABEL[t.status] || t.status }}</span>
      </router-link>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, STATUS_LABEL } from '../api'

const tasks = ref([])
const loading = ref(true)
const error = ref('')

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

onMounted(async () => {
  try {
    const data = await api.listTasks()
    tasks.value = data.items || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
