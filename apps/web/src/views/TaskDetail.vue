<template>
  <section class="task-detail">
    <div v-if="loading" class="muted">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <template v-else-if="task">
      <div class="hero hero-compact">
        <h1>{{ task.title }}</h1>
        <p>
          {{ [task.company, task.targetRole].filter(Boolean).join(' · ') || '未填公司/岗位' }}
          · {{ STATUS_LABEL[task.status] || task.status }}
        </p>
        <div class="row" style="margin-top: 10px">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="coachStarting" @click="openCoach">
            {{ coachStarting ? '创建中…' : '就这个任务找教练聊聊' }}
          </button>
          <router-link class="btn btn-ghost btn-sm" to="/tasks">返回过关列表</router-link>
        </div>
        <p v-if="coachError" class="error">{{ coachError }}</p>
      </div>

      <div class="gates">
        <button type="button" class="gate" :class="{ active: activeGate === 'hr' }" @click="activeGate = 'hr'">
          <h4>① 人事关</h4>
          <p v-if="hrReport">总分 {{ hrReport.totalScore }} · 点击查看</p>
          <p v-else>评分 / 硬伤 / 改写清单</p>
        </button>
        <button type="button" class="gate" :class="{ active: activeGate === 'interview' }" @click="goInterview">
          <h4>② 业务关</h4>
          <p v-if="interview?.status === 'done'">深度分 {{ interview.diagnosis?.depthScore ?? '—' }} · 点击查看</p>
          <p v-else-if="interview?.messages?.length">进行中 · 第 {{ interview.round }}/{{ interview.maxRounds }} 轮</p>
          <p v-else>模拟深挖与反表层化</p>
        </button>
        <button type="button" class="gate" :class="{ active: activeGate === 'salary' }" @click="activeGate = 'salary'">
          <h4>③ 谈薪关</h4>
          <p v-if="salary?.analysis">已生成对照 · 点击查看</p>
          <p v-else>结构对照与话术</p>
        </button>
      </div>

      <!-- 人事关：左报告 / 中改写 / 右简历 -->
      <div v-show="activeGate === 'hr'" class="hr-workspace">
        <div class="workspace-toolbar">
          <div>
            <h2 class="section-title" style="margin: 0">人事关</h2>
            <p class="muted" style="margin: 4px 0 0">左报告 · 中改写 · 右简历详情</p>
          </div>
          <div class="row">
            <button class="btn btn-primary" type="button" :disabled="analyzing" @click="runHR">
              {{ analyzing ? '分析中…' : hrReport ? '重新分析' : '生成简历评分' }}
            </button>
            <button v-if="hrReport" class="btn btn-primary" type="button" @click="goInterview">
              进入业务关 →
            </button>
          </div>
        </div>
        <p v-if="hrError" class="error">{{ hrError }}</p>

        <div class="tri-pane">
          <!-- 左：报告 -->
          <section class="pane">
            <header class="pane-head">
              <h3>优化报告</h3>
            </header>
            <div class="pane-body">
              <template v-if="hrReport">
                <div class="score-hero score-hero-compact">
                  <div class="score-num">{{ hrReport.totalScore }}</div>
                  <div>
                    <div class="score-label">综合分 · {{ sourceLabel }}</div>
                    <p class="muted" style="margin: 6px 0 0">{{ hrReport.summary }}</p>
                  </div>
                </div>
                <div class="dim-grid dim-grid-stack">
                  <div v-for="d in hrReport.dimensions || []" :key="d.key" class="dim-card">
                    <div class="dim-top">
                      <strong>{{ d.label }}</strong>
                      <span>{{ d.score }}</span>
                    </div>
                    <p class="muted">{{ d.comment }}</p>
                  </div>
                </div>
                <h4 class="pane-sub">硬伤 / 风险</h4>
                <ul class="issue-list">
                  <li v-for="(iss, idx) in hrReport.issues || []" :key="idx" :class="'sev-' + iss.severity">
                    <div class="issue-title">
                      <span class="sev-tag">{{ severityLabel(iss.severity) }}</span>
                      {{ iss.title }}
                    </div>
                    <p>{{ iss.detail }}</p>
                  </li>
                </ul>
              </template>
              <p v-else class="muted pane-empty">生成评分后，报告会显示在这里。</p>
            </div>
          </section>

          <!-- 中：改写清单 -->
          <section class="pane">
            <header class="pane-head pane-head-actions">
              <h3>改写清单</h3>
              <div class="row pane-actions" v-if="hrReport">
                <button
                  class="btn btn-ghost btn-sm"
                  type="button"
                  :disabled="!resumeUndoText || applyBusy"
                  @click="undoResumeApply"
                >
                  撤销
                </button>
                <button
                  class="btn btn-ghost btn-sm"
                  type="button"
                  :disabled="applyBusy || !(hrReport.rewrites || []).length"
                  @click="applyAllRewrites('direct')"
                >
                  直接替换
                </button>
                <button
                  class="btn btn-primary btn-sm"
                  type="button"
                  :disabled="applyBusy || !(hrReport.rewrites || []).length"
                  @click="applyAllRewrites('ai')"
                >
                  {{ applyBusy ? '写入中…' : 'AI 一键写入' }}
                </button>
              </div>
            </header>
            <div class="pane-body">
              <template v-if="hrReport">
                <p class="muted tiny-hint">AI 写入会精修对应段落；直接替换按原文→建议替换。</p>
                <p v-if="applyMsg" :class="applyOk ? 'success' : 'error'">{{ applyMsg }}</p>
                <div
                  v-for="(rw, idx) in hrReport.rewrites || []"
                  :key="idx"
                  class="rewrite-card"
                  :class="{ applied: appliedIndexes.has(idx) }"
                >
                  <div class="rewrite-head">
                    <strong>{{ rw.target }}</strong>
                    <div class="row">
                      <span class="badge">{{ actionLabel(rw.action) }}</span>
                      <span v-if="appliedIndexes.has(idx)" class="badge badge-ok">已应用</span>
                    </div>
                  </div>
                  <p class="muted">{{ rw.reason }}</p>
                  <div class="split rewrite-body" v-if="rw.before || rw.after">
                    <div>
                      <div class="tiny">原文</div>
                      <pre>{{ rw.before || '—' }}</pre>
                    </div>
                    <div>
                      <div class="tiny">建议</div>
                      <pre>{{ rw.after || '—' }}</pre>
                    </div>
                  </div>
                  <div class="row" style="margin-top: 8px">
                    <button
                      class="btn btn-primary btn-sm"
                      type="button"
                      :disabled="applyBusy || appliedIndexes.has(idx)"
                      @click="applyOneRewrite(idx, 'ai')"
                    >
                      {{ appliedIndexes.has(idx) ? '已写入' : 'AI 写入' }}
                    </button>
                    <button
                      class="btn btn-ghost btn-sm"
                      type="button"
                      :disabled="applyBusy || !rw.before || !rw.after || appliedIndexes.has(idx)"
                      @click="applyOneRewrite(idx, 'direct')"
                    >
                      直接替换
                    </button>
                  </div>
                </div>
                <p v-if="!(hrReport.rewrites || []).length" class="muted">暂无改写项。</p>
              </template>
              <p v-else class="muted pane-empty">先生成评分，改写建议会出现在此列。</p>
            </div>
          </section>

          <!-- 右：简历详情 -->
          <section class="pane">
            <header class="pane-head pane-head-actions">
              <h3>简历详情</h3>
              <div class="row pane-actions">
                <button class="btn btn-primary btn-sm" type="button" :disabled="saving" @click="save">
                  {{ saving ? '保存中…' : '保存' }}
                </button>
              </div>
            </header>
            <div class="pane-body form pane-form">
              <div class="split pane-meta">
                <label>公司<input v-model="draft.company" /></label>
                <label>目标岗位<input v-model="draft.targetRole" /></label>
              </div>
              <label class="pane-meta">任务标题<input v-model="draft.title" /></label>
              <label class="pane-meta">上传简历<input type="file" accept=".md,.markdown,.txt,.docx,.pdf" @change="onFile" /></label>
              <p v-if="fileHint" class="muted pane-meta">{{ fileHint }}</p>
              <p v-if="fileError" class="error pane-meta">{{ fileError }}</p>
              <label class="grow-field grow-field-jd">目标 JD<textarea v-model="draft.jdText" /></label>
              <label class="grow-field grow-field-resume">简历正文<textarea v-model="draft.resumeText" /></label>
              <label class="pane-meta">备注<textarea v-model="draft.notes" rows="2" /></label>
              <p v-if="saveMsg" :class="['pane-meta', saveOk ? 'success' : 'error']">{{ saveMsg }}</p>
              <div class="row pane-meta">
                <button class="btn btn-danger btn-sm" type="button" :disabled="saving" @click="remove">删除任务</button>
                <router-link class="btn btn-ghost btn-sm" to="/tasks">返回过关列表</router-link>
              </div>
            </div>
          </section>
        </div>
      </div>

      <!-- 业务关 -->
      <div v-show="activeGate === 'interview'" class="panel form" style="margin-bottom: 16px">
        <div class="row" style="justify-content: space-between">
          <div>
            <h2 class="section-title">业务关</h2>
            <p class="muted" style="margin: 0">业务面试官追问；结束后给出反表层化诊断与补强稿。</p>
          </div>
          <button class="btn btn-primary" type="button" :disabled="interviewBusy" @click="startInterview">
            {{ interviewBusy ? '准备中…' : interview?.messages?.length ? '重新开始' : '开始模拟' }}
          </button>
        </div>
        <p v-if="interviewError" class="error">{{ interviewError }}</p>

        <div v-if="interview?.messages?.length" class="chat">
          <div
            v-for="(m, idx) in interview.messages"
            :key="idx"
            class="bubble"
            :class="m.role"
          >
            <div class="tiny">{{ roleLabel(m.role) }}</div>
            <div class="bubble-body">{{ m.content }}</div>
          </div>
        </div>

        <form v-if="interview?.status === 'active'" class="form" style="margin-top: 12px" @submit.prevent="sendAnswer">
          <label>
            你的回答
            <textarea v-model="answer" rows="5" placeholder="用 STAR 讲：情境、任务、你的行动与取舍、结果" />
          </label>
          <div class="row">
            <button class="btn btn-primary" type="submit" :disabled="interviewBusy || !answer.trim()">
              {{ interviewBusy ? '提交中…' : '提交回答' }}
            </button>
            <span class="muted">第 {{ interview.round }}/{{ interview.maxRounds }} 轮</span>
          </div>
        </form>

        <div v-if="interview?.status === 'done'" class="row" style="margin-top: 14px">
          <button class="btn btn-primary" type="button" @click="activeGate = 'salary'">进入谈薪关 →</button>
        </div>
      </div>

      <!-- 谈薪关 -->
      <div v-show="activeGate === 'salary'" class="panel salary-panel" style="margin-bottom: 16px">
        <h2 class="section-title">谈薪关</h2>
        <p class="muted salary-lead">填写当前薪酬与 Offer 结构，生成年包对照、避坑点与话术。</p>

        <div class="salary-cols">
          <div class="salary-col">
            <h3>当前薪酬</h3>
            <label>月薪（基本+岗位等固定）<input type="number" v-model.number="salaryForm.current.monthlyBase" /></label>
            <label>月补贴<input type="number" v-model.number="salaryForm.current.monthlyAllowance" /></label>
            <label>发薪月数（如 13）<input type="number" v-model.number="salaryForm.current.months" /></label>
            <label>年终月数<input type="number" v-model.number="salaryForm.current.yearEndMonths" /></label>
            <label>年终是否保底
              <select v-model="salaryForm.current.yearEndGuaranteed">
                <option :value="true">是</option>
                <option :value="false">否</option>
              </select>
            </label>
            <label>年度奖金（其他）<input type="number" v-model.number="salaryForm.current.bonusYearly" /></label>
            <label>公积金基数<input type="number" v-model.number="salaryForm.current.housingFundBase" /></label>
            <label>公积金比例（如 0.08）<input type="number" step="0.01" v-model.number="salaryForm.current.housingFundRate" /></label>
          </div>
          <div class="salary-col">
            <h3>Offer</h3>
            <label>月薪（基本+岗位）<input type="number" v-model.number="salaryForm.offer.monthlyBase" /></label>
            <label>月补贴<input type="number" v-model.number="salaryForm.offer.monthlyAllowance" /></label>
            <label>发薪月数<input type="number" v-model.number="salaryForm.offer.months" /></label>
            <label>年终月数<input type="number" v-model.number="salaryForm.offer.yearEndMonths" /></label>
            <label>年终是否保底
              <select v-model="salaryForm.offer.yearEndGuaranteed">
                <option :value="true">是</option>
                <option :value="false">否</option>
              </select>
            </label>
            <label>年度奖金<input type="number" v-model.number="salaryForm.offer.bonusYearly" /></label>
            <label>公积金基数<input type="number" v-model.number="salaryForm.offer.housingFundBase" /></label>
            <label>公积金比例<input type="number" step="0.01" v-model.number="salaryForm.offer.housingFundRate" /></label>
          </div>
        </div>

        <div class="salary-targets">
          <label>你的底线年包（元）<input type="number" v-model.number="salaryForm.floorPkg" /></label>
          <label>目标年包（元）<input type="number" v-model.number="salaryForm.targetPkg" /></label>
        </div>

        <p v-if="salaryError" class="error">{{ salaryError }}</p>
        <div class="row salary-actions">
          <button class="btn btn-primary" type="button" :disabled="salaryBusy" @click="runSalary">
            {{ salaryBusy ? '计算中…' : '生成薪资建议' }}
          </button>
        </div>

        <template v-if="salary?.analysis">
          <div class="score-hero salary-result">
            <div>
              <div class="score-label">年包对照</div>
              <p class="muted" style="margin: 6px 0 0">{{ salary.analysis.summary }}</p>
              <p class="muted">
                当前 保守/目标：{{ wan(salary.analysis.currentConservative) }} / {{ wan(salary.analysis.currentTarget) }} 万；
                Offer 保守/目标：{{ wan(salary.analysis.offerConservative) }} / {{ wan(salary.analysis.offerTarget) }} 万
              </p>
            </div>
          </div>
          <h3 class="section-title">风险点</h3>
          <ul class="issue-list">
            <li v-for="(g, i) in salary.analysis.gaps || []" :key="'g'+i">{{ g }}</li>
          </ul>
          <h3 class="section-title">优先谈的点</h3>
          <ul class="issue-list">
            <li v-for="(g, i) in salary.analysis.askPoints || []" :key="'a'+i">{{ g }}</li>
          </ul>
          <h3 class="section-title">话术</h3>
          <div v-for="(s, i) in salary.analysis.scripts || []" :key="'s'+i" class="rewrite-card">
            <pre>{{ s }}</pre>
          </div>
        </template>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, STATUS_LABEL } from '../api'

