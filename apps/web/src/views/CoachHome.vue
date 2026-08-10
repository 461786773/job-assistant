<template>
  <section>
    <div class="hero">
      <h1>教练工作台</h1>
      <p>
        按主故事线：疏导前先完成问卷；需要时可预约私人辅导；找工作再进三关。
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="coach-home-grid">
      <section class="panel">
        <h2 class="section-title">我的路径</h2>
        <p v-if="profileLoading" class="muted">加载中…</p>
        <template v-else>
          <p>
            当前诉求：
            <strong>{{ NEED_LABEL[profile?.primaryNeed] || '未设置' }}</strong>
          </p>
          <p v-if="profile?.latestAssessmentAt" class="muted">
            最近评测：{{ formatTime(profile.latestAssessmentAt) }}
          </p>
          <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 8px">
            <router-link class="btn btn-ghost" to="/assessments">我的评估</router-link>
            <router-link class="btn btn-ghost" to="/booking">私人辅导预约</router-link>
            <router-link class="btn btn-ghost" to="/onboarding/need">调整诉求</router-link>
          </div>
        </template>
      </section>

      <section class="panel">
        <h2 class="section-title">本周状态</h2>
        <p v-if="summaryLoading" class="muted">加载中…</p>
        <template v-else>
          <div class="stat-row">
            <div>
              <div class="stat-num">{{ formatAvg(summary?.avgStress7) }}</div>
              <div class="muted">近 7 日平均压力</div>
            </div>
            <div>
              <div class="stat-num">{{ summary?.checkInCount7 || 0 }}</div>
              <div class="muted">近 7 日打卡</div>
            </div>
          </div>
          <div class="row" style="margin-top: 14px; flex-wrap: wrap; gap: 8px">
            <router-link class="btn btn-primary" to="/wellbeing/quick">三分钟自评</router-link>
            <router-link class="btn btn-ghost" to="/wellbeing">心理跟踪</router-link>
          </div>
        </template>
      </section>
    </div>

    <section class="panel" style="margin-top: 16px">
      <h2 class="section-title">AI 心理疏导</h2>
      <p class="muted">进入疏导前需在近 24 小时内完成三分钟自评（问卷门禁）。</p>
      <div class="scene-grid">
        <button
          v-for="s in visibleScenes"
          :key="s.value"
          class="scene-card"
          type="button"
          :disabled="starting"
          @click="startCoach(s.value)"
        >
          <strong>{{ s.label }}</strong>
          <span>{{ s.desc }}</span>
        </button>
      </div>
    </section>

    <section class="panel" style="margin-top: 16px">
      <div class="workspace-toolbar">
        <h2 class="section-title" style="margin: 0">最近教练会话</h2>
        <router-link class="btn btn-ghost btn-sm" to="/booking">去预约</router-link>
      </div>
      <p v-if="sessionsLoading" class="muted">加载中…</p>
      <p v-else-if="!sessions.length" class="muted">还没有会话。从上方场景开始第一次疏导。</p>
      <div v-else class="grid" style="margin-top: 12px">
        <router-link
          v-for="s in sessions"
          :key="s.id"
          :to="`/coach/${s.id}`"
          class="task-card"
        >
          <div class="task-card-main">
            <h3>{{ s.title || SCENE_LABEL[s.scene] || s.scene }}</h3>
            <div class="meta">
              <div>{{ SCENE_LABEL[s.scene] || s.scene }} · {{ s.status === 'done' ? '已结束' : '进行中' }}</div>
              <div>更新于 {{ formatTime(s.updatedAt) }}</div>
            </div>
          </div>
          <span class="badge">{{ s.crisisFlag ? '已转介' : s.status === 'done' ? '完成' : '进行中' }}</span>
        </router-link>
      </div>
    </section>

    <section v-if="showJobGates" class="panel" style="margin-top: 16px">
      <div class="workspace-toolbar">
        <h2 class="section-title" style="margin: 0">求职过关训练</h2>
        <router-link class="btn btn-primary btn-sm" to="/tasks/new">新建任务</router-link>
      </div>
      <p class="muted">人事 → 业务 → 谈薪。需要练表达时再进。</p>
      <p v-if="tasksLoading" class="muted">加载中…</p>
      <p v-else-if="!tasks.length" class="muted">暂无求职任务。<router-link to="/tasks">查看全部</router-link></p>
      <div v-else class="grid" style="margin-top: 12px">
        <router-link
          v-for="t in tasks.slice(0, 4)"
          :key="t.id"
          :to="`/tasks/${t.id}`"
          class="task-card"
        >
          <div class="task-card-main">
            <h3>{{ t.title || '未命名任务' }}</h3>
            <div class="meta">
              <div>{{ [t.company, t.targetRole].filter(Boolean).join(' · ') || '未填公司/岗位' }}</div>
            </div>
          </div>
          <span class="badge">{{ STATUS_LABEL[t.status] || t.status }}</span>
        </router-link>
      </div>
      <div class="row" style="margin-top: 12px">
        <router-link class="btn btn-ghost" to="/tasks">全部过关任务</router-link>
      </div>
    </section>

    <section v-else class="panel" style="margin-top: 16px">
      <h2 class="section-title">分支服务</h2>
      <p class="muted">
        当前诉求更偏{{ NEED_LABEL[profile?.primaryNeed] || '疏导' }}。需要找工作时可
        <router-link to="/tasks">进入求职三关</router-link>
        ，或先继续上方场景疏导。
      </p>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, NEED_LABEL, SCENE_LABEL, STATUS_LABEL } from '../api'

