<template>
  <section>
    <div class="hero">
      <h1>教练工作台</h1>
      <p>职场心理教练陪你过求职、晋升、沟通节点；需要时再调用人事 / 业务 / 谈薪过关训练。</p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="coach-home-grid">
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
          <p v-if="!summary?.checkInCount7" class="muted">还没有打卡。高压节点后记一笔，方便回看触发点。</p>
          <div class="row" style="margin-top: 14px">
            <router-link class="btn btn-ghost" to="/wellbeing">心理跟踪</router-link>
          </div>
        </template>
      </section>

      <section class="panel">
        <h2 class="section-title">开始教练</h2>
        <p class="muted">选一个当前节点，进入澄清 → 行动。</p>
        <div class="scene-grid">
          <button
            v-for="s in scenes"
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
    </div>

    <section class="panel" style="margin-top: 16px">
      <div class="workspace-toolbar">
        <h2 class="section-title" style="margin: 0">最近教练会话</h2>
        <router-link class="btn btn-ghost btn-sm" to="/wellbeing">去打卡</router-link>
      </div>
      <p v-if="sessionsLoading" class="muted">加载中…</p>
      <p v-else-if="!sessions.length" class="muted">还没有会话。从上方场景开始第一次教练。</p>
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

    <section class="panel" style="margin-top: 16px">
      <div class="workspace-toolbar">
        <h2 class="section-title" style="margin: 0">过关训练</h2>
        <router-link class="btn btn-primary btn-sm" to="/tasks/new">新建任务</router-link>
      </div>
      <p class="muted">简历人事关 · 业务模拟 · 谈薪对照——需要练表达时再进。</p>
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
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, SCENE_LABEL, STATUS_LABEL } from '../api'

const router = useRouter()
const error = ref('')
const starting = ref(false)
const sessions = ref([])
const sessionsLoading = ref(true)
const tasks = ref([])
const tasksLoading = ref(true)
const summary = ref(null)
const summaryLoading = ref(true)

const scenes = [
  { value: 'job_search', label: '求职 / 跳槽', desc: '挂面、投递耗竭、在职偷投焦虑' },
  { value: 'promotion', label: '晋升 / 述职', desc: '述职压力、与上级预期错位' },
  { value: 'communication', label: '沟通 / 冲突', desc: '会后内耗、边界与诉求' },
]

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
    error.value = e.message
  } finally {
    starting.value = false
  }
}

onMounted(async () => {
  try {
    const [sessData, taskData, checkData] = await Promise.all([
      api.listCoachSessions(),
      api.listTasks(),
      api.listCheckIns(),
    ])
    sessions.value = sessData.items || []
    tasks.value = taskData.items || []
    summary.value = checkData.summary || null
  } catch (e) {
    error.value = e.message
  } finally {
    sessionsLoading.value = false
    tasksLoading.value = false
    summaryLoading.value = false
  }
})
</script>
