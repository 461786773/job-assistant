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
    if (!window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/register')) {
      const redirect = encodeURIComponent(window.location.pathname + window.location.search)
      window.location.assign(`/login?redirect=${redirect}`)
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
}

export const STATUS_LABEL = {
  draft: '草稿',
  hr_done: '简历优化完成',
  interview_done: '面试模拟完成',
  salary_done: '薪资确认完成',
}
