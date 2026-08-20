<template>
  <section class="home-warm">
    <header class="home-greet reveal">
      <p class="home-brand-line">求职助手 · 职场心理教练</p>
      <h1 class="home-hello">{{ greeting }}</h1>
      <p class="home-lead">想说一件卡住的事，点下面就行。</p>
    </header>

    <p v-if="error" class="error">{{ error }}</p>

    <div v-if="crisisElevated" class="crisis-banner reveal">
      <p>{{ crisisHelp }}</p>
      <p style="margin-top: 10px">
        当前心理评估标出了需要被认真对待的信号，不宜继续常规 AI 陪聊。
        <router-link to="/booking">去预约私教</router-link>
      </p>
    </div>

    <section
      v-if="!crisisElevated"
      id="talk"
      class="home-talk reveal reveal-delay-1"
      :class="{ 'home-talk-focus': focusTalk }"
    >
      <div class="home-talk-head">
        <h2>想聊聊吗</h2>
        <div class="mode-toggle">
          <button
            type="button"
            :class="{ active: coachMode === 'trial' }"
            @click="coachMode = 'trial'"
          >
            先轻松聊聊
          </button>
          <button
            type="button"
            :class="{ active: coachMode === 'formal' }"
            @click="coachMode = 'formal'"
          >
            认真聊一轮
          </button>
        </div>
      </div>

      <template v-if="coachMode === 'formal'">
        <div class="formal-gate">
          <template v-if="hasRecentQuick">
            <p>
              今天你已经留下过此刻记录——可以直接开聊；想换个说法再记一笔也可以。
            </p>
            <div class="formal-gate-actions">
              <button
                class="formal-gate-cta"
                type="button"
                :disabled="starting"
                @click="startWithTodaySnapshot"
              >
                {{ starting ? '正在开门…' : '用今天的记录开聊' }}
              </button>
              <router-link
                class="btn btn-ghost btn-sm"
                :to="{ path: '/wellbeing/quick', query: { next: 'coach', mode: 'formal' } }"
              >
                再测此刻
              </router-link>
            </div>
          </template>
          <template v-else>
            <p>
              <strong>认真聊之前，想先邀你花三分钟看看自己。</strong>
              这条此刻记录会进入本轮对话；轻松聊聊可以跳过。
            </p>
            <div class="formal-gate-actions">
              <router-link
                class="formal-gate-cta"
                :to="{ path: '/wellbeing/quick', query: { next: 'coach', mode: 'formal' } }"
              >
                好，先看看自己
              </router-link>
              <span class="muted">做完后再点下面；或直接点场景，我会带你过去。</span>
            </div>
          </template>
        </div>
      </template>
      <p v-else class="home-mode-hint muted">
        轻松聊聊不用先填表。若你已有画像或评估，我仍会带上它们一起听。
      </p>

      <div class="scene-list" role="list">
        <button
          v-for="s in talkOptions"
          :key="s.value"
          class="scene-row"
          type="button"
          role="listitem"
          :disabled="starting"
          @click="startFromNeed(s.value)"
        >
          <span class="scene-row-main">
            <strong>{{ s.label }}</strong>
            <span class="scene-row-desc">{{ s.desc }}</span>
          </span>
          <span class="scene-row-action">聊</span>
        </button>
      </div>
    </section>

    <!-- P1-1 会话列表 -->
    <section class="home-sessions reveal reveal-delay-2">
      <h2 class="home-know-title">最近聊过</h2>
      <p v-if="sessionsLoading" class="muted">上次聊到哪了，我看看…</p>
      <template v-else-if="sessions.length">
        <ul class="session-list">
          <li v-for="s in sessions.slice(0, 8)" :key="s.id">
            <router-link :to="`/coach/${s.id}`" class="session-list-main">
              <strong>{{ sessionTitle(s) }}</strong>
              <span class="muted">{{ formatTime(s.updatedAt || s.createdAt) }}</span>
            </router-link>
            <button
              class="btn btn-ghost btn-sm"
              type="button"
              :disabled="deletingId === s.id"
              @click="removeSession(s.id)"
            >
              删除
            </button>
          </li>
        </ul>
      </template>
      <p v-else class="muted">还没有开过场。准备好了，点上面任意一句就行。</p>
    </section>

    <details class="home-materials reveal reveal-delay-3">
      <summary>让对话更贴你</summary>
      <p class="muted home-materials-lead">
        有哪项用哪项，缺了也能聊。完整回看在
        <router-link to="/assessments">我的</router-link>。
      </p>
      <div class="know-list">
        <div class="know-row">
          <div class="know-row-main">
            <strong>职场画像</strong>
            <span class="know-row-desc">
              <template v-if="profile?.hasBigFiveProfile">
                你更像「{{ profile.bigFivePersonaTitle }}」
              </template>
              <template v-else>两三分钟小像，用来调接话节奏。</template>
            </span>
          </div>
          <div class="know-row-actions">
            <router-link
              v-if="profile?.hasBigFiveProfile"
              :to="`/bigfive/${profile.latestBigFiveId}`"
            >
              看看
            </router-link>
            <router-link v-else to="/bigfive">测一下</router-link>
          </div>
        </div>
        <div class="know-row">
          <div class="know-row-main">
            <strong>心理评估</strong>
            <span class="know-row-desc">
              <template v-if="profile?.hasInitialAssessment">已有一份底，近况变了可以更新。</template>
              <template v-else>场景与期望；想被更深接住时再做。</template>
            </span>
          </div>
          <div class="know-row-actions">
            <router-link v-if="profile?.hasInitialAssessment" to="/onboarding/assessment">
              重新评估一下
            </router-link>
            <router-link v-else to="/onboarding/assessment">去做</router-link>
          </div>
        </div>
        <div class="know-row">
          <div class="know-row-main">
            <strong>此刻记录</strong>
            <span class="know-row-desc">
              <template v-if="hasRecentQuick">今天已经留过一笔。</template>
              <template v-else>认真聊前会用到；轻松聊聊可以跳过。</template>
            </span>
          </div>
          <div class="know-row-actions">
            <router-link to="/wellbeing/quick">再测此刻</router-link>
          </div>
        </div>
      </div>
    </details>

    <div class="home-more reveal reveal-delay-4">
      <p class="home-more-label muted">需要时，走这两条支线</p>
      <div class="home-more-actions">
        <router-link class="home-more-card" to="/booking">
          <strong>预约私教</strong>
          <span>想找真人聊一轮</span>
        </router-link>
        <router-link class="home-more-card" to="/tasks">
          <strong>练习室</strong>
          <span>练几道职场小关</span>
        </router-link>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  api,
  coachSceneForNeed,
  NEED_LABEL,
  NEED_OPTIONS,
  SCENE_LABEL,
} from '../api'
import { crisisHelp } from '../copy'
import { clearProfileCache } from '../router'

