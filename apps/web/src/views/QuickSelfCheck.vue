<template>
  <section>
    <div class="hero">
      <h1>{{ goCoachNext ? '先花三分钟看看自己' : '花三分钟看看自己' }}</h1>
      <p>
        <template v-if="goCoachNext">先用三分钟看看自己，再开始，我会更接得住你。做完后若还想被更深接住，也可以再多说几句。</template>
        <template v-else>没有标准答案，勾最贴近此刻的即可。这会轻轻记在你的状态里。不是临床测评。</template>
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <template v-if="!saved">
      <form class="panel form quick-form" @submit.prevent="submit">
        <fieldset class="tag-fieldset">
          <legend>一、你今天主要是哪种感觉？（不超过 2 项）</legend>
          <label v-for="o in FEELING_OPTIONS" :key="o.value" class="option-row">
            <input
              v-model="form.feelings"
              type="checkbox"
              :value="o.value"
              :disabled="form.feelings.length >= 2 && !form.feelings.includes(o.value)"
            />
            <span><strong>{{ o.label }}</strong> —— {{ o.hint }}</span>
          </label>
        </fieldset>

        <fieldset class="tag-fieldset">
          <legend>二、这种感觉有多久了？</legend>
          <label v-for="o in DURATION_OPTIONS" :key="o.value" class="option-row">
            <input v-model="form.duration" type="radio" :value="o.value" required />
            <span>{{ o.label }}</span>
          </label>
        </fieldset>

        <fieldset class="tag-fieldset">
          <legend>三、它影响了你生活的哪部分？（可多选）</legend>
          <label v-for="o in IMPACT_OPTIONS" :key="o.value" class="option-row">
            <input v-model="form.impacts" type="checkbox" :value="o.value" @change="onImpactChange(o.value)" />
            <span>
              <template v-if="o.hint"><strong>{{ o.label }}</strong> —— {{ o.hint }}</template>
              <template v-else>{{ o.label }}</template>
            </span>
          </label>
        </fieldset>

        <label>
          四、如果给现在的状态打分（0 = 完全没事，10 = 难受到影响日常生活）
          <div class="score-row">
            <input v-model.number="form.distressScore" type="range" min="0" max="10" step="1" />
            <strong class="stat-num">{{ form.distressScore }}</strong>
          </div>
        </label>

        <label>
          五、最近有没有具体的事，让你状态明显往下走的？（选填）
          <textarea v-model="form.triggerNote" rows="2" placeholder="写一两句就行，不想写可以空着" />
        </label>

        <fieldset class="tag-fieldset">
          <legend>六、今天走出这个门的时候，你最想带走什么？</legend>
          <label v-for="o in TAKEAWAY_OPTIONS" :key="o.value" class="option-row">
            <input v-model="form.takeaway" type="radio" :value="o.value" required />
            <span><strong>{{ o.label }}</strong> —— {{ o.hint }}</span>
          </label>
        </fieldset>

        <div class="row" style="flex-wrap: wrap; gap: 10px">
          <button class="btn btn-primary" type="submit" :disabled="busy">
            {{ busy ? '保存中…' : '好，记下这一刻' }}
          </button>
          <router-link class="btn btn-ghost" to="/assessments">我的评估</router-link>
        </div>
      </form>
    </template>

    <template v-else>
      <section class="panel">
        <h2 class="section-title">已记下这一刻</h2>
        <p class="muted">下面是你勾选的内容复述（不做诊断）。</p>
        <ul class="plain-list" style="margin-top: 12px">
          <li>感觉：{{ feelingText }}</li>
          <li>持续：{{ durationText }}</li>
          <li>影响：{{ impactText }}</li>
          <li>困扰分：{{ saved.distressScore }}/10</li>
          <li v-if="saved.triggerNote">触发：{{ saved.triggerNote }}</li>
          <li>想带走：{{ takeawayText }}</li>
        </ul>
        <p v-if="saved.distressScore >= 8" class="muted" style="margin-top: 12px">
          困扰分偏高时，步子可以放慢；若持续很难受，请考虑寻求专业支持。本产品是职场教练，不能替代诊疗。
        </p>

        <template v-if="!hasBaseline">
          <h3 class="pane-sub" style="margin-top: 18px">想被更深接住一点吗？</h3>
          <p class="muted">刚才看的是此刻。若愿意，再花几分钟做一份心理评估，说说近况和期望，之后会更贴你。</p>
          <div class="row" style="margin-top: 12px; flex-wrap: wrap; gap: 10px">
            <router-link
              class="btn btn-primary"
              :to="{ path: '/onboarding/assessment', query: { from: 'quick', quickId: saved.id } }"
            >
              好，去做心理评估
            </router-link>
            <button
              v-if="goCoachNext && pendingScene"
              class="btn btn-ghost"
              type="button"
              :disabled="starting"
              @click="startCoach(pendingScene)"
            >
              {{ starting ? '正在开门…' : '先去聊也行' }}
            </button>
          </div>
        </template>

        <template v-else>
          <h3 class="pane-sub" style="margin-top: 18px">接着进入教练？</h3>
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
          <p class="muted" style="margin-top: 14px">
            若近况变了，也可以
            <router-link
              :to="{ path: '/onboarding/assessment', query: { from: 'quick', quickId: saved.id } }"
            >
              重新评估一下
            </router-link>
          </p>
        </template>

        <div class="row" style="margin-top: 14px; flex-wrap: wrap; gap: 10px">
          <router-link class="btn btn-ghost" to="/wellbeing">看趋势</router-link>
          <router-link class="btn btn-ghost" to="/home">先回到安静的一页</router-link>
        </div>
      </section>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  api,
  DURATION_OPTIONS,
  FEELING_OPTIONS,
  IMPACT_OPTIONS,
  TAKEAWAY_OPTIONS,
} from '../api'

