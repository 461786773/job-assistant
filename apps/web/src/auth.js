import { reactive } from 'vue'

const TOKEN_KEY = 'ja_token'
const USER_KEY = 'ja_user'

const state = reactive({
  token: localStorage.getItem(TOKEN_KEY) || '',
  user: readUser(),
})

function readUser() {
  try {
    const raw = localStorage.getItem(USER_KEY)
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
  localStorage.setItem(TOKEN_KEY, state.token)
  localStorage.setItem(USER_KEY, JSON.stringify(state.user))
}

export function clearSession() {
  state.token = ''
  state.user = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isLoggedIn() {
  return Boolean(state.token)
}

export function useAuthState() {
  return state
}
