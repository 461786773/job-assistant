const API_BASE = ''

import { clearSession, getToken } from './auth'

function networkHint(err) {
  const msg = err && err.message ? err.message : String(err)
  if (msg === 'Failed to fetch' || msg.includes('NetworkError') || msg.includes('Load failed')) {
    return '无法连接后端 API。请确认 http://127.0.0.1:8080/api/health 可访问，或重新执行 ./scripts/dev-watch.sh'
  }
  return msg
}

async function request(path, options = {}) {
  let res
  const token = getToken()
  const headers = {
    ...(options.body && !(options.body instanceof FormData)
      ? { 'Content-Type': 'application/json' }
      : {}),
    ...options.headers,
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    })
  } catch (err) {
    throw new Error(networkHint(err))
  }
  if (res.status === 204) return null
  const data = await res.json().catch(() => ({}))
  if (res.status === 401) {
    clearSession()
    if (!window.location.pathname.startsWith('/login') && window.location.pathname !== '/') {
      const redirect = encodeURIComponent(window.location.pathname + window.location.search)
      window.location.assign(`/?redirect=${redirect}`)
    }
    throw new Error(data.error || '请先登录')
  }
  if (!res.ok) {
    throw new Error(data.error || `请求失败 (${res.status})`)
  }
  return data
}

export const api = {
  health: () => request('/api/health'),
  register: (body) =>
    request('/api/auth/register', { method: 'POST', body: JSON.stringify(body) }),
  login: (body) =>
    request('/api/auth/login', { method: 'POST', body: JSON.stringify(body) }),
  me: () => request('/api/auth/me'),
  listTasks: () => request('/api/tasks/'),
  getTask: (id) => request(`/api/tasks/${id}`),
  createTask: (body) =>
    request('/api/tasks/', { method: 'POST', body: JSON.stringify(body) }),
  updateTask: (id, body) =>
    request(`/api/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  deleteTask: (id) => request(`/api/tasks/${id}`, { method: 'DELETE' }),
  uploadResume: (id, file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request(`/api/tasks/${id}/resume`, { method: 'POST', body: fd })
  },
  parseResume: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request('/api/resume/parse', { method: 'POST', body: fd })
  },
  analyzeHR: (id) =>
    request(`/api/tasks/${id}/hr/analyze`, { method: 'POST', body: '{}' }),
  applyHRRewrites: (id, body) =>
    request(`/api/tasks/${id}/hr/apply-rewrites`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  interviewStart: (id) =>
    request(`/api/tasks/${id}/interview/start`, { method: 'POST', body: '{}' }),
  interviewReply: (id, answer) =>
    request(`/api/tasks/${id}/interview/reply`, {
      method: 'POST',
      body: JSON.stringify({ answer }),
    }),
  salaryAnalyze: (id, body) =>
    request(`/api/tasks/${id}/salary/analyze`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  listCoachSessions: () => request('/api/coach/sessions/'),
  getCoachSession: (id) => request(`/api/coach/sessions/${id}`),
  createCoachSession: (body) =>
    request('/api/coach/sessions/', { method: 'POST', body: JSON.stringify(body) }),
  replyCoachSession: (id, message) =>
    request(`/api/coach/sessions/${id}/reply`, {
      method: 'POST',
      body: JSON.stringify({ message }),
    }),
  deleteCoachSession: (id) => request(`/api/coach/sessions/${id}`, { method: 'DELETE' }),
  listCheckIns: () => request('/api/wellbeing/checkins/'),
  createCheckIn: (body) =>
    request('/api/wellbeing/checkins/', { method: 'POST', body: JSON.stringify(body) }),
  deleteCheckIn: (id) => request(`/api/wellbeing/checkins/${id}`, { method: 'DELETE' }),
}

export const STATUS_LABEL = {
  draft: '草稿',
  hr_done: '人事关完成',
  interview_done: '业务关完成',
  salary_done: '谈薪关完成',
}

export const SCENE_LABEL = {
  job_search: '求职 / 跳槽',
  promotion: '晋升 / 述职',
  communication: '职场沟通 / 冲突',
}

export const EVENT_OPTIONS = [
  { value: '', label: '无特定事件' },
  { value: 'interview', label: '面试' },
  { value: 'reject', label: '挂面' },
  { value: 'salary_talk', label: '谈薪' },
  { value: 'promotion_review', label: '晋升述职' },
  { value: 'conflict', label: '冲突会' },
  { value: 'other', label: '其他' },
]

export const MOOD_OPTIONS = ['平静', '焦虑', '低落', '烦躁', '兴奋', '耗竭', '羞耻']

export const CRISIS_HELP = `如果你正在经历强烈的自我伤害念头，或担心可能伤害他人，请立刻寻求专业或紧急帮助（如 120 / 当地心理援助热线），并联系身边可信的人。本产品是职场心理教练，不能替代持证心理咨询或精神科诊疗。`
