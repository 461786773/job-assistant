<template>
  <section>
    <div class="hero">
      <h1>你的职场画像</h1>
      <p v-if="profile">
        你更像「{{ profile.personaTitle }}」——{{ profile.personaBlurb }}
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="muted">加载中…</p>

    <template v-else-if="profile">
      <section class="panel bigfive-persona">
        <p class="persona-title">{{ profile.personaTitle }}</p>
        <p class="muted">{{ profile.personaBody || profile.personaBlurb }}</p>
        <div class="tag-cloud" v-if="profile.tags?.length">
          <span v-for="t in profile.tags" :key="t" class="tag-pill">{{ t }}</span>
        </div>
        <p class="muted" style="margin-top: 12px; font-size: 0.86rem">
          这是风格速写，会随阶段变化；不是诊断，也不是能力排名。
        </p>
      </section>

      <section class="panel" style="margin-top: 12px">
        <h2 class="section-title">五维速览</h2>
        <div class="dim-bars">
          <div v-for="d in dims" :key="d.key" class="dim-row">
            <div class="dim-label">
              <strong>{{ d.label }}</strong>
              <span class="muted">{{ bandCN(d.band) }} · {{ d.display }}</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: `${d.display}%` }" />
            </div>
          </div>
        </div>
      </section>

      <div class="row" style="margin-top: 16px; flex-wrap: wrap; gap: 10px">
        <button class="btn btn-primary" type="button" :disabled="starting" @click="startTalk">
          {{ starting ? '正在开门…' : '带着画像去聊聊' }}
        </button>
        <router-link class="btn btn-ghost" to="/home">先去别处转转</router-link>
        <router-link class="btn btn-ghost" to="/bigfive">更新画像</router-link>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const starting = ref(false)
const error = ref('')
const profile = ref(null)

const DIM_META = [
  { key: 'openness', label: '开放性 O' },
  { key: 'conscientiousness', label: '尽责性 C' },
  { key: 'extraversion', label: '外向性 E' },
  { key: 'agreeableness', label: '宜人性 A' },
  { key: 'neuroticism', label: '情绪波动 N' },
]

const dims = computed(() => {
  const scores = parseScores(profile.value?.scores)
  if (!scores) return []
  return DIM_META.map((m) => {
    const d = scores[m.key] || {}
    return {
      key: m.key,
      label: m.label,
      display: Number(d.display) || 0,
      band: d.band || 'mid',
    }
  })
})

function parseScores(raw) {
  if (!raw) return null
  if (typeof raw === 'object') return raw
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function bandCN(b) {
  if (b === 'high') return '偏高'
  if (b === 'low') return '偏低'
  return '中等'
}

async function startTalk() {
  starting.value = true
  error.value = ''
  try {
    const sess = await api.createCoachSession({
      scene: 'job_search',
      mode: 'trial',
      skipQuickGate: true,
    })
    await router.push(`/coach/${sess.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    starting.value = false
  }
}

onMounted(async () => {
  try {
    const id = route.params.id
    if (id) {
      profile.value = await api.getBigFive(id)
    } else {
      profile.value = await api.latestBigFive()
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
