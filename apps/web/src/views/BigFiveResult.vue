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
        <p class="persona-hook">{{ profile.personaBlurb }}</p>
        <div class="persona-split">
          <div class="persona-block">
            <span class="persona-kicker">双刃</span>
            <p>{{ personaParts.edge }}</p>
          </div>
          <div class="persona-block persona-block-shadow" v-if="personaParts.shadow">
            <span class="persona-kicker">硬伤</span>
            <p>{{ personaParts.shadow }}</p>
          </div>
        </div>
        <div class="tag-cloud" v-if="profile.tags?.length">
          <span v-for="t in profile.tags" :key="t" class="tag-pill">{{ t }}</span>
        </div>
        <p class="muted" style="margin-top: 12px; font-size: 0.86rem">
          会刺一点，才有用。风格速写不是定论，不是诊断，也不是能力排名。
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

const personaParts = computed(() => {
  const body = String(profile.value?.personaBody || '').trim()
  if (!body) {
    return { edge: profile.value?.personaBlurb || '', shadow: '' }
  }
  const marker = '\n\n硬伤：'
  const idx = body.indexOf(marker)
  if (idx >= 0) {
    let edge = body.slice(0, idx).trim()
    if (edge.startsWith('双刃：')) edge = edge.slice('双刃：'.length).trim()
    return {
      edge,
      shadow: body.slice(idx + marker.length).trim(),
    }
  }
  const alt = body.indexOf('硬伤：')
  if (alt >= 0) {
    let edge = body.slice(0, alt).trim()
    if (edge.startsWith('双刃：')) edge = edge.slice('双刃：'.length).trim()
    return {
      edge,
      shadow: body.slice(alt + '硬伤：'.length).trim(),
    }
  }
  return { edge: body, shadow: '' }
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

<style scoped>
.persona-hook {
  margin: 0 0 14px;
  font-size: 1.05rem;
  line-height: 1.45;
  color: var(--text, #1c241f);
}
.persona-split {
  display: grid;
  gap: 10px;
}
.persona-block {
  padding: 12px 14px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-2, #eef3ef) 80%, transparent);
}
.persona-block p {
  margin: 6px 0 0;
  line-height: 1.55;
}
.persona-block-shadow {
  background: color-mix(in srgb, #c45c3e 10%, var(--surface, #fff));
  border: 1px solid color-mix(in srgb, #c45c3e 28%, transparent);
}
.persona-kicker {
  display: inline-block;
  font-size: 0.75rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted, #66726a);
  font-weight: 600;
}
.persona-block-shadow .persona-kicker {
  color: #a3472e;
}
</style>
