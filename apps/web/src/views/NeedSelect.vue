<template>
  <section>
    <div class="hero">
      <h1>你现在最想谈什么</h1>
      <p>点一句就好，我会直接带你去聊。说不清也没关系。</p>
    </div>

    <p v-if="hint" class="muted">评测里，我更偏向陪你从「{{ NEED_LABEL[hint] || hint }}」聊起。</p>
    <p v-else-if="!hasAssessment" class="muted">还没有做过详细评估也没关系，先选一句最贴的即可。</p>
    <p v-if="error" class="error">{{ error }}</p>

    <div class="scene-grid" style="margin-top: 14px">
      <button
        v-for="o in NEED_OPTIONS"
        :key="o.value"
        class="scene-card"
        type="button"
        :class="{ selected: selected === o.value }"
        :disabled="busy"
        @click="choose(o.value)"
      >
        <strong>{{ o.label }}</strong>
        <span>{{ o.desc }}</span>
      </button>
    </div>

    <div class="row" style="margin-top: 16px; flex-wrap: wrap; gap: 10px">
      <button class="btn btn-primary" type="button" :disabled="busy || !selected" @click="confirm">
        {{ busy ? '正在带你过去…' : primaryCta }}
      </button>
      <router-link class="btn btn-ghost" to="/home">先四处看看</router-link>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, coachSceneForNeed, NEED_LABEL, NEED_OPTIONS } from '../api'
import { clearProfileCache } from '../router'

const router = useRouter()
const selected = ref('')
const hint = ref('')
const hasAssessment = ref(false)
const busy = ref(false)
const error = ref('')

const primaryCta = computed(() => {
  if (selected.value === 'unsure') return '好，回工作台慢慢找'
  if (selected.value === 'counsel_first') return '好，先把心安下来'
  return '好，开始聊'
})

async function startCoach(scene) {
  try {
    const sess = await api.createCoachSession({
      scene,
      mode: 'trial',
      skipQuickGate: true,
    })
    await router.replace(`/coach/${sess.id}`)
  } catch (e) {
    if (e.code === 'quick_check_required' || e.status === 409) {
      await router.replace({
        path: '/wellbeing/quick',
        query: { next: 'coach', scene, mode: 'formal' },
      })
      return
    }
    throw e
  }
}

async function confirm() {
  if (!selected.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    await api.setPrimaryNeed(selected.value)
    clearProfileCache()
    // 产品方案 §0.5-B：unsure → 回工作台并聚焦「想聊聊吗」
    if (selected.value === 'unsure') {
      await router.replace({ path: '/home', query: { focus: 'talk' } })
      return
    }
    await startCoach(coachSceneForNeed(selected.value))
  } catch (e) {
    error.value = e.message || '没带过去，请再试一次'
  } finally {
    busy.value = false
  }
}

function choose(value) {
  selected.value = value
  confirm()
}

onMounted(async () => {
  try {
    const me = await api.me()
    hasAssessment.value = Boolean(me.hasInitialAssessment)
    hint.value = me.suggestedNeed || ''
    selected.value = me.primaryNeed || me.suggestedNeed || ''
  } catch (e) {
    error.value = e.message
  }
})
</script>
