<template>
  <section>
    <div class="hero">
      <h1>想被更深地接住时</h1>
      <p>
        {{
          fromQuick
            ? '刚才看的是此刻。下面是心理评估：再花几分钟说说近况和期望；不是诊断，答案只有你自己看得见。'
            : '这是心理评估，大约几分钟的心里盘点。不是诊断，答案只有你自己看得见。轻松聊聊时可以不做。'
        }}
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div v-if="quickStrip" class="panel" style="margin-bottom: 12px">
      <div class="muted">刚才你留下的此刻</div>
      <p style="margin: 6px 0 0">{{ quickStrip }}</p>
    </div>

    <div class="panel" style="margin-bottom: 12px">
      <div class="muted">进度 {{ step }}/{{ totalSteps }}</div>
      <div class="progress-bar"><div class="progress-fill" :style="{ width: `${(step / totalSteps) * 100}%` }" /></div>
    </div>

    <form class="panel form quick-form" @submit.prevent="onNext">
      <!-- Step 0 consent + A -->
      <template v-if="step === 1">
        <fieldset class="tag-fieldset">
          <legend>知情同意</legend>
          <label class="option-row">
            <input v-model="form.consent" type="checkbox" required />
            <span v-if="fromQuick">
              我已知晓：本卷为职场自我评估（非临床诊断），用于教练个性化；与刚才留下的此刻一并使用，不作录用/晋升评判。
            </span>
            <span v-else>
              我已知晓：本问卷为职场自我评估，非临床诊断；用于教练个性化与趋势对照，不作录用/晋升评判。
            </span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>A1. 当前最想优先处理的场景</legend>
          <label v-for="o in SCENE_OPTS" :key="o.value" class="option-row">
            <input v-model="form.primaryScene" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>A2. 目前工作状态</legend>
          <label v-for="o in WORK_OPTS" :key="o.value" class="option-row">
            <input v-model="form.workStatus" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>A3. 近 4 周影响最大的事件（最多 2）</legend>
          <label v-for="o in EVENT_OPTS" :key="o.value" class="option-row">
            <input
              v-model="form.keyEvents"
              type="checkbox"
              :value="o.value"
              :disabled="form.keyEvents.length >= 2 && !form.keyEvents.includes(o.value)"
            />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>A4. 工作年限大致区间</legend>
          <label v-for="o in TENURE_OPTS" :key="o.value" class="option-row">
            <input v-model="form.tenureBand" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
      </template>

      <template v-else-if="step === 2">
        <p class="muted">近 2 周状态（1 = 几乎没有，5 = 非常频繁/非常同意；B5 越高越充沛）</p>
        <label v-for="q in BASELINE_QS" :key="q.key">
          {{ q.label }}
          <div class="score-row">
            <input v-model.number="form[q.key]" type="range" min="1" max="5" step="1" />
            <strong>{{ form[q.key] }}</strong>
          </div>
        </label>
        <fieldset class="tag-fieldset">
          <legend>B7. 最常出现的情绪（最多 3）</legend>
          <label v-for="m in MOOD_ASSESS" :key="m" class="tag-check">
            <input
              v-model="form.moodTags"
              type="checkbox"
              :value="m"
              :disabled="form.moodTags.length >= 3 && !form.moodTags.includes(m)"
            />
            {{ m }}
          </label>
        </fieldset>
      </template>

      <template v-else-if="step === 3">
        <fieldset class="tag-fieldset">
          <legend>C1. 哪些正在明显消耗你？（可多选）</legend>
          <label v-for="o in STRESSOR_OPTS" :key="o" class="option-row">
            <input v-model="form.stressors" type="checkbox" :value="o" />
            <span>{{ o }}</span>
          </label>
        </fieldset>
        <label>
          C2. 对你当前最痛的一项，强度是？
          <div class="score-row">
            <input v-model.number="form.c2" type="range" min="1" max="5" step="1" />
            <strong>{{ form.c2 }}</strong>
          </div>
        </label>
        <template v-if="form.primaryScene === 'job_search'">
          <label>
            C3a. 挂面后我会较长时间陷入自我攻击
            <div class="score-row">
              <input v-model.number="form.c3a" type="range" min="1" max="5" />
              <strong>{{ form.c3a }}</strong>
            </div>
          </label>
          <label>
            C3b. 我能把「发挥问题」和「我这个人不行」分开看（越高越好）
            <div class="score-row">
              <input v-model.number="form.c3b" type="range" min="1" max="5" />
              <strong>{{ form.c3b }}</strong>
            </div>
          </label>
        </template>
        <template v-else-if="form.primaryScene === 'promotion'">
          <label>
            C4a. 述职前我会反复 rumination、难以专注准备
            <div class="score-row">
              <input v-model.number="form.c4a" type="range" min="1" max="5" />
              <strong>{{ form.c4a }}</strong>
            </div>
          </label>
          <label>
            C4b. 我清楚上级/职级标准（越高越好）
            <div class="score-row">
              <input v-model.number="form.c4b" type="range" min="1" max="5" />
              <strong>{{ form.c4b }}</strong>
            </div>
          </label>
        </template>
        <template v-else-if="form.primaryScene === 'communication'">
          <label>
            C5a. 冲突后情绪会劫持我很久
            <div class="score-row">
              <input v-model.number="form.c5a" type="range" min="1" max="5" />
              <strong>{{ form.c5a }}</strong>
            </div>
          </label>
          <label>
            C5b. 我能在冷却后表达边界与诉求（越高越好）
            <div class="score-row">
              <input v-model.number="form.c5b" type="range" min="1" max="5" />
              <strong>{{ form.c5b }}</strong>
            </div>
          </label>
        </template>
      </template>

      <template v-else-if="step === 4">
        <fieldset class="tag-fieldset">
          <legend>D1. 压力大时通常会怎么做？（可多选）</legend>
          <label v-for="o in COPING_OPTS" :key="o" class="option-row">
            <input v-model="form.coping" type="checkbox" :value="o" />
            <span>{{ o }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>D2. 可依靠的支持够不够？</legend>
          <label v-for="o in SUPPORT_OPTS" :key="o.value" class="option-row">
            <input v-model="form.supportLevel" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>D3. 是否愿意做轻量状态跟踪？</legend>
          <label v-for="o in CHECKIN_OPTS" :key="o.value" class="option-row">
            <input v-model="form.checkinWillingness" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
        <fieldset class="tag-fieldset">
          <legend>E1. 近 2 周是否有过「活着没有意思」或伤害自己的想法？</legend>
          <label v-for="o in CRISIS_OPTS" :key="o.value" class="option-row">
            <input v-model="form.crisisLevel" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>
      </template>

      <template v-else>
        <fieldset class="tag-fieldset">
          <legend>F1. 希望教练优先帮你什么？（最多 3）</legend>
          <label v-for="o in GOAL_OPTS" :key="o" class="option-row">
            <input
              v-model="form.goals"
              type="checkbox"
              :value="o"
              :disabled="form.goals.length >= 3 && !form.goals.includes(o)"
            />
            <span>{{ o }}</span>
          </label>
        </fieldset>
        <label>
          F2. 此刻最卡你的是什么？（选填）
          <textarea v-model="form.freeTextBlockers" rows="2" maxlength="100" />
        </label>
        <label>
          F3. 还有什么希望教练知道？（选填）
          <textarea v-model="form.freeTextOther" rows="2" maxlength="100" />
        </label>
      </template>

      <div class="row" style="flex-wrap: wrap; gap: 10px">
        <button v-if="step > 1" class="btn btn-ghost" type="button" @click="step -= 1">上一步</button>
        <button class="btn btn-primary" type="submit" :disabled="busy">
          {{ busy ? '我在看…' : step < totalSteps ? '下一步' : '好，看看我的摘要' }}
        </button>
        <button class="btn btn-ghost" type="button" @click="skipForLater">先去别处转转</button>
      </div>
    </form>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, FEELING_OPTIONS } from '../api'
