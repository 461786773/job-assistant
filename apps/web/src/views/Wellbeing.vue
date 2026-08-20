<template>
  <section>
    <div class="hero">
      <h1>这几天还好吗</h1>
      <p>
        这里是你留下的那些「此刻」连成的样子。这些记录也会进入教练对话；想记一笔就
        <router-link to="/wellbeing/quick">花三分钟看看自己</router-link>。
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="coach-home-grid">
      <section class="panel">
        <h2 class="section-title">这几天的样子</h2>
        <div class="stat-row">
          <div>
            <div class="stat-num">{{ formatAvg(summary.avg7) }}</div>
            <div class="muted">近 7 日心里</div>
          </div>
          <div>
            <div class="stat-num">{{ formatAvg(summary.avg30) }}</div>
            <div class="muted">近 30 日心里</div>
          </div>
          <div>
            <div class="stat-num">{{ summary.count7 || '—' }}</div>
            <div class="muted">近 7 日记下的</div>
          </div>
        </div>
        <h3 class="pane-sub">偏沉的日子</h3>
        <ul v-if="summary.highItems?.length" class="plain-list">
          <li v-for="h in summary.highItems" :key="h.id">
            {{ formatTime(h.at) }} · 心里 {{ h.distressScore }}/10
            <span v-if="h.feelings?.length"> · {{ feelingLabels(h.feelings) }}</span>
          </li>
        </ul>
        <p v-else class="muted">近段没有特别沉的记录。</p>
        <p class="muted" style="margin-top: 12px">若一直很难受，请优先寻求专业支持；我这边是职场教练，不能替代诊疗。</p>
        <div class="row" style="margin-top: 10px; flex-wrap: wrap; gap: 8px">
          <router-link class="btn btn-primary" to="/wellbeing/quick">花三分钟看看自己</router-link>
          <router-link class="btn btn-ghost" to="/home">去找教练聊聊</router-link>
          <router-link class="btn btn-ghost" to="/assessments">我的评估</router-link>
        </div>
      </section>

      <section class="panel">
        <h2 class="section-title">想被更好地接住时</h2>
        <p class="muted">
          三分钟是对齐此刻；心理评估补场景与期望。两者都会进入之后的疏导，不是两套无关的表。
        </p>
        <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
          <router-link class="btn btn-primary" to="/onboarding/assessment">
            {{ hasBaseline ? '重新评估一下' : '好，我想被更好地理解' }}
          </router-link>
          <router-link class="btn btn-ghost" to="/wellbeing/quick">花三分钟看看自己</router-link>
        </div>
      </section>
    </div>

    <section class="panel" style="margin-top: 16px">
      <div class="workspace-toolbar">
        <h2 class="section-title" style="margin: 0">你留下的那些此刻</h2>
        <router-link class="btn btn-ghost btn-sm" to="/wellbeing/quick">再记一笔</router-link>
      </div>
      <p v-if="loading" class="muted">加载中…</p>
      <p v-else-if="!items.length" class="muted">还没有留下什么。认真聊或主动看看自己时，会记下一笔。</p>
      <ul v-else class="checkin-list">
        <li v-for="q in items" :key="q.id">
          <div>
            <strong>心里 {{ q.distressScore }}/10</strong>
            <span class="muted"> · {{ formatTime(q.at || q.createdAt) }}</span>
            <div class="meta">
              <span>{{ feelingLabels(q.feelings) }}</span>
              <span v-if="q.takeaway"> · {{ takeawayLabel(q.takeaway) }}</span>
              <span v-if="q.relatedCoachSessionId">
                · <router-link :to="`/coach/${q.relatedCoachSessionId}`">那天聊过的</router-link>
              </span>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" type="button" @click="removeQuick(q.id)">删除</button>
        </li>
      </ul>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { api, FEELING_OPTIONS, TAKEAWAY_OPTIONS } from '../api'

const loading = ref(true)
const error = ref('')
const items = ref([])
const hasBaseline = ref(false)

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

function formatAvg(v) {
  if (v == null || v === 0) return '—'
  return Number(v).toFixed(1)
}

function feelingLabels(values) {
  return (values || [])
    .map((v) => FEELING_OPTIONS.find((o) => o.value === v)?.label || v)
    .join('、')
}

function takeawayLabel(v) {
  return TAKEAWAY_OPTIONS.find((o) => o.value === v)?.label || v
}

const summary = computed(() => {
  const list = items.value || []
  const now = Date.now()
  const inDays = (q, days) => {
    const t = Date.parse(q.createdAt || q.at || '')
    return !Number.isNaN(t) && t >= now - days * 24 * 60 * 60 * 1000
  }
  const week = list.filter((q) => inDays(q, 7))
  const month = list.filter((q) => inDays(q, 30))
  const avg = (arr) =>
    arr.length ? arr.reduce((s, q) => s + (Number(q.distressScore) || 0), 0) / arr.length : 0
  const highItems = month
    .filter((q) => Number(q.distressScore) >= 8)
    .slice(0, 8)
    .map((q) => ({ ...q, at: q.createdAt || q.at }))
  return {
    count7: week.length,
    avg7: avg(week),
    avg30: avg(month),
    highItems,
  }
})

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const [quick, me] = await Promise.all([
      api.listQuickSelfChecks(),
      api.me().catch(() => null),
    ])
    items.value = quick.items || []
    hasBaseline.value = Boolean(me?.hasInitialAssessment)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function removeQuick(id) {
  if (!confirm('删掉这一笔？')) return
  try {
    await api.deleteQuickSelfCheck(id)
    await refresh()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(refresh)
</script>
