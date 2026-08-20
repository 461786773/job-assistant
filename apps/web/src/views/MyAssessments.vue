<template>
  <section>
    <div class="hero">
      <h1>我的</h1>
      <p>心理评估、跟踪快照和职场画像放在这里回看。对话会合读这三块。</p>
    </div>

    <GuideNote title="这些资料会进入教练对话">
      <p>
        之后疏导时，我会合读<strong>最新心理评估</strong>、<strong>职场画像</strong>和<strong>此刻记录</strong>。更新任一一项，下一轮聊天就会跟着变。
      </p>
    </GuideNote>

    <p v-if="error" class="error">{{ error }}</p>
    <div class="row" style="margin-bottom: 12px; flex-wrap: wrap; gap: 8px">
      <router-link class="btn btn-primary" to="/onboarding/assessment">
        {{ hasBaseline ? '重新评估一下' : '好，我想被更好地理解' }}
      </router-link>
      <router-link class="btn btn-ghost" to="/wellbeing/quick">花三分钟看看自己</router-link>
      <router-link class="btn btn-ghost" to="/bigfive">职场画像</router-link>
    </div>

    <p v-if="loading" class="muted">加载中…</p>
    <template v-else>
      <!-- ① 心理评估（基线） -->
      <section class="panel" style="margin-bottom: 14px">
        <h2 class="section-title">① 心理评估</h2>
        <template v-if="latestBaseline">
          <p>
            <strong>{{ SCENE_LABEL[latestBaseline.primaryScene] || latestBaseline.primaryScene || '综合' }}</strong>
            <span class="muted"> · {{ formatTime(latestBaseline.completedAt || latestBaseline.createdAt) }}</span>
          </p>
          <p style="margin-top: 8px">{{ headline(latestBaseline) || '你已经有一份心理评估了。' }}</p>
          <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
            <router-link class="btn btn-primary" :to="`/assessments/${latestBaseline.id}`">
              看看这份摘要
            </router-link>
            <router-link class="btn btn-ghost" to="/onboarding/assessment">重新评估一下</router-link>
            <router-link class="btn btn-ghost" to="/home">去聊聊</router-link>
          </div>
          <details v-if="olderBaselines.length" style="margin-top: 14px">
            <summary class="muted">更早的评估（{{ olderBaselines.length }}）</summary>
            <ul class="plain-list" style="margin-top: 8px">
              <li v-for="a in olderBaselines" :key="a.id">
                <router-link :to="`/assessments/${a.id}`">
                  {{ SCENE_LABEL[a.primaryScene] || a.primaryScene || '综合' }}
                </router-link>
                <span class="muted"> · {{ formatTime(a.completedAt || a.createdAt) }}</span>
              </li>
            </ul>
          </details>
        </template>
        <template v-else>
          <p class="muted">还没有心理评估。做完后，教练疏导会优先参考这份底；想被更好地接住时再慢慢做也行。</p>
          <router-link class="btn btn-primary" style="margin-top: 10px" to="/onboarding/assessment">
            好，我想被更好地理解
          </router-link>
        </template>
      </section>

      <!-- ② 跟踪快照 -->
      <section class="panel" style="margin-bottom: 14px">
        <h2 class="section-title">② 跟踪快照</h2>
        <template v-if="snapshots.length">
          <div class="stat-row" style="margin-bottom: 12px">
            <div>
              <div class="stat-num">{{ formatAvg(snapshotSummary.avg7) }}</div>
              <div class="muted">近 7 日心里</div>
            </div>
            <div>
              <div class="stat-num">{{ formatAvg(snapshotSummary.avg30) }}</div>
              <div class="muted">近 30 日心里</div>
            </div>
            <div>
              <div class="stat-num">{{ snapshotSummary.count7 || '—' }}</div>
              <div class="muted">近 7 日记下的</div>
            </div>
          </div>
          <p v-if="snapshotSummary.highItems.length" class="muted" style="margin-bottom: 10px">
            偏沉的日子：
            <span v-for="(h, i) in snapshotSummary.highItems" :key="h.id">
              {{ i ? '、' : '' }}{{ formatTime(h.at) }} · {{ h.distressScore }}/10
            </span>
          </p>
          <ul class="checkin-list">
            <li v-for="q in recentSnapshots" :key="q.id">
              <div>
                <strong>心里 {{ q.distressScore }}/10</strong>
                <span class="muted"> · {{ formatTime(q.createdAt || q.at) }}</span>
                <div class="meta">
                  <span>{{ feelingLabels(q.feelings) }}</span>
                  <span v-if="q.takeaway"> · {{ takeawayLabel(q.takeaway) }}</span>
                </div>
              </div>
              <router-link
                v-if="q.relatedCoachSessionId"
                class="btn btn-ghost btn-sm"
                :to="`/coach/${q.relatedCoachSessionId}`"
              >
                那天聊过的
              </router-link>
            </li>
          </ul>
          <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
            <router-link class="btn btn-ghost" to="/wellbeing/quick">再看一眼自己</router-link>
          </div>
        </template>
        <template v-else>
          <p class="muted">还没有留下此刻。认真聊前或你主动看看自己时，会记下一笔；这些点也会进入下一轮对话。</p>
          <router-link class="btn btn-ghost" style="margin-top: 10px" to="/wellbeing/quick">
            花三分钟看看自己
          </router-link>
        </template>
      </section>

      <!-- ③ 职场画像 -->
      <section class="panel">
        <h2 class="section-title">③ 职场画像</h2>
        <template v-if="bigFive">
          <p>
            <strong>{{ bigFive.personaTitle }}</strong>
            —— {{ bigFive.personaBlurb }}
          </p>
          <p v-if="personaBodyText" style="margin-top: 10px; white-space: pre-line">{{ personaBodyText }}</p>
          <div class="tag-cloud" v-if="bigFive.tags?.length" style="margin-top: 8px">
            <span v-for="t in bigFive.tags" :key="t" class="tag-pill">{{ t }}</span>
          </div>
          <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
            <router-link class="btn btn-primary" :to="`/bigfive/${bigFive.id}`">看看详情</router-link>
            <router-link class="btn btn-ghost" to="/bigfive">更新画像</router-link>
            <router-link class="btn btn-ghost" to="/home">带着画像去聊</router-link>
          </div>
        </template>
        <template v-else>
          <p class="muted">两三分钟的趣味小像。有了它，教练会更懂你怎么扛压、怎么沟通；不做诊断，也不挡着你去聊。</p>
          <router-link class="btn btn-ghost" style="margin-top: 10px" to="/bigfive">好，测一下</router-link>
        </template>
      </section>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { api, FEELING_OPTIONS, SCENE_LABEL, TAKEAWAY_OPTIONS } from '../api'
