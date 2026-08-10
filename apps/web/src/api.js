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
    const err = new Error(data.error || `请求失败 (${res.status})`)
    err.status = res.status
    err.code = data.code
    err.redirect = data.redirect
    err.payload = data
    throw err
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
  setPrimaryNeed: (need) =>
    request('/api/needs/', { method: 'POST', body: JSON.stringify({ need }) }),
  listAssessments: () => request('/api/assessments/'),
  createAssessment: (body) =>
    request('/api/assessments/', { method: 'POST', body: JSON.stringify(body) }),
  latestAssessment: () => request('/api/assessments/latest'),
  getAssessment: (id) => request(`/api/assessments/${id}`),
  listBookings: () => request('/api/counseling/bookings/'),
  createBooking: (body) =>
    request('/api/counseling/bookings/', { method: 'POST', body: JSON.stringify(body) }),
  patchBooking: (id, body) =>
    request(`/api/counseling/bookings/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
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
  listQuickSelfChecks: () => request('/api/wellbeing/quick-checks/'),
  getQuickSelfCheck: (id) => request(`/api/wellbeing/quick-checks/${id}`),
  createQuickSelfCheck: (body) =>
    request('/api/wellbeing/quick-checks/', { method: 'POST', body: JSON.stringify(body) }),
  deleteQuickSelfCheck: (id) => request(`/api/wellbeing/quick-checks/${id}`, { method: 'DELETE' }),
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

export const FEELING_OPTIONS = [
  { value: 'tired', label: '累', hint: '身体累、脑子累、说不出为什么累' },
  { value: 'irritable', label: '烦', hint: '坐不住、容易炸、看见什么都想关掉' },
  { value: 'numb', label: '空', hint: '没什么感觉，也对什么都没什么感觉' },
  { value: 'afraid', label: '怕', hint: '担心做不好、担心被否定、担心未来' },
  { value: 'stuck', label: '堵', hint: '有话说不出、有火发不出、有泪流不出' },
  { value: 'indifferent', label: '无所谓', hint: '就这样吧，好不好都行' },
]

export const DURATION_OPTIONS = [
  { value: 'few_days', label: '就这两三天' },
  { value: 'one_two_weeks', label: '一两周了' },
  { value: 'over_month', label: '一个月以上' },
  { value: 'unclear_chronic', label: '说不清，好像一直这样' },
]

export const IMPACT_OPTIONS = [
  { value: 'sleep', label: '睡眠', hint: '睡不着、早醒、睡不实' },
  { value: 'appetite', label: '胃口', hint: '不想吃、吃得没味、暴吃' },
  { value: 'focus', label: '集中力', hint: '盯不住屏幕、看什么都飘' },
  { value: 'temper', label: '脾气', hint: '对家人 / 同事容易不耐烦' },
  { value: 'body', label: '身体', hint: '头疼、胸闷、胃不舒服、肩膀紧' },
  { value: 'mood_only', label: '都没影响，就是心里不爽', hint: '' },
]

export const TAKEAWAY_OPTIONS = [
  { value: 'clarity', label: '想通一件事', hint: '脑子里有个结想解开' },
  { value: 'strength', label: '找回一点力量', hint: '最近被消耗得太厉害了' },
  { value: 'tiny_tool', label: '有一个能用的小办法', hint: '明天就能用上的那种' },
  { value: 'just_talk', label: '只是想找个人说说话', hint: '说出来就舒服了' },
  { value: 'unsure_but_here', label: '我也不知道想带走什么', hint: '但我来了' },
]

export const NEED_OPTIONS = [
  { value: 'job_search', label: '找工作 / 跳槽', desc: '简历、面试、谈薪与求职压力' },
  { value: 'promotion', label: '晋升 / 述职', desc: '述职压力、与上级预期对齐' },
  { value: 'communication', label: '沟通 / 冲突', desc: '边界、向上表达、会后内耗' },
  { value: 'counsel_first', label: '先疏导稳住', desc: '暂不进过关，先把状态理清' },
  { value: 'unsure', label: '暂不确定', desc: '先看看评估建议再决定' },
]

export const NEED_LABEL = Object.fromEntries(NEED_OPTIONS.map((o) => [o.value, o.label]))

export const BOOKING_STATUS_LABEL = {
  requested: '待确认',
  confirmed: '已确认',
  done: '已完成',
  cancelled: '已取消',
}

export const CRISIS_HELP = `如果你正在经历强烈的自我伤害念头，或担心可能伤害他人，请立刻寻求专业或紧急帮助（如 120 / 当地心理援助热线），并联系身边可信的人。本产品是职场心理教练，不能替代持证心理咨询或精神科诊疗。`

export function postLoginPath(me) {
  if (!me?.hasInitialAssessment) return '/onboarding/assessment'
  if (!me?.primaryNeed) return '/onboarding/need'
  return '/home'
}
