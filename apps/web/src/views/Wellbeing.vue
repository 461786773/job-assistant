<template>
  <section>
    <div class="hero">
      <h1>心理健康跟踪</h1>
      <p>轻量打卡，看见状态与事件的关联。不做临床诊断，数据仅本人可见。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="coach-home-grid">
      <section class="panel">
        <h2 class="section-title">今日打卡</h2>
        <form class="form" @submit.prevent="submit">
          <label>
            压力分（1–10）
            <input v-model.number="form.stressScore" type="number" min="1" max="10" required />
          </label>
          <label>
            精力（可选，1–5）
            <input v-model.number="form.energyScore" type="number" min="0" max="5" />
          </label>
          <fieldset class="tag-fieldset">
            <legend>情绪标签</legend>
            <label v-for="m in MOOD_OPTIONS" :key="m" class="tag-check">
              <input v-model="form.moodTags" type="checkbox" :value="m" />
              {{ m }}
            </label>
          </fieldset>
          <label>
            关联事件
            <select v-model="form.eventType">
              <option v-for="e in EVENT_OPTIONS" :key="e.value" :value="e.value">{{ e.label }}</option>
            </select>
          </label>
          <label>
            一句话备注
            <textarea v-model="form.note" rows="2" placeholder="可选：今天触发点是什么" />
          </label>
          <button class="btn btn-primary" type="submit" :disabled="busy">
            {{ busy ? '保存中…' : '保存打卡' }}
          </button>
        </form>
      </section>

      <section class="panel">
        <h2 class="section-title">趋势摘要</h2>
        <div class="stat-row">
          <div>
            <div class="stat-num">{{ formatAvg(summary?.avgStress7) }}</div>
            <div class="muted">近 7 日压力</div>
          </div>
          <div>
            <div class="stat-num">{{ formatAvg(summary?.avgStress30) }}</div>
            <div class="muted">近 30 日压力</div>
          </div>
        </div>
        <h3 class="pane-sub">高压日对照</h3>
        <ul v-if="summary?.recentHighStress?.length" class="plain-list">
          <li v-for="(h, i) in summary.recentHighStress" :key="i">
            {{ formatTime(h.at) }} · 压力 {{ h.stress }}
            <span v-if="h.eventType"> · {{ eventLabel(h.eventType) }}</span>
            <span v-if="h.note"> — {{ h.note }}</span>
          </li>
        </ul>
        <p v-else class="muted">暂无高压打卡记录。</p>
        <p class="muted" style="margin-top: 12px">连续高压时，请优先寻求专业支持，而不是只依赖本产品。</p>
        <router-link class="btn btn-ghost" to="/home" style="margin-top: 8px">去找教练聊聊</router-link>
      </section>
    </div>

    <section class="panel" style="margin-top: 16px">
      <h2 class="section-title">最近打卡</h2>
      <p v-if="loading" class="muted">加载中…</p>
      <p v-else-if="!items.length" class="muted">还没有记录。</p>
      <ul v-else class="checkin-list">
        <li v-for="c in items" :key="c.id">
          <div>
            <strong>压力 {{ c.stressScore }}</strong>
            <span class="muted"> · {{ formatTime(c.at) }}</span>
            <div class="meta">
              <span v-if="c.moodTags?.length">{{ c.moodTags.join('、') }}</span>
              <span v-if="c.eventType"> · {{ eventLabel(c.eventType) }}</span>
              <span v-if="c.note"> · {{ c.note }}</span>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" type="button" @click="remove(c.id)">删除</button>
        </li>
      </ul>
    </section>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { api, EVENT_OPTIONS, MOOD_OPTIONS } from '../api'

const loading = ref(true)
const busy = ref(false)
const error = ref('')
const items = ref([])
const summary = ref(null)
const form = reactive({
  stressScore: 5,
  energyScore: 0,
  moodTags: [],
  eventType: '',
  note: '',
})

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

function formatAvg(v) {
  if (!v) return '—'
  return Number(v).toFixed(1)
}

function eventLabel(v) {
  return EVENT_OPTIONS.find((e) => e.value === v)?.label || v
}

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.listCheckIns()
    items.value = data.items || []
    summary.value = data.summary || null
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function submit() {
  busy.value = true
  error.value = ''
  try {
    await api.createCheckIn({
      stressScore: form.stressScore,
      energyScore: form.energyScore || 0,
      moodTags: form.moodTags,
      eventType: form.eventType,
      note: form.note,
    })
    form.note = ''
    form.moodTags = []
    await refresh()
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

async function remove(id) {
  if (!confirm('删除这条打卡？')) return
  try {
    await api.deleteCheckIn(id)
    await refresh()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(refresh)
</script>
