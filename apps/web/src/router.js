import { createRouter, createWebHistory } from 'vue-router'
import CoachHome from './views/CoachHome.vue'
import CoachSession from './views/CoachSession.vue'
import Wellbeing from './views/Wellbeing.vue'
import Settings from './views/Settings.vue'
import Workbench from './views/Workbench.vue'
import TaskNew from './views/TaskNew.vue'
import TaskDetail from './views/TaskDetail.vue'
import Login from './views/Login.vue'
import { isLoggedIn } from './auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'login', component: Login, meta: { guest: true } },
    { path: '/login', redirect: '/' },
    { path: '/register', redirect: '/' },
    { path: '/home', name: 'coach-home', component: CoachHome, meta: { auth: true } },
    { path: '/workbench', redirect: '/home' },
    { path: '/coach/:id', name: 'coach-session', component: CoachSession, meta: { auth: true } },
    { path: '/wellbeing', name: 'wellbeing', component: Wellbeing, meta: { auth: true } },
    { path: '/settings', name: 'settings', component: Settings, meta: { auth: true } },
    { path: '/tasks', name: 'tasks', component: Workbench, meta: { auth: true } },
    { path: '/tasks/new', name: 'task-new', component: TaskNew, meta: { auth: true } },
    { path: '/tasks/:id', name: 'task-detail', component: TaskDetail, props: true, meta: { auth: true } },
  ],
})

router.beforeEach((to) => {
  const loggedIn = isLoggedIn()
  if (to.meta.auth && !loggedIn) {
    return { path: '/', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && loggedIn) {
    return { path: '/home' }
  }
  return true
})

export default router