const router = useRouter()
const route = useRoute()
const error = ref('')
const starting = ref(false)
const deletingId = ref('')
const coachMode = ref(route.query.mode === 'formal' ? 'formal' : 'trial')
const sessions = ref([])
const sessionsLoading = ref(true)
const profile = ref(null)
const hasRecentQuick = ref(false)
const focusTalk = ref(false)
const crisisElevated = ref(false)

const talkOptions = computed(() => {
  const need = profile.value?.primaryNeed
  const list = [...NEED_OPTIONS]
  if (!need) return list
  const idx = list.findIndex((o) => o.value === need)
  if (idx <= 0) return list
  const [hit] = list.splice(idx, 1)
  return [hit, ...list]
})

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 11) return '早上好'
  if (h < 14) return '你好，又见面了'
  if (h < 19) return '下午好'
  return '今晚还好吗'
})

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN', {
      month: 'numeric',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

function sessionTitle(s) {
  return s.title || NEED_LABEL[s.scene] || SCENE_LABEL[s.scene] || s.scene || '一席对话'
}

function isRecentQuick(item) {
  const raw = item?.createdAt || item?.at
  if (!raw) return false
  const t = Date.parse(raw)
  if (Number.isNaN(t)) return false
  // 与后端一致：按本地日历日；后端用 Asia/Shanghai，前端用本机时区（国内部署通常一致）
  const a = new Date(t)
  const b = new Date()
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function handleCoachError(e, need) {
  if (e.code === 'crisis_elevated') {
    crisisElevated.value = true
    error.value = e.message
    return router.push({ path: '/booking', query: { crisis: '1' } })
  }
  if (e.code === 'quick_check_required' || e.status === 409) {
    return router.push({
      path: '/wellbeing/quick',
      query: { next: 'coach', scene: coachSceneForNeed(need), mode: 'formal' },
    })
  }
  error.value = e.message
}

async function startFromNeed(need) {
  if (crisisElevated.value) {
    error.value = '当前不宜继续常规陪聊，请先查看求助资源或预约私教。'
    return
  }
  starting.value = true
  error.value = ''
  try {
    await api.setPrimaryNeed(need)
    clearProfileCache()
    if (profile.value) profile.value.primaryNeed = need

    const scene = coachSceneForNeed(need)
    if (coachMode.value === 'formal' && !hasRecentQuick.value) {
      await router.push({
        path: '/wellbeing/quick',
        query: { next: 'coach', scene, mode: 'formal' },
      })
      return
    }
    const sess = await api.createCoachSession({
      scene,
      mode: coachMode.value,
      skipQuickGate: coachMode.value === 'trial',
    })
    await router.push(`/coach/${sess.id}`)
  } catch (e) {
    await handleCoachError(e, need)
  } finally {
    starting.value = false
  }
}

/** 沿用今日快照：用推荐诉求或默认场景直接开认真聊 */
async function startWithTodaySnapshot() {
  const need = profile.value?.primaryNeed || talkOptions.value[0]?.value || 'job_search'
  await startFromNeed(need)
}

async function removeSession(id) {
  if (!confirm('删除这场对话？')) return
  deletingId.value = id
  error.value = ''
  try {
    await api.deleteCoachSession(id)
    sessions.value = sessions.value.filter((s) => s.id !== id)
  } catch (e) {
    error.value = e.message
  } finally {
    deletingId.value = ''
  }
}

async function refreshQuickGate() {
  try {
    const quick = await api.listQuickSelfChecks()
    const items = quick.items || []
    hasRecentQuick.value = items.some(isRecentQuick)
  } catch {
    hasRecentQuick.value = false
  }
}

watch(
  () => route.query.focus,
  async (focus) => {
    if (focus === 'talk') {
      focusTalk.value = true
      await nextTick()
      document.getElementById('talk')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  },
  { immediate: true },
)

onMounted(async () => {
  try {
    const [me, sessData] = await Promise.all([
      api.me(),
      api.listCoachSessions(),
      refreshQuickGate(),
    ])
    profile.value = me
    crisisElevated.value = me?.crisisLevel === 'elevated'
    sessions.value = sessData.items || []
  } catch (e) {
    error.value = e.message
  } finally {
    sessionsLoading.value = false
  }
})
</script>
