<template>
  <section class="coach-session session-warm">
    <div v-if="loading" class="muted">我在准备…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <template v-else-if="session">
      <div class="session-top">
        <div>
          <p class="session-scene">{{ sceneHuman }}</p>
          <p class="muted">
            {{ session.status === 'done' || session.crisisFlag ? '这轮先停在这里' : '我们正在聊' }}
            <span v-if="isTrial"> · 轻松聊聊</span>
          </p>
        </div>
        <router-link class="btn btn-ghost btn-sm" to="/home">先回到安静的一页</router-link>
      </div>

      <GuideNote
        v-if="session.status !== 'done' && !session.crisisFlag"
        title="这轮对话会怎么接你"
        tone="calm"
      >
        <p>{{ contextGuide }}</p>
        <div class="context-chips">
          <span class="context-chip" :class="hasBigFive ? 'on' : 'off'">
            职场画像{{ hasBigFive ? ' · 已带上' : ' · 暂无' }}
          </span>
          <span class="context-chip" :class="hasAssessment ? 'on' : 'off'">
            心理评估{{ hasAssessment ? ' · 已带上' : ' · 暂无' }}
          </span>
          <span class="context-chip" :class="hasQuick ? 'on' : 'off'">
            此刻记录{{ hasQuick ? ' · 已带上' : ' · 暂无' }}
          </span>
        </div>
        <p v-if="!hasBigFive || !hasAssessment || !hasQuick" style="margin-top: 8px">
          缺的可以稍后补：
          <router-link v-if="!hasBigFive" to="/bigfive">测画像</router-link>
          <template v-if="!hasBigFive && (!hasAssessment || !hasQuick)"> · </template>
          <router-link v-if="!hasAssessment" to="/onboarding/assessment">做心理评估</router-link>
          <template v-if="!hasAssessment && !hasQuick"> · </template>
          <router-link v-if="!hasQuick" to="/wellbeing/quick">看一眼自己</router-link>
        </p>
      </GuideNote>

      <div v-if="session.crisisFlag" class="crisis-banner">
        {{ crisisHelp }}
      </div>

      <div class="coach-layout">
        <div class="chat-panel session-chat">
          <div class="chat-log">
            <div
              v-for="(m, i) in session.messages || []"
              :key="i"
              class="chat-bubble letter"
              :class="m.role"
            >
              <div class="chat-role">{{ m.role === 'user' ? '我' : '教练' }}</div>
              <div class="chat-content">{{ m.content }}</div>
            </div>
          </div>
          <form v-if="!session.crisisFlag && session.status !== 'done'" class="chat-input" @submit.prevent="send">
            <textarea
              v-model="draft"
              rows="3"
              placeholder="此刻最卡住的，是一件事，还是一种感觉？"
              :disabled="busy"
            />
            <div class="row">
              <button class="btn btn-primary" type="submit" :disabled="busy || !draft.trim()">
                {{ busy ? '我在听…' : '告诉教练' }}
              </button>
              <button class="btn btn-ghost" type="button" :disabled="busy" @click="finish">先说到这里</button>
            </div>
            <p v-if="sendError" class="error">{{ sendError }}</p>
          </form>
          <p v-else class="muted end-note">
            这轮先停在这里。若还想练表达或投递，可以从旁边慢慢选；也可以就这样休息。
          </p>
        </div>

        <aside class="side-panel session-side">
          <h3>今天可以带走的</h3>
          <ul v-if="session.actionItems?.length" class="plain-list">
            <li v-for="(a, i) in session.actionItems" :key="i">{{ a }}</li>
          </ul>
          <p v-else class="muted">我们再往下聊几句，会一起找出一件今天就能做的小事。</p>

          <h3 style="margin-top: 18px">下一句，可以这样开口</h3>
          <ul v-if="session.scripts?.length" class="plain-list">
            <li v-for="(s, i) in session.scripts" :key="i">{{ s }}</li>
          </ul>
          <p v-else class="muted">需要开口时，会写在这里。</p>

          <h3 style="margin-top: 18px">若你想换种准备方式</h3>
          <div v-if="gateHint" class="gate-suggest">
            <p>{{ gateHint.text }}</p>
            <router-link class="btn btn-primary btn-sm" :to="gateHint.to">{{ gateHint.cta }}</router-link>
          </div>
          <div class="row" style="flex-wrap: wrap">
            <router-link
              v-if="session.relatedTaskId"
              class="btn btn-ghost btn-sm"
              :to="`/tasks/${session.relatedTaskId}`"
            >
              打开这份练习
            </router-link>
            <router-link class="btn btn-ghost btn-sm" to="/tasks">去练习室看看</router-link>
            <router-link class="btn btn-ghost btn-sm" to="/wellbeing/quick">留下今天的状态</router-link>
            <router-link class="btn btn-ghost btn-sm" to="/booking">预约私教</router-link>
          </div>
        </aside>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, SCENE_LABEL } from '../api'
