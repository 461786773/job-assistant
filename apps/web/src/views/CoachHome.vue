<template>
  <section class="home-warm">
    <header class="home-greet reveal">
      <p class="home-brand-line">求职助手 · 职场心理教练</p>
      <h1 class="home-hello">{{ greeting }}</h1>
      <p class="home-lead">想先随便看看，或直接说一件卡住的事，都可以。</p>
    </header>

    <p v-if="error" class="error">{{ error }}</p>

    <!-- 画像置顶；详细评估软邀请（不挡聊天） -->
    <section class="home-know reveal reveal-delay-1">
      <h2 class="home-know-title">先让我多懂你一点</h2>
      <div class="know-list">
        <div class="know-row">
          <div class="know-row-main">
            <strong>职场画像</strong>
            <span class="know-row-desc">
              <template v-if="profile?.hasBigFiveProfile">
                你更像「{{ profile.bigFivePersonaTitle }}」——随时可以带着它去聊。
              </template>
              <template v-else>两三分钟的趣味小像，不做诊断，也不挡着你去聊。</template>
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
        <div v-if="!profileLoading && profile && !profile.hasInitialAssessment" class="know-row">
          <div class="know-row-main">
            <strong>想让我更懂一点现在的你吗？</strong>
            <span class="know-row-desc">大约几分钟；也可以稍后再说。不做，也能先聊。</span>
          </div>
          <div class="know-row-actions">
            <router-link to="/onboarding/assessment">好，我想被更好地理解</router-link>
          </div>
        </div>
      </div>
    </section>

    <section id="talk" class="home-talk reveal reveal-delay-2" :class="{ 'home-talk-focus': focusTalk }">
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
            <p>近一天你已经看过自己一眼，点下面任意一句，我们就可以认真聊。</p>
          </template>
          <template v-else>
            <p>
              <strong>认真聊之前，想先邀你花三分钟看看自己。</strong>
              不是考试，只是让我更接得住你；轻松聊聊可以跳过。
            </p>
            <div class="formal-gate-actions">
              <router-link
                class="formal-gate-cta"
                :to="{ path: '/wellbeing/quick', query: { next: 'coach', mode: 'formal' } }"
              >
                好，先看看自己
              </router-link>
              <span class="muted">做完后再点下面；或直接点，我会带你过去。</span>
            </div>
          </template>
        </div>
      </template>
      <p v-else class="home-mode-hint muted">轻松聊聊不用先填表，点一句就开始。</p>

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

      <p class="talk-resume">
        <template v-if="sessionsLoading">上次聊到哪了，我看看…</template>
        <template v-else-if="sessions.length">
          上次我们聊到
          <router-link :to="`/coach/${sessions[0].id}`">
            {{ sessions[0].title || NEED_LABEL[sessions[0].scene] || SCENE_LABEL[sessions[0].scene] || sessions[0].scene }}
          </router-link>
          ，也可以从上面另开一句。
        </template>
        <template v-else>还没有开过场。准备好了，点上面任意一句就行。</template>
      </p>
    </section>

    <section class="home-aside reveal reveal-delay-3">
      <p class="aside-status">
        <template v-if="summaryLoading">最近状态我看看…</template>
        <template v-else-if="summary?.checkInCount7">
          近一周你留下了 {{ summary.checkInCount7 }} 笔，压力大约 {{ formatAvg(summary.avgStress7) }}。
          <router-link to="/wellbeing/quick">先记一笔</router-link>
          <span class="dot">·</span>
          <router-link to="/assessments">我的记录</router-link>
        </template>
        <template v-else>
          最近还没有留下什么。
          <router-link to="/wellbeing/quick">先记一笔</router-link>
          <span class="dot">·</span>
          <router-link to="/assessments">我的记录</router-link>
        </template>
      </p>
      <p class="aside-links">
        <router-link to="/booking">预约私教</router-link>
      </p>
    </section>

    <section class="practice-room reveal reveal-delay-4">
      <div class="know-list">
        <div class="know-row">
          <div class="know-row-main">
            <strong>求职练习室</strong>
            <span class="know-row-desc">人事、业务、谈薪——想把「下一句」练熟时再进。</span>
          </div>
          <div class="know-row-actions">
            <router-link to="/tasks">进去看看</router-link>
          </div>
        </div>
      </div>
    </section>
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
import { clearProfileCache } from '../router'

const router = useRouter()
const route = useRoute()
const error = ref('')
const starting = ref(false)
const coachMode = ref(route.query.mode === 'formal' ? 'formal' : 'trial')
const sessions = ref([])
const sessionsLoading = ref(true)
const summary = ref(null)
const summaryLoading = ref(true)
const profile = ref(null)
const profileLoading = ref(true)
const hasRecentQuick = ref(false)
const focusTalk = ref(false)

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

function formatAvg(v) {
  if (!v) return '—'
  return Number(v).toFixed(1)
}

function isRecentQuick(item) {
  if (!item?.createdAt) return false
  const t = Date.parse(item.createdAt)
  if (Number.isNaN(t)) return false
  return Date.now() - t < 24 * 60 * 60 * 1000
}

async function startFromNeed(need) {
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
    if (e.code === 'quick_check_required' || e.status === 409) {
      await router.push({
        path: '/wellbeing/quick',
        query: { next: 'coach', scene: coachSceneForNeed(need), mode: 'formal' },
      })
      return
    }
    error.value = e.message
  } finally {
    starting.value = false
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
    const [me, sessData, checkData] = await Promise.all([
      api.me(),
      api.listCoachSessions(),
      api.listCheckIns(),
      refreshQuickGate(),
    ])
    profile.value = me
    sessions.value = sessData.items || []
    summary.value = checkData.summary || null
  } catch (e) {
    error.value = e.message
  } finally {
    sessionsLoading.value = false
    summaryLoading.value = false
    profileLoading.value = false
  }
})
</script>