const router = useRouter()
const route = useRoute()
const busy = ref(false)
const starting = ref(false)
const error = ref('')
const saved = ref(null)
const hasBaseline = ref(true)

const pendingScene = computed(() =>
  typeof route.query.scene === 'string' ? route.query.scene : '',
)
const goCoachNext = computed(() => route.query.next === 'coach')

const form = reactive({
  feelings: [],
  duration: '',
  impacts: [],
  distressScore: 5,
  triggerNote: '',
  takeaway: '',
})

const scenes = [
  { value: 'job_search', label: '求职这件事', desc: '投递、挂面、不敢告诉同事时' },
  { value: 'promotion', label: '晋升与述职', desc: '怕讲浅、怕被看轻时' },
  { value: 'communication', label: '沟通与冲突', desc: '会后还在心里吵时' },
]

function labelOf(options, value) {
  return options.find((o) => o.value === value)?.label || value
}

const feelingText = computed(() =>
  (saved.value?.feelings || []).map((v) => labelOf(FEELING_OPTIONS, v)).join('、') || '—',
)
const durationText = computed(() => labelOf(DURATION_OPTIONS, saved.value?.duration) || '—')
const impactText = computed(() =>
  (saved.value?.impacts || []).map((v) => labelOf(IMPACT_OPTIONS, v)).join('、') || '—',
)
const takeawayText = computed(() => labelOf(TAKEAWAY_OPTIONS, saved.value?.takeaway) || '—')

function onImpactChange(value) {
  if (value === 'mood_only' && form.impacts.includes('mood_only')) {
    form.impacts = ['mood_only']
    return
  }
  if (value !== 'mood_only' && form.impacts.includes('mood_only')) {
    form.impacts = form.impacts.filter((v) => v !== 'mood_only')
  }
}

async function submit() {
  error.value = ''
  if (!form.feelings.length) {
    error.value = '请至少勾选一种感觉（最多 2 项）'
    return
  }
  if (!form.impacts.length) {
    error.value = '请至少勾选一项影响面'
    return
  }
  busy.value = true
  try {
    saved.value = await api.createQuickSelfCheck({
      feelings: form.feelings,
      duration: form.duration,
      impacts: form.impacts,
      distressScore: form.distressScore,
      triggerNote: form.triggerNote,
      takeaway: form.takeaway,
    })
    // 有基线且指定了场景 → 直达会话；无基线则留在结果页引导续做基线
    if (goCoachNext.value && pendingScene.value && hasBaseline.value) {
      await startCoach(pendingScene.value)
    }
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

async function startCoach(scene) {
  if (!saved.value?.id) return
  starting.value = true
  error.value = ''
  try {
    const mode = typeof route.query.mode === 'string' ? route.query.mode : 'formal'
    const sess = await api.createCoachSession({
      scene,
      relatedQuickCheckId: saved.value.id,
      mode,
    })
    router.push(`/coach/${sess.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    starting.value = false
  }
}

onMounted(async () => {
  try {
    const me = await api.me()
    hasBaseline.value = Boolean(me.hasInitialAssessment)
  } catch {
    hasBaseline.value = true
  }
})
</script>
