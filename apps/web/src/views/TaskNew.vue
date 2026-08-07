<template>
  <section>
    <div class="hero">
      <h1>新建任务</h1>
      <p>一次投递 = 一个任务。支持 Markdown、纯文本、docx、PDF 简历。</p>
    </div>

    <form class="panel form" @submit.prevent="submit">
      <div class="split">
        <label>
          公司
          <input v-model="form.company" placeholder="例如：吉利" />
        </label>
        <label>
          目标岗位
          <input v-model="form.targetRole" placeholder="例如：数据合规" />
        </label>
      </div>

      <label>
        任务标题（可空，默认公司·岗位）
        <input v-model="form.title" placeholder="留空则自动生成" />
      </label>

      <label>
        上传简历（md / txt / docx / pdf）
        <input type="file" accept=".md,.markdown,.txt,.docx,.pdf" @change="onFile" />
      </label>
      <p v-if="fileHint" class="muted">{{ fileHint }}</p>
      <p v-if="parseError" class="error">{{ parseError }}</p>

      <label>
        简历正文
        <textarea v-model="form.resumeText" placeholder="上传后自动填充，也可直接粘贴 Markdown/纯文本" rows="12" />
      </label>

      <label>
        目标 JD
        <textarea v-model="form.jdText" placeholder="粘贴岗位职责与任职要求" rows="10" />
      </label>

      <label>
        备注
        <textarea v-model="form.notes" placeholder="可选：约面时间、HR 对接人等" rows="3" />
      </label>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="row">
        <button class="btn btn-primary" type="submit" :disabled="saving">
          {{ saving ? '创建中…' : '创建任务' }}
        </button>
        <router-link class="btn btn-ghost" to="/">返回</router-link>
      </div>
    </form>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'

const router = useRouter()
const saving = ref(false)
const error = ref('')
const parseError = ref('')
const fileHint = ref('')

const form = reactive({
  title: '',
  company: '',
  targetRole: '',
  resumeText: '',
  jdText: '',
  notes: '',
})

let pendingFile = null

async function onFile(e) {
  const file = e.target.files?.[0]
  parseError.value = ''
  fileHint.value = ''
  pendingFile = null
  if (!file) return
  try {
    const data = await api.parseResume(file)
    form.resumeText = data.text || ''
    pendingFile = file
    fileHint.value = `已解析 ${data.filename}（${data.format}，约 ${data.chars} 字）`
  } catch (err) {
    parseError.value = err.message
  }
}

async function submit() {
  error.value = ''
  if (!form.resumeText.trim()) {
    error.value = '请上传或粘贴简历'
    return
  }
  if (!form.jdText.trim()) {
    error.value = '请粘贴目标 JD'
    return
  }
  saving.value = true
  try {
    // 简历正文已在解析时写入，创建任务只提交文本，避免二次上传失败导致整单报错
    const task = await api.createTask({ ...form })
    if (pendingFile) {
      try {
        await api.uploadResume(task.id, pendingFile)
      } catch (uploadErr) {
        console.warn('resume re-upload skipped:', uploadErr)
      }
    }
    router.push(`/tasks/${task.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}
</script>