const router = useRouter()
const error = ref('')
const starting = ref(false)
const sessions = ref([])
const sessionsLoading = ref(true)
const tasks = ref([])
const tasksLoading = ref(true)
const summary = ref(null)
const summaryLoading = ref(true)
const profile = ref(null)
const profileLoading = ref(true)

const allScenes = [
  { value: 'job_search', label: '求职 / 跳槽', desc: '挂面、投递耗竭、在职偷投焦虑' },
  { value: 'promotion', label: '晋升 / 述职', desc: '述职压力、与上级预期错位' },
  { value: 'communication', label: '沟通 / 冲突', desc: '会后内耗、边界与诉求' },
]

const visibleScenes = computed(() => {
  const need = profile.value?.primaryNeed
  if (need === 'promotion') {
    return [allScenes[1], allScenes[0], allScenes[2]]
  }
  if (need === 'communication') {
    return [allScenes[2], allScenes[1], allScenes[0]]
  }
  if (need === 'job_search') {
    return allScenes
  }
  return allScenes
})

const showJobGates = computed(() => {
  const need = profile.value?.primaryNeed
  return !need || need === 'job_search' || need === 'unsure'
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

async function startCoach(scene) {
  starting.value = true
  error.value = ''
  try {
    const sess = await api.createCoachSession({ scene })
    router.push(`/coach/${sess.id}`)
  } catch (e) {
    if (e.code === 'quick_check_required' || e.status === 409) {
      router.push({
        path: '/wellbeing/quick',
        query: { next: 'coach', scene },
      })
      return
    }
    error.value = e.message
  } finally {
    starting.value = false
  }
}

onMounted(async () => {
  try {
    const [me, sessData, taskData, checkData] = await Promise.all([
      api.me(),
      api.listCoachSessions(),
      api.listTasks(),
      api.listCheckIns(),
    ])
    profile.value = me
    sessions.value = sessData.items || []
    tasks.value = taskData.items || []
    summary.value = checkData.summary || null
  } catch (e) {
    error.value = e.message
  } finally {
    sessionsLoading.value = false
    tasksLoading.value = false
    summaryLoading.value = false
    profileLoading.value = false
  }
})
</script>
