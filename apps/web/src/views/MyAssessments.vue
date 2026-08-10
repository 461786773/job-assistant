<template>
  <section>
    <div class="hero">
      <h1>我的评估</h1>
      <p>回看历次问卷与 AI 分析；可与跟踪曲线对照。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <div class="row" style="margin-bottom: 12px; flex-wrap: wrap; gap: 8px">
      <router-link class="btn btn-ghost" to="/onboarding/assessment">更新基线评测</router-link>
      <router-link class="btn btn-ghost" to="/wellbeing">心理跟踪</router-link>
    </div>

    <p v-if="loading" class="muted">加载中…</p>
    <p v-else-if="!items.length" class="muted">还没有评测。请先完成初次心理评测。</p>
    <div v-else class="grid">
      <router-link
        v-for="a in items"
        :key="a.id"
        :to="`/assessments/${a.id}`"
        class="task-card"
      >
        <div class="task-card-main">
          <h3>{{ SCENE_LABEL[a.primaryScene] || a.primaryScene }} · 评测</h3>
          <div class="meta">
            <div>{{ headline(a) }}</div>
            <div>{{ formatTime(a.completedAt) }}</div>
          </div>
        </div>
        <span class="badge">{{ a.crisisLevel === 'elevated' ? '需关注' : '可回看' }}</span>
      </router-link>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, SCENE_LABEL } from '../api'

const items = ref([])
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

function headline(a) {
  const raw = a.aiAnalysis
  let obj = raw
  if (typeof raw === 'string') {
    try {
      obj = JSON.parse(raw)
    } catch {
      return a.summaryForCoach?.slice(0, 80) || ''
    }
  }
  return obj?.headline || a.summaryForCoach?.slice(0, 80) || ''
}

onMounted(async () => {
  try {
    const data = await api.listAssessments()
    items.value = data.items || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