import { clearProfileCache } from '../router'

const router = useRouter()
const route = useRoute()
const step = ref(1)
const totalSteps = 5
const busy = ref(false)
const error = ref('')
const quickStrip = ref('')

const fromQuick = computed(() => route.query.from === 'quick')
const quickId = computed(() => (typeof route.query.quickId === 'string' ? route.query.quickId : ''))

const form = reactive({
  consent: false,
  primaryScene: '',
  workStatus: '',
  keyEvents: [],
  tenureBand: '',
  b1: 3, b2: 3, b3: 3, b4: 3, b5: 3, b6: 3,
  moodTags: [],
  stressors: [],
  c2: 3, c3a: 3, c3b: 3, c4a: 3, c4b: 3, c5a: 3, c5b: 3,
  coping: [],
  supportLevel: '',
  checkinWillingness: '',
  crisisLevel: '',
  goals: [],
  freeTextBlockers: '',
  freeTextOther: '',
})

const SCENE_OPTS = [
  { value: 'job_search', label: '求职这件事' },
  { value: 'promotion', label: '晋升和述职' },
  { value: 'communication', label: '和人较劲 / 说不清边界' },
  { value: 'mixed', label: '好几件事缠在一起' },
  { value: 'other', label: '别的职场心事' },
]
const WORK_OPTS = [
  { value: 'employed_stable', label: '在职，暂无明确离职计划' },
  { value: 'employed_looking', label: '在职，正在看机会 / 面试中' },
  { value: 'unemployed', label: '已离职 / 空窗求职中' },
  { value: 'other', label: '其他' },
]
const EVENT_OPTS = [
  { value: 'dense_apps', label: '投递/约面密集' },
  { value: 'interview_fail', label: '面试发挥失常或挂面' },
  { value: 'offer_talk', label: 'Offer / 谈薪拉扯' },
  { value: 'promo_review', label: '晋升述职或职级反馈' },
  { value: 'boss_mismatch', label: '与上级预期错位' },
  { value: 'conflict', label: '跨部门冲突或被否定' },
  { value: 'workload', label: '工作量/交付高压' },
  { value: 'chronic', label: '无明显事件，但持续紧绷' },
]
const TENURE_OPTS = [
  { value: 'lt3', label: '3 年以下' },
  { value: '3to5', label: '3–5 年' },
  { value: '6to10', label: '6–10 年' },
  { value: 'gt10', label: '10 年以上' },
]
const BASELINE_QS = [
  { key: 'b1', label: 'B1. 整体压力感' },
  { key: 'b2', label: 'B2. 情绪低落或提不起劲' },
  { key: 'b3', label: 'B3. 焦虑、担心停不下来' },
  { key: 'b4', label: 'B4. 因职场事件自我否定' },
  { key: 'b5', label: 'B5. 精力水平' },
  { key: 'b6', label: 'B6. 睡眠受职场思虑影响' },
]
const MOOD_ASSESS = ['平静', '焦虑', '低落', '烦躁', '羞耻 / 尴尬', '愤怒', '麻木', '兴奋 / 紧绷中的亢奋']
const STRESSOR_OPTS = [
  '简历主线说不清 / 与 JD 对不齐',
  '面试讲述表层、怕被追问穿帮',
  '谈薪不懂结构、怕被流水压价',
  '投递久了耗竭、想放弃',
  '在职偷投的焦虑或羞耻',
  '晋升述职不知怎么呈现价值',
  '怕暴露短板、被上级看轻',
  '冲突后反复内耗',
  '该争取时说不出口 / 该止损时停不下来',
  '身份迁移叙事弱（跨岗/跨赛道）',
]
const COPING_OPTS = [
  '硬扛、假装没事', '运动 / 睡觉 / 娱乐分心', '和朋友或家人说', '写下来 / 自己复盘',
  '找同事或导师聊', '专业心理咨询 / EAP', '刷手机到更累', '饮酒或其他麻痹方式', '尚无稳定办法',
]
const SUPPORT_OPTS = [
  { value: 'enough', label: '比较够' },
  { value: 'ok', label: '一般' },
  { value: 'low', label: '明显不够' },
  { value: 'unsure', label: '不确定' },
]
const CHECKIN_OPTS = [
  { value: 'daily', label: '愿意，尽量每天' },
  { value: 'nodes', label: '愿意，关键节点再打' },
  { value: 'maybe', label: '先看看，不确定' },
  { value: 'no', label: '暂时不想' },
]
const CRISIS_OPTS = [
  { value: 'none', label: '完全没有' },
  { value: 'fleeting', label: '偶尔有过一闪而过的念头，但能自己拉开' },
  { value: 'elevated', label: '较频繁，或已有具体计划 / 行动' },
]
const GOAL_OPTS = [
  '稳住情绪、减少内耗', '把事实和自我评判拆开', '挂面 / 冲突后的复盘',
  '晋升或向上沟通怎么开口', '谈薪前的心理准备与话术底气', '建立可持续的状态跟踪',
  '需要时带我去做简历/面试/谈薪训练',
]

