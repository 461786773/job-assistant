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

      <div v-if="session.crisisFlag" class="crisis-banner">
        {{ CRISIS_HELP }}
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
import { api, CRISIS_HELP, SCENE_LABEL } from '../api'

const route = useRoute()
const session = ref(null)
const loading = ref(true)
const error = ref('')
const draft = ref('')
const busy = ref(false)
const sendError = ref('')

const sceneHuman = computed(() => {
  const map = {
    job_search: '求职这件事',
    promotion: '晋升与述职',
    communication: '沟通与冲突',
  }
  const scene = session.value?.scene
  return map[scene] || SCENE_LABEL[scene] || session.value?.title || '我们正在聊'
})

const isTrial = computed(() => (session.value?.title || '').includes('轻试'))

async function load() {
  loading.value = true
  error.value = ''
  try {
    session.value = await api.getCoachSession(route.params.id)
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