import GuideNote from '../components/GuideNote.vue'

const baselines = ref([])
const snapshots = ref([])
const bigFive = ref(null)
const hasBaseline = ref(false)
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

function formatAvg(v) {
  if (v == null || v === 0) return '—'
  return Number(v).toFixed(1)
}

function headline(a) {
  const raw = a.aiAnalysis
  let obj = raw
  if (typeof raw === 'string') {
    try {
      obj = JSON.parse(raw)
    } catch {
      return a.summaryForCoach?.slice(0, 120) || ''
    }
  }
  return obj?.headline || a.summaryForCoach?.slice(0, 120) || ''
}

function feelingLabels(values) {
  return (values || [])
    .map((v) => FEELING_OPTIONS.find((o) => o.value === v)?.label || v)
    .join('、')
}

function takeawayLabel(v) {
  return TAKEAWAY_OPTIONS.find((o) => o.value === v)?.label || v
}

const latestBaseline = computed(() => baselines.value[0] || null)
const olderBaselines = computed(() => baselines.value.slice(1))
const recentSnapshots = computed(() => snapshots.value.slice(0, 5))

const snapshotSummary = computed(() => {
  const list = snapshots.value || []
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
  return { count7: week.length, avg7: avg(week), avg30: avg(month), highItems }
})

/** 我的评估块③：一段合写，不拆双刃/硬伤两栏 */
const personaBodyText = computed(() => {
  const raw = String(bigFive.value?.personaBody || '').trim()
  if (!raw) return ''
  return raw
    .replace(/^双刃[：:]\s*/m, '')
    .replace(/\n+硬伤[：:]\s*/m, '\n\n')
    .trim()
})

onMounted(async () => {
  try {
    const [assess, quick, bf, me] = await Promise.all([
      api.listAssessments(),
      api.listQuickSelfChecks(),
      api.latestBigFive().catch(() => null),
      api.me().catch(() => null),
    ])
    baselines.value = assess.items || []
    snapshots.value = quick.items || []
    bigFive.value = bf
    hasBaseline.value = Boolean(me?.hasInitialAssessment || baselines.value.length)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
