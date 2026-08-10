<template>
  <section>
    <div class="hero">
      <h1>确认具体诉求</h1>
      <p>结合评测结果选一个主路径；之后可在工作台随时调整重心。</p>
    </div>

    <p v-if="hint" class="muted">评测建议：{{ NEED_LABEL[hint] || hint }}</p>
    <p v-if="error" class="error">{{ error }}</p>

    <div class="scene-grid" style="margin-top: 14px">
      <button
        v-for="o in NEED_OPTIONS"
        :key="o.value"
        class="scene-card"
        type="button"
        :class="{ selected: selected === o.value }"
        :disabled="busy"
        @click="selected = o.value"
      >
        <strong>{{ o.label }}</strong>
        <span>{{ o.desc }}</span>
      </button>
    </div>

    <div class="row" style="margin-top: 16px; flex-wrap: wrap; gap: 10px">
      <button class="btn btn-primary" type="button" :disabled="busy || !selected" @click="confirm">
        {{ busy ? '保存中…' : '进入教练工作台' }}
      </button>
      <router-link class="btn btn-ghost" to="/assessments">先回看评估</router-link>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, NEED_LABEL, NEED_OPTIONS } from '../api'
import { clearProfileCache } from '../router'

const router = useRouter()
const selected = ref('')
const hint = ref('')
const busy = ref(false)
const error = ref('')

async function confirm() {
  if (!selected.value) return
  busy.value = true
  error.value = ''
  try {
    await api.setPrimaryNeed(selected.value)
    clearProfileCache()
    router.replace('/home')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  try {
    const me = await api.me()
    hint.value = me.suggestedNeed || ''
    selected.value = me.primaryNeed || me.suggestedNeed || 'counsel_first'
  } catch (e) {
    error.value = e.message
  }
})
</script>
