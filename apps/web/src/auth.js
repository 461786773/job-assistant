import { reactive } from 'vue'

const TOKEN_KEY = 'ja_token'
const USER_KEY = 'ja_user'

// 用 sessionStorage：关闭浏览器后再开会回到登录页
const storage = sessionStorage

const state = reactive({
  token: storage.getItem(TOKEN_KEY) || '',
  user: readUser(),
})

function readUser() {
  try {
    const raw = storage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export function getToken() {
  return state.token
}

export function getUser() {
  return state.user
}

export function setSession(token, user) {
  state.token = token || ''
  state.user = user || null
  storage.setItem(TOKEN_KEY, state.token)
  storage.setItem(USER_KEY, JSON.stringify(state.user))
  // 清掉旧的 localStorage，避免误用历史登录态
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function clearSession() {
  state.token = ''
  state.user = null
  storage.removeItem(TOKEN_KEY)
  storage.removeItem(USER_KEY)
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isLoggedIn() {
  return Boolean(state.token)
}

export function useAuthState() {
  return state
}
