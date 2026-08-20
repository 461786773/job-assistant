<template>
  <section>
    <div class="hero">
      <h1>你现在最想谈什么</h1>
      <p>先点一句高亮，再点下面确认——我会按你的选择去开聊。</p>
    </div>

    <GuideNote title="推荐怎么来的">
      <p>
        若你做过心理评估、留下过此刻，或有职场画像，我会合起来建议从哪聊起；都没有也没关系，先选一句最贴的即可。
      </p>
    </GuideNote>

    <p v-if="hint" class="muted">
      结合你留下的资料，我更偏向陪你从「{{ NEED_LABEL[hint] || hint }}」聊起。
    </p>
    <p v-else-if="!hasAssessment && !hasBigFive && !hasQuick" class="muted">
      还没有画像或评估也没关系。之后补上，对话与推荐会更贴你。
    </p>
    <p v-else-if="!hasAssessment" class="muted">
      还没有心理评估也没关系。补上后，我会更清楚你的场景与期望。
    </p>
    <p v-else-if="!hasBigFive" class="muted">
      还没有职场画像也没关系。补测后，接话节奏会更贴你的风格。
    </p>
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
import GuideNote from '../components/GuideNote.vue'

const router = useRouter()
const selected = ref('')
const hint = ref('')
const hasAssessment = ref(false)
const hasBigFive = ref(false)
const hasQuick = ref(false)
const busy = ref(false)
const error = ref('')

const primaryCta = computed(() => {
  if (selected.value === 'unsure') return '好，回首页慢慢找'
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
    if (e.code === 'crisis_elevated') {
      error.value = e.message
      await router.replace({ path: '/booking', query: { crisis: '1' } })
      return
    }
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
}

onMounted(async () => {
  try {
    const me = await api.me()
    hasAssessment.value = Boolean(me.hasInitialAssessment)
    hasBigFive.value = Boolean(me.hasBigFiveProfile)
    hasQuick.value = Boolean(me.hasQuickSnapshot)
    hint.value = me.suggestedNeed || ''
    selected.value = me.primaryNeed || me.suggestedNeed || ''
  } catch (e) {
    error.value = e.message
  }
})
</script>
