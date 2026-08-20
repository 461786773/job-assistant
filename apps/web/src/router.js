import { createRouter, createWebHistory } from 'vue-router'
import CoachHome from './views/CoachHome.vue'
import CoachSession from './views/CoachSession.vue'
import QuickSelfCheck from './views/QuickSelfCheck.vue'
import Settings from './views/Settings.vue'
import Workbench from './views/Workbench.vue'
import TaskNew from './views/TaskNew.vue'
import TaskDetail from './views/TaskDetail.vue'
import Login from './views/Login.vue'
import InitialAssessment from './views/InitialAssessment.vue'
import AssessmentReport from './views/AssessmentReport.vue'
import NeedSelect from './views/NeedSelect.vue'
import MyAssessments from './views/MyAssessments.vue'
import CounselingBooking from './views/CounselingBooking.vue'
import BigFiveQuiz from './views/BigFiveQuiz.vue'
import BigFiveResult from './views/BigFiveResult.vue'
import { api, postLoginPath } from './api'
import { isLoggedIn } from './auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'login', component: Login, meta: { guest: true } },
    { path: '/login', redirect: '/' },
    { path: '/register', redirect: '/' },
    {
      path: '/bigfive',
      name: 'bigfive-quiz',
      component: BigFiveQuiz,
      meta: { auth: true },
    },
    {
      path: '/bigfive/:id',
      name: 'bigfive-result',
      component: BigFiveResult,
      meta: { auth: true },
    },
    {
      path: '/onboarding/assessment',
      name: 'initial-assessment',
      component: InitialAssessment,
      meta: { auth: true },
    },
    {
      path: '/onboarding/report/:id?',
      name: 'assessment-report',
      component: AssessmentReport,
      meta: { auth: true },
    },
    {
      path: '/onboarding/need',
      name: 'need-select',
      component: NeedSelect,
      meta: { auth: true },
    },
    { path: '/assessments', name: 'my-assessments', component: MyAssessments, meta: { auth: true } },
    {
      path: '/assessments/:id',
      name: 'assessment-detail',
      component: AssessmentReport,
      meta: { auth: true },
    },
    { path: '/booking', name: 'booking', component: CounselingBooking, meta: { auth: true } },
    { path: '/home', name: 'coach-home', component: CoachHome, meta: { auth: true } },
    { path: '/workbench', redirect: '/home' },
    { path: '/coach/:id', name: 'coach-session', component: CoachSession, meta: { auth: true } },
    { path: '/wellbeing', redirect: '/assessments' },
    { path: '/wellbeing/quick', name: 'quick-self-check', component: QuickSelfCheck, meta: { auth: true } },
    { path: '/settings', name: 'settings', component: Settings, meta: { auth: true } },
    { path: '/tasks', name: 'tasks', component: Workbench, meta: { auth: true } },
    { path: '/tasks/new', name: 'task-new', component: TaskNew, meta: { auth: true } },
    { path: '/tasks/:id', name: 'task-detail', component: TaskDetail, props: true, meta: { auth: true } },
  ],
})

let profileCache = null
let profileCacheAt = 0

export function clearProfileCache() {
  profileCache = null
  profileCacheAt = 0
}

export async function loadProfile(force = false) {
  const now = Date.now()
  if (!force && profileCache && now - profileCacheAt < 15000) return profileCache
  profileCache = await api.me()
  profileCacheAt = now
  return profileCache
}

router.beforeEach(async (to) => {
  const loggedIn = isLoggedIn()
  if (to.meta.auth && !loggedIn) {
    return { path: '/', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && loggedIn) {
    const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : ''
    if (redirect) return redirect
    return postLoginPath()
  }
  return true
})

export default router