function validateStep() {
  if (step.value === 1) {
    if (!form.consent) return '请先勾选知情同意'
    if (!form.primaryScene || !form.workStatus || !form.tenureBand) return '请完成背景题'
  }
  if (step.value === 4) {
    if (!form.supportLevel || !form.checkinWillingness || !form.crisisLevel) return '请完成支持与安全筛查'
  }
  return ''
}

async function onNext() {
  error.value = validateStep()
  if (error.value) return
  if (step.value < totalSteps) {
    step.value += 1
    return
  }
  busy.value = true
  try {
    const rec = await api.createAssessment({ ...form })
    clearProfileCache()
    router.replace(`/onboarding/report/${rec.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

function skipForLater() {
  router.push('/home')
}

onMounted(async () => {
  if (!quickId.value) return
  try {
    const q = await api.getQuickSelfCheck(quickId.value)
    if (!q) return
    const feels = (q.feelings || [])
      .map((v) => FEELING_OPTIONS.find((o) => o.value === v)?.label || v)
      .filter(Boolean)
      .join('、')
    const bits = []
    if (feels) bits.push(`感觉：${feels}`)
    if (q.distressScore) bits.push(`困扰 ${q.distressScore}/10`)
    if (q.triggerNote) bits.push(q.triggerNote)
    quickStrip.value = bits.join(' · ')
  } catch {
    quickStrip.value = ''
  }
})
</script>
