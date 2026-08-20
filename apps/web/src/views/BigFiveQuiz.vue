<template>
  <section>
    <div class="hero">
      <h1>{{ isWelcome ? '先认识一下现在的你' : '测测你的职场画像' }}</h1>
      <p>
        <template v-if="isWelcome">
          进来坐坐。想先花两三分钟画一张职场小像也可以——之后聊天会参考它；不想测的话，随时可以去别处看看。
        </template>
        <template v-else>
          约 15 题、三分钟左右。结果会进入教练对话，用来调接话风格；不是诊断，也不影响录用。
        </template>
      </p>
    </div>

    <div class="panel" style="margin-bottom: 12px">
      <div class="muted">进度 {{ answered }}/{{ QUESTIONS.length }}</div>
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: `${(answered / QUESTIONS.length) * 100}%` }" />
      </div>
    </div>

    <form class="panel form quick-form" @submit.prevent="submit">
      <fieldset
        v-for="q in QUESTIONS"
        :id="`bf-${q.key}`"
        :key="q.key"
        class="tag-fieldset"
        :class="{ 'fieldset-miss': missKey === q.key }"
      >
        <legend>{{ q.n }}. {{ q.text }}</legend>
        <div class="likert-row">
          <label v-for="n in 5" :key="n" class="likert-opt">
            <input v-model.number="form[q.key]" type="radio" :name="q.key" :value="n" />
            <span>{{ LIKERT[n] }}</span>
          </label>
        </div>
      </fieldset>

      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
      <p v-else-if="answered < QUESTIONS.length" class="muted" style="margin-top: 8px">
        还差 {{ QUESTIONS.length - answered }} 题，勾完再生成画像。
      </p>

      <div class="row" style="flex-wrap: wrap; gap: 10px; margin-top: 12px">
        <button class="btn btn-primary" type="button" :disabled="busy" @click="submit">
          {{ busy ? '正在画你的像…' : '看看我的画像' }}
        </button>
        <router-link class="btn btn-ghost" to="/home">
          {{ isWelcome ? '我想先四处看看' : '先去别处转转' }}
        </router-link>
      </div>
    </form>
  </section>
</template>

<script setup>
import { computed, nextTick, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { clearProfileCache } from '../router'

const router = useRouter()
const route = useRoute()
const busy = ref(false)
const error = ref('')
const missKey = ref('')

const isWelcome = computed(() => route.query.welcome === '1')

const LIKERT = {
  1: '非常不符合',
  2: '比较不符合',
  3: '一般',
  4: '比较符合',
  5: '非常符合',
}

const QUESTIONS = [
  { n: 1, key: 'q1', text: '面对新业务或新技术，我常会主动去摸一摸「还有没有更好的做法」。' },
  { n: 2, key: 'q2', text: '聊工作方案时，我喜欢把不同领域的经验扯到一起，看看有没有新组合。' },
  { n: 3, key: 'q3', text: '只要现有流程还能用，我更想少改结构、少碰不确定的方案。' },
  { n: 4, key: 'q4', text: '重要交付前，我会把里程碑和风险点提前排清楚，而不是临场抱佛脚。' },
  { n: 5, key: 'q5', text: '答应别人的事，我通常会跟到底，很少不了了之。' },
  { n: 6, key: 'q6', text: '计划经常变，我更习惯走到哪想到哪，细则以后再说。' },
  { n: 7, key: 'q7', text: '跨部门对齐或客户会，我往往愿意先开口带动气氛、把议题推起来。' },
  { n: 8, key: 'q8', text: '头脑风暴或团建里，和一群人一起吵方案会让我更有劲。' },
  { n: 9, key: 'q9', text: '连续社交或开会后，我更需要独自待一会儿才能缓过来。' },
  { n: 10, key: 'q10', text: '意见不合时，我更想先找都能接受的点，而不是先争对错。' },
  { n: 11, key: 'q11', text: '同事卡住时，我愿意先伸手帮一把，哪怕会占一点自己的时间。' },
  { n: 12, key: 'q12', text: '为了把事情做成，我可以比较直接地指出问题，哪怕场面一时尴尬。' },
  { n: 13, key: 'q13', text: '被挑战或挂面后，我脑子里容易反复回放，短时间难抽离。' },
  { n: 14, key: 'q14', text: '临近述职、谈薪、答辩或关键评审，我的紧张感会明显抬头。' },
  { n: 15, key: 'q15', text: '压力来了，我大多能较快稳住情绪，继续把下一步做掉。' },
]

const form = reactive({
  q1: 0, q2: 0, q3: 0, q4: 0, q5: 0,
  q6: 0, q7: 0, q8: 0, q9: 0, q10: 0,
  q11: 0, q12: 0, q13: 0, q14: 0, q15: 0,
})

const answered = computed(() => QUESTIONS.filter((q) => form[q.key] >= 1 && form[q.key] <= 5).length)

function firstMissing() {
  return QUESTIONS.find((q) => !(form[q.key] >= 1 && form[q.key] <= 5))
}

async function submit() {
  if (busy.value) return
  error.value = ''
  missKey.value = ''
  const miss = firstMissing()
  if (miss) {
    error.value = `还有题目没选，先看第 ${miss.n} 题`
    missKey.value = miss.key
    await nextTick()
    document.getElementById(`bf-${miss.key}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    return
  }
  busy.value = true
  try {
    const answers = Object.fromEntries(QUESTIONS.map((q) => [q.key, form[q.key]]))
    const rec = await api.createBigFive({ answers })
    clearProfileCache()
    await router.replace(`/bigfive/${rec.id}`)
  } catch (e) {
    error.value = e.message || '生成失败，请稍后再试'
  } finally {
    busy.value = false
  }
}
</script>
