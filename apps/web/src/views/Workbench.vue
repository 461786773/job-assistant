<template>
  <section>
    <div class="hero">
      <h1>过关训练</h1>
      <p>按 JD 建任务，跑通人事关、业务关、谈薪关。高压节点可随时回到教练会话。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-else-if="loading" class="muted">加载中…</p>

    <div v-else-if="!tasks.length" class="panel empty">
      <h2>还没有过关任务</h2>
      <p>上传简历、粘贴 JD，创建第一条训练任务。</p>
      <div class="row" style="justify-content: center; margin-top: 18px">
        <router-link class="btn btn-primary" to="/tasks/new">新建任务</router-link>
      </div>
    </div>

    <div v-else class="grid">
      <div v-for="t in tasks" :key="t.id" class="task-card">
        <router-link :to="`/tasks/${t.id}`" class="task-card-main">
          <h3>{{ t.title || '未命名任务' }}</h3>
          <div class="meta">
            <div>{{ [t.company, t.targetRole].filter(Boolean).join(' · ') || '未填公司/岗位' }}</div>
            <div>更新于 {{ formatTime(t.updatedAt) }}</div>
          </div>
        </router-link>
        <div class="task-card-side">
          <span class="badge">{{ STATUS_LABEL[t.status] || t.status }}</span>
          <button
            class="btn btn-ghost btn-sm"
            type="button"
            :disabled="deletingId === t.id"
            @click="remove(t)"
          >
            {{ deletingId === t.id ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, STATUS_LABEL } from '../api'

const tasks = ref([])
const loading = ref(true)
const error = ref('')
const deletingId = ref('')

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

async function remove(t) {
  const name = t.title || '未命名任务'
  if (!confirm(`确认删除任务「${name}」？`)) return
  deletingId.value = t.id
  error.value = ''
  try {
    await api.deleteTask(t.id)
    tasks.value = tasks.value.filter((item) => item.id !== t.id)
  } catch (e) {
    error.value = e.message
  } finally {
    deletingId.value = ''
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