const props = defineProps({ id: { type: String, required: true } })
const router = useRouter()

const task = ref(null)
const loading = ref(true)
const error = ref('')
const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(false)
const fileHint = ref('')
const fileError = ref('')
const analyzing = ref(false)
const hrError = ref('')
const activeGate = ref('hr')
const applyBusy = ref(false)
const applyMsg = ref('')
const applyOk = ref(false)
const appliedIndexes = ref(new Set())
const resumeUndoText = ref('')

const interviewBusy = ref(false)
const interviewError = ref('')
const answer = ref('')

const salaryBusy = ref(false)
const salaryError = ref('')
const coachStarting = ref(false)
const coachError = ref('')
const salaryForm = reactive({
  current: emptyComp(),
  offer: emptyComp(),
  floorPkg: 0,
  targetPkg: 0,
})

const draft = reactive({
  title: '',
  company: '',
  targetRole: '',
  resumeText: '',
  jdText: '',
  notes: '',
})

function emptyComp() {
  return {
    monthlyBase: 0,
    monthlyAllowance: 0,
    months: 12,
    yearEndMonths: 0,
    yearEndGuaranteed: false,
    bonusYearly: 0,
    housingFundBase: 0,
    housingFundRate: 0.08,
    notes: '',
  }
}

function parseJSONField(raw) {
  if (!raw) return null
  if (typeof raw === 'object') return raw
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

const hrReport = computed(() => parseJSONField(task.value?.hrReport))
const interview = computed(() => parseJSONField(task.value?.interview))
const salary = computed(() => parseJSONField(task.value?.salary))

const sourceLabel = computed(() => {
  if (!hrReport.value) return ''
  return hrReport.value.source === 'llm' ? '模型评分' : '规则初评'
})

function severityLabel(s) {
  return { critical: '硬伤', warn: '风险', info: '提示' }[s] || s
}
function actionLabel(a) {
  return { keep: '保留', compress: '压缩', rewrite: '重写', quantify: '补量化' }[a] || a
}
function roleLabel(r) {
  return { interviewer: '面试官', candidate: '你', coach: '教练点评' }[r] || r
}
function wan(n) {
  return n ? (n / 10000).toFixed(1) : '0.0'
}

function syncDraft(t) {
  draft.title = t.title || ''
  draft.company = t.company || ''
  draft.targetRole = t.targetRole || ''
  draft.resumeText = t.resumeText || ''
  draft.jdText = t.jdText || ''
  draft.notes = t.notes || ''
  const s = parseJSONField(t.salary)
  if (s?.current) Object.assign(salaryForm.current, emptyComp(), s.current)
  if (s?.offer) Object.assign(salaryForm.offer, emptyComp(), s.offer)
  if (s?.floorPkg != null) salaryForm.floorPkg = s.floorPkg
  if (s?.targetPkg != null) salaryForm.targetPkg = s.targetPkg
}

function goInterview() {
  activeGate.value = 'interview'
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    task.value = await api.getTask(props.id)
    syncDraft(task.value)
    if (task.value.status === 'interview_done') activeGate.value = 'interview'
    else if (task.value.status === 'salary_done') activeGate.value = 'salary'
    else if (hrReport.value) activeGate.value = 'hr'
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function onFile(e) {
  const file = e.target.files?.[0]
  fileHint.value = ''
  fileError.value = ''
  if (!file) return
  try {
    const data = await api.uploadResume(props.id, file)
    task.value = data.task
    syncDraft(task.value)
    fileHint.value = `已更新简历（${data.format}，约 ${data.chars} 字）`
  } catch (err) {
    fileError.value = err.message
  }
}

async function save() {
  saving.value = true
  saveMsg.value = ''
  try {
    task.value = await api.updateTask(props.id, { ...draft })
    syncDraft(task.value)
    saveOk.value = true
    saveMsg.value = '已保存'
  } catch (e) {
    saveOk.value = false
    saveMsg.value = e.message
  } finally {
    saving.value = false
  }
}

async function runHR() {
  hrError.value = ''
  analyzing.value = true
  applyMsg.value = ''
  appliedIndexes.value = new Set()
  resumeUndoText.value = ''
  try {
    task.value = await api.updateTask(props.id, { ...draft })
    const data = await api.analyzeHR(props.id)
    task.value = data.task
    syncDraft(task.value)
  } catch (e) {
    hrError.value = e.message
  } finally {
    analyzing.value = false
  }
}

async function applyRewrites(body) {
  applyBusy.value = true
  applyMsg.value = ''
  try {
    await api.updateTask(props.id, { ...draft })
    const data = await api.applyHRRewrites(props.id, body)
    task.value = data.task
    syncDraft(task.value)
    const next = new Set(appliedIndexes.value)
    let fail = 0
    for (const r of data.results || []) {
      if (r.ok) next.add(r.index)
      else fail++
    }
    appliedIndexes.value = next
    if (data.changed && data.previousResumeText != null) {
      resumeUndoText.value = data.previousResumeText
    }
    applyOk.value = data.changed
    const via = data.mode === 'ai' ? 'AI 精修' : '直接替换'
    if (data.changed && fail === 0) {
      applyMsg.value = `${via}完成：已更新下方「简历正文」`
    } else if (data.changed) {
      applyMsg.value = `${via}部分成功（${data.applied} 条），${fail} 条未落地`
    } else {
      applyMsg.value = data.mode === 'ai'
        ? 'AI 未能安全改写对应段落（可能缺事实依据），请改建议或手动编辑'
        : '未能匹配到原文，可改用「AI 写入」或检查原文摘录'
    }
  } catch (e) {
    applyOk.value = false
    applyMsg.value = e.message
  } finally {
    applyBusy.value = false
  }
}

function applyOneRewrite(idx, mode = 'ai') {
  return applyRewrites({ indexes: [idx], all: false, mode })
}

function applyAllRewrites(mode = 'ai') {
  return applyRewrites({ all: true, mode })
}

async function undoResumeApply() {
  if (!resumeUndoText.value) return
  applyBusy.value = true
  applyMsg.value = ''
  try {
    task.value = await api.updateTask(props.id, { resumeText: resumeUndoText.value })
    syncDraft(task.value)
    resumeUndoText.value = ''
    appliedIndexes.value = new Set()
    applyOk.value = true
    applyMsg.value = '已撤销到应用前的简历正文'
  } catch (e) {
    applyOk.value = false
    applyMsg.value = e.message
  } finally {
    applyBusy.value = false
  }
}

async function startInterview() {
  interviewError.value = ''
  interviewBusy.value = true
  answer.value = ''
  try {
    await api.updateTask(props.id, { ...draft })
    const data = await api.interviewStart(props.id)
    task.value = data.task
  } catch (e) {
    interviewError.value = e.message
  } finally {
    interviewBusy.value = false
  }
}

async function sendAnswer() {
  if (!answer.value.trim()) return
  interviewError.value = ''
  interviewBusy.value = true
  try {
    const data = await api.interviewReply(props.id, answer.value.trim())
    task.value = data.task
    answer.value = ''
  } catch (e) {
    interviewError.value = e.message
  } finally {
    interviewBusy.value = false
  }
}

async function runSalary() {
  salaryError.value = ''
  salaryBusy.value = true
  try {
    const data = await api.salaryAnalyze(props.id, {
      current: { ...salaryForm.current },
      offer: { ...salaryForm.offer },
      floorPkg: salaryForm.floorPkg,
      targetPkg: salaryForm.targetPkg,
    })
    task.value = data.task
  } catch (e) {
    salaryError.value = e.message
  } finally {
    salaryBusy.value = false
  }
}

async function openCoach() {
  coachError.value = ''
  coachStarting.value = true
  try {
    const sess = await api.createCoachSession({
      scene: 'job_search',
      relatedTaskId: props.id,
      relatedEvent: '求职任务复盘',
    })
    router.push(`/coach/${sess.id}`)
  } catch (e) {
    coachError.value = e.message
  } finally {
    coachStarting.value = false
  }
}

async function remove() {
  if (!confirm('确认删除该任务？')) return
  saving.value = true
  try {
    await api.deleteTask(props.id)
    router.push('/tasks')
  } catch (e) {
    saveOk.value = false
    saveMsg.value = e.message
    saving.value = false
  }
}

watch(() => props.id, load)
onMounted(load)
</script>
