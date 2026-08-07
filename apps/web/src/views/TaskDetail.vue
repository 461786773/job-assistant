<template>
  <section>
    <div v-if="loading" class="muted">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <template v-else-if="task">
      <div class="hero">
        <h1>{{ task.title }}</h1>
        <p>
          {{ [task.company, task.targetRole].filter(Boolean).join(' · ') || '未填公司/岗位' }}
          · {{ STATUS_LABEL[task.status] || task.status }}
        </p>
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

      <!-- 人事关 -->
      <div v-show="activeGate === 'hr'" class="panel form" style="margin-bottom: 16px">
        <div class="row" style="justify-content: space-between; align-items: flex-start">
          <div>
            <h2 class="section-title">人事关报告</h2>
            <p class="muted" style="margin: 0">生成评分后，可进入业务关继续模拟面试。</p>
          </div>
          <div class="row">
            <button class="btn btn-primary" type="button" :disabled="analyzing" @click="runHR">
              {{ analyzing ? '分析中…' : hrReport ? '重新分析' : '生成人事评分' }}
            </button>
            <button
              v-if="hrReport"
              class="btn btn-primary"
              type="button"
              @click="goInterview"
            >
              进入业务关 →
            </button>
          </div>
        </div>
        <p v-if="hrError" class="error">{{ hrError }}</p>

        <template v-if="hrReport">
          <div class="score-hero">
            <div class="score-num">{{ hrReport.totalScore }}</div>
            <div>
              <div class="score-label">人事综合分 · {{ sourceLabel }}</div>
              <p class="muted" style="margin: 6px 0 0">{{ hrReport.summary }}</p>
            </div>
          </div>

          <div class="dim-grid">
            <div v-for="d in hrReport.dimensions || []" :key="d.key" class="dim-card">
              <div class="dim-top">
                <strong>{{ d.label }}</strong>
                <span>{{ d.score }}</span>
              </div>
              <p class="muted">{{ d.comment }}</p>
            </div>
          </div>

          <h3 class="section-title">硬伤 / 风险</h3>
          <ul class="issue-list">
            <li v-for="(iss, idx) in hrReport.issues || []" :key="idx" :class="'sev-' + iss.severity">
              <div class="issue-title">
                <span class="sev-tag">{{ severityLabel(iss.severity) }}</span>
                {{ iss.title }}
              </div>
              <p>{{ iss.detail }}</p>
            </li>
          </ul>

          <h3 class="section-title">改写清单</h3>
          <div v-for="(rw, idx) in hrReport.rewrites || []" :key="idx" class="rewrite-card">
            <div class="rewrite-head">
              <strong>{{ rw.target }}</strong>
              <span class="badge">{{ actionLabel(rw.action) }}</span>
            </div>
            <p class="muted">{{ rw.reason }}</p>
            <div class="split rewrite-body" v-if="rw.before || rw.after">
              <div>
                <div class="tiny">原文 / 现状</div>
                <pre>{{ rw.before || '—' }}</pre>
              </div>
              <div>
                <div class="tiny">建议写法</div>
                <pre>{{ rw.after || '—' }}</pre>
              </div>
            </div>
          </div>

          <div class="row" style="margin-top: 12px">
            <button class="btn btn-primary" type="button" @click="goInterview">进入业务关 →</button>
          </div>
        </template>
      </div>

      <!-- 业务关 -->
      <div v-show="activeGate === 'interview'" class="panel form" style="margin-bottom: 16px">
        <div class="row" style="justify-content: space-between">
          <div>
            <h2 class="section-title">业务关 · 面试模拟</h2>
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
            {{ salaryBusy ? '计算中…' : '生成谈薪建议' }}
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

      <form class="panel form" @submit.prevent="save">
        <h2 class="section-title">任务材料</h2>
        <div class="split">
          <label>公司<input v-model="draft.company" /></label>
          <label>目标岗位<input v-model="draft.targetRole" /></label>
        </div>
        <label>任务标题<input v-model="draft.title" /></label>
        <label>重新上传简历<input type="file" accept=".md,.markdown,.txt,.docx,.pdf" @change="onFile" /></label>
        <p v-if="fileHint" class="muted">{{ fileHint }}</p>
        <p v-if="fileError" class="error">{{ fileError }}</p>
        <label>简历正文<textarea v-model="draft.resumeText" rows="10" /></label>
        <label>目标 JD<textarea v-model="draft.jdText" rows="8" /></label>
        <label>备注<textarea v-model="draft.notes" rows="3" /></label>
        <p v-if="saveMsg" :class="saveOk ? 'success' : 'error'">{{ saveMsg }}</p>
        <div class="row">
          <button class="btn btn-primary" type="submit" :disabled="saving">{{ saving ? '保存中…' : '保存修改' }}</button>
          <button class="btn btn-danger" type="button" :disabled="saving" @click="remove">删除任务</button>
          <router-link class="btn btn-ghost" to="/">返回工作台</router-link>
        </div>
      </form>
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

const interviewBusy = ref(false)
const interviewError = ref('')
const answer = ref('')

const salaryBusy = ref(false)
const salaryError = ref('')
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
  return hrReport.value.source === 'llm' ? '模型人事评分' : '规则初评'
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

async function remove() {
  if (!confirm('确认删除该任务？')) return
  saving.value = true
  try {
    await api.deleteTask(props.id)
    router.push('/')
  } catch (e) {
    saveOk.value = false
    saveMsg.value = e.message
    saving.value = false
  }
}

watch(() => props.id, load)
onMounted(load)
</script>
