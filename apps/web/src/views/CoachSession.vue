<template>
  <section class="coach-session">
    <div v-if="loading" class="muted">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <template v-else-if="session">
      <div class="hero hero-compact">
        <h1>{{ session.title || SCENE_LABEL[session.scene] }}</h1>
        <p>
          {{ SCENE_LABEL[session.scene] || session.scene }}
          <span v-if="session.relatedEvent"> · {{ session.relatedEvent }}</span>
          · {{ session.status === 'done' ? '已结束' : '进行中' }}
        </p>
      </div>

      <div v-if="session.crisisFlag" class="crisis-banner">
        {{ CRISIS_HELP }}
      </div>

      <div class="coach-layout">
        <div class="panel chat-panel">
          <div class="chat-log">
            <div
              v-for="(m, i) in session.messages || []"
              :key="i"
              class="chat-bubble"
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
              placeholder="说说发生了什么，以及此刻最卡住的感受…"
              :disabled="busy"
            />
            <div class="row">
              <button class="btn btn-primary" type="submit" :disabled="busy || !draft.trim()">
                {{ busy ? '思考中…' : '发送' }}
              </button>
              <button class="btn btn-ghost" type="button" :disabled="busy" @click="finish">结束本轮</button>
            </div>
            <p v-if="sendError" class="error">{{ sendError }}</p>
          </form>
          <p v-else class="muted">本轮已结束。可回到工作台新开场景，或去做过关训练。</p>
        </div>

        <aside class="panel side-panel">
          <h3>行动卡</h3>
          <ul v-if="session.actionItems?.length" class="plain-list">
            <li v-for="(a, i) in session.actionItems" :key="i">{{ a }}</li>
          </ul>
          <p v-else class="muted">对话推进后会出现 24h 可执行动作。</p>

          <h3 style="margin-top: 18px">下一句怎么说</h3>
          <ul v-if="session.scripts?.length" class="plain-list">
            <li v-for="(s, i) in session.scripts" :key="i">{{ s }}</li>
          </ul>
          <p v-else class="muted">需要话术时会写在这里。</p>

          <h3 style="margin-top: 18px">需要过关时</h3>
          <div class="row" style="flex-wrap: wrap">
            <router-link
              v-if="session.relatedTaskId"
              class="btn btn-primary btn-sm"
              :to="`/tasks/${session.relatedTaskId}`"
            >
              打开关联任务
            </router-link>
            <router-link class="btn btn-ghost btn-sm" to="/tasks">过关任务列表</router-link>
            <router-link class="btn btn-ghost btn-sm" to="/wellbeing">记一笔打卡</router-link>
          </div>
          <div class="row" style="margin-top: 12px">
            <router-link class="btn btn-ghost btn-sm" to="/home">返回教练工作台</router-link>
          </div>
        </aside>
      </div>
    </template>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, CRISIS_HELP, SCENE_LABEL } from '../api'

const route = useRoute()
const session = ref(null)
const loading = ref(true)
const error = ref('')
const draft = ref('')
const busy = ref(false)
const sendError = ref('')

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
