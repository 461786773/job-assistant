<template>
  <section>
    <div class="hero">
      <h1>我留下来的记录</h1>
      <p>职场画像、详细评估，都在这里，想回看时随时翻开。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <div class="row" style="margin-bottom: 12px; flex-wrap: wrap; gap: 8px">
      <router-link class="btn btn-ghost" to="/bigfive">职场画像</router-link>
      <router-link class="btn btn-ghost" to="/onboarding/assessment">想被更深接住时</router-link>
      <router-link class="btn btn-ghost" to="/wellbeing">我的状态</router-link>
    </div>

    <p v-if="loading" class="muted">加载中…</p>
    <template v-else>
      <section v-if="bigFive" class="panel" style="margin-bottom: 14px">
        <h2 class="section-title">职场画像</h2>
        <p>
          <strong>{{ bigFive.personaTitle }}</strong>
          —— {{ bigFive.personaBlurb }}
        </p>
        <div class="tag-cloud" v-if="bigFive.tags?.length" style="margin-top: 8px">
          <span v-for="t in bigFive.tags" :key="t" class="tag-pill">{{ t }}</span>
        </div>
        <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
          <router-link class="btn btn-primary" :to="`/bigfive/${bigFive.id}`">查看详情</router-link>
          <router-link class="btn btn-ghost" to="/bigfive">更新画像</router-link>
        </div>
      </section>
      <section v-else class="panel" style="margin-bottom: 14px">
        <h2 class="section-title">还没有职场画像</h2>
        <p class="muted">两三分钟的趣味小像，不做诊断，也不挡着你去聊。</p>
        <router-link class="btn btn-ghost" style="margin-top: 10px" to="/bigfive">好，测一下</router-link>
      </section>

      <div v-if="!items.length" class="panel">
        <h2 class="section-title">还没有详细评估</h2>
        <p class="muted">轻松聊聊时不必做；想被更深接住时再来也不迟。</p>
        <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
          <router-link class="btn btn-primary" to="/onboarding/assessment">我想被更好地理解</router-link>
          <router-link class="btn btn-ghost" to="/home">回此刻</router-link>
        </div>
      </div>
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
    </template>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, SCENE_LABEL } from '../api'

const items = ref([])
const bigFive = ref(null)
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
    const [assess, bf] = await Promise.all([
      api.listAssessments(),
      api.latestBigFive().catch(() => null),
    ])
    items.value = assess.items || []
    bigFive.value = bf
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