import { crisisHelp } from '../copy'
import GuideNote from '../components/GuideNote.vue'

const route = useRoute()
const session = ref(null)
const loading = ref(true)
const error = ref('')
const draft = ref('')
const busy = ref(false)
const sendError = ref('')
const hasBigFive = ref(false)
const hasAssessment = ref(false)
const hasQuick = ref(false)

const sceneHuman = computed(() => {
  const map = {
    job_search: '求职这件事',
    promotion: '晋升与述职',
    communication: '沟通与冲突',
  }
  const scene = session.value?.scene
  return map[scene] || SCENE_LABEL[scene] || session.value?.title || '我们正在聊'
})

const isTrial = computed(() => {
  const t = session.value?.title || ''
  return t.includes('轻松聊聊') || t.includes('轻试')
})

const contextGuide = computed(() => {
  const n = [hasBigFive.value, hasAssessment.value, hasQuick.value].filter(Boolean).length
  if (n === 3) {
    return '我会合读你的画像风格、最新心理评估，以及此刻记录，再接话与给下一步——不是泛泛安慰。'
  }
  if (n > 0) {
    return '我会尽量合读你已留下的画像 / 评估 / 此刻记录；缺的部分不影响开聊，补上后会更贴。'
  }
  return '你还没留下画像、评估或此刻记录。也能聊；若想让我更接得住，可以会后再补。'
})

const gateHint = computed(() => {
  const g = session.value?.suggestGate
  const taskQ = session.value?.relatedTaskId
    ? { path: `/tasks/${session.value.relatedTaskId}`, query: { gate: g } }
    : { path: '/tasks/new', query: { gate: g } }
  if (g === 'hr') {
    return { text: '聊到这里，或许可以去练一练「人事关」——把简历主线说清楚。', cta: '去练人事关', to: taskQ }
  }
  if (g === 'interview') {
    return { text: '表达可以再压一压。练习室里有业务/面试关，把下一句练熟。', cta: '去练业务关', to: taskQ }
  }
  if (g === 'salary') {
    return { text: '若卡在开口谈钱，可以去谈薪关对一下结构与话术。', cta: '去练谈薪关', to: taskQ }
  }
  return null
})

async function loadProfileHints() {
  try {
    const me = await api.me()
    hasBigFive.value = Boolean(me.hasBigFiveProfile)
    hasAssessment.value = Boolean(me.hasInitialAssessment)
    hasQuick.value = Boolean(me.hasQuickSnapshot)
  } catch {
    hasBigFive.value = false
    hasAssessment.value = false
    hasQuick.value = false
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [sess] = await Promise.all([
      api.getCoachSession(route.params.id),
      loadProfileHints(),
    ])
    session.value = sess
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function send() {
  sendError.value = ''
  busy.value = true
  try {
    session.value = await api.replyCoachSession(route.params.id, draft.value.trim())
    draft.value = ''
  } catch (e) {
    if (e.code === 'crisis_elevated') {
      sendError.value = e.message || '当前不宜继续常规陪聊'
      session.value = {
        ...session.value,
        crisisFlag: true,
        status: 'done',
      }
      return
    }
    sendError.value = e.message
  } finally {
    busy.value = false
  }
}

async function finish() {
  if (!draft.value.trim()) {
    draft.value = '我想先结束这一轮，请帮我收成一个今天就能做的动作。'
  }
  await send()
}

watch(() => route.params.id, load)
onMounted(load)
</script>

<style scoped>
.gate-suggest {
  margin: 0 0 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-2, #eef3ef) 85%, transparent);
}
.gate-suggest p {
  margin: 0 0 10px;
  line-height: 1.5;
}
</style>
