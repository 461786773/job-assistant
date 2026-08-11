<template>
  <section>
    <div class="hero">
      <h1>{{ fromOnboarding ? '我看到的你' : '这份评估' }}</h1>
      <p>描述性摘要与建议路径，不是临床诊断。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="muted">加载中…</p>

    <template v-else-if="item">
      <div v-if="analysis?.crisis || item.crisisLevel === 'elevated'" class="crisis-banner">
        {{ CRISIS_HELP }}
      </div>

      <section class="panel">
        <h2 class="section-title">状态摘要</h2>
        <p>{{ analysis?.headline || item.summaryForCoach || '这份心理评估已经完成。' }}</p>
        <div class="meta" style="margin-top: 10px">
          完成于 {{ formatTime(item.completedAt) }} · 主场景
          {{ SCENE_LABEL[item.primaryScene] || item.primaryScene }}
        </div>
      </section>

      <section class="panel" style="margin-top: 14px">
        <h2 class="section-title">可以怎么走</h2>
        <p>
          更适合从这里开始：<strong>{{ SCENE_LABEL[analysis?.suggestedScene || item.primaryScene] || analysis?.suggestedScene }}</strong>
        </p>
        <ul v-if="analysis?.nextSteps?.length" class="plain-list">
          <li v-for="(s, i) in analysis.nextSteps" :key="i">{{ s }}</li>
        </ul>
        <p class="muted" style="margin-top: 12px">{{ analysis?.boundaryNote || '本评估用于自我觉察与教练个性化，不是心理诊断。' }}</p>
      </section>

      <section class="panel" style="margin-top: 14px">
        <h2 class="section-title">你勾选的要点</h2>
        <ul class="plain-list">
          <li v-if="item.moodTags?.length">情绪：{{ item.moodTags.join('、') }}</li>
          <li v-if="item.stressors?.length">卡住你的：{{ item.stressors.slice(0, 5).join('、') }}</li>
          <li v-if="item.goals?.length">你希望：{{ item.goals.join('、') }}</li>
          <li v-if="item.freeTextBlockers">最卡住：{{ item.freeTextBlockers }}</li>
        </ul>
      </section>

      <div class="row" style="margin-top: 16px; flex-wrap: wrap; gap: 10px">
        <router-link
          v-if="fromOnboarding"
          class="btn btn-primary"
          to="/onboarding/need"
        >
          我想清楚要什么了
        </router-link>
        <router-link v-else class="btn btn-primary" to="/home">先回到安静的一页</router-link>
        <router-link class="btn btn-ghost" to="/assessments">我的评估</router-link>
        <router-link class="btn btn-ghost" to="/booking">预约私教</router-link>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, CRISIS_HELP, SCENE_LABEL } from '../api'

const route = useRoute()
const item = ref(null)
const loading = ref(true)
const error = ref('')

const fromOnboarding = computed(() => route.path.startsWith('/onboarding'))

const analysis = computed(() => {
  const raw = item.value?.aiAnalysis
  if (!raw) return null
  if (typeof raw === 'object') return raw
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
})

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const id = route.params.id
    item.value = id ? await api.getAssessment(id) : await api.latestAssessment()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(() => route.params.id, load)
onMounted(load)
</script>
