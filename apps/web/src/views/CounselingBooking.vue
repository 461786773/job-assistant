<template>
  <section>
    <div class="hero">
      <h1>预约私教</h1>
      <p>留下你方便的时间，我们人工确认。这是转介通道，不是站内诊疗。</p>
    </div>

    <section v-if="!hasAssessment" class="panel soft-banner" style="margin-bottom: 14px">
      <div>
        <strong>若你愿意，可先做一次详细评估</strong>
        <p class="muted" style="margin: 6px 0 0">方便对方更懂你；也可以边约边补，不会拦住你。</p>
      </div>
      <router-link class="btn btn-ghost btn-sm" to="/onboarding/assessment">去做</router-link>
    </section>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="coach-home-grid">
      <section class="panel">
        <h2 class="section-title">提交预约意向</h2>
        <form class="form" @submit.prevent="submit">
          <label>
            时段偏好
            <textarea
              v-model="form.preferredSlots"
              rows="2"
              required
              placeholder="例如：本周三晚 20:00 后，或周末上午"
            />
          </label>
          <label>
            联系方式约定
            <input v-model="form.contactChannel" placeholder="微信 / 电话 / 邮件约定方式" required />
          </label>
          <label>
            诉求摘要（选填）
            <textarea v-model="form.note" rows="3" placeholder="希望辅导关注什么节点" />
          </label>
          <button class="btn btn-primary" type="submit" :disabled="busy">
            {{ busy ? '提交中…' : '提交预约' }}
          </button>
        </form>
      </section>

      <section class="panel">
        <h2 class="section-title">我的预约</h2>
        <p v-if="loading" class="muted">加载中…</p>
        <p v-else-if="!items.length" class="muted">还没有预约记录。</p>
        <ul v-else class="checkin-list">
          <li v-for="b in items" :key="b.id">
            <div>
              <strong>{{ BOOKING_STATUS_LABEL[b.status] || b.status }}</strong>
              <span class="muted"> · {{ formatTime(b.updatedAt) }}</span>
              <div class="meta">
                <div>{{ b.preferredSlots }}</div>
                <div v-if="b.contactChannel">联系：{{ b.contactChannel }}</div>
                <div v-if="b.note">{{ b.note }}</div>
              </div>
            </div>
            <button
              v-if="b.status === 'requested' || b.status === 'confirmed'"
              class="btn btn-ghost btn-sm"
              type="button"
              @click="cancel(b.id)"
            >
              取消
            </button>
          </li>
        </ul>
        <p class="muted" style="margin-top: 12px">{{ CRISIS_HELP }}</p>
      </section>
    </div>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { api, BOOKING_STATUS_LABEL, CRISIS_HELP } from '../api'

const items = ref([])
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const hasAssessment = ref(true)
const form = reactive({
  preferredSlots: '',
  contactChannel: '',
  note: '',
})

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

async function refresh() {
  loading.value = true
  try {
    const [data, me] = await Promise.all([api.listBookings(), api.me()])
    items.value = data.items || []
    hasAssessment.value = Boolean(me.hasInitialAssessment)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function submit() {
  busy.value = true
  error.value = ''
  try {
    await api.createBooking({ ...form })
    form.preferredSlots = ''
    form.contactChannel = ''
    form.note = ''
    await refresh()
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

async function cancel(id) {
  if (!confirm('取消这条预约？')) return
  try {
    await api.patchBooking(id, { status: 'cancelled' })
    await refresh()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(refresh)
</script>
