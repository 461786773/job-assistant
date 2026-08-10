import { createRouter, createWebHistory } from 'vue-router'
import Workbench from './views/Workbench.vue'
import TaskNew from './views/TaskNew.vue'
import TaskDetail from './views/TaskDetail.vue'
import Login from './views/Login.vue'
import Register from './views/Register.vue'
import { isLoggedIn } from './auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: Login, meta: { guest: true } },
    { path: '/register', name: 'register', component: Register, meta: { guest: true } },
    { path: '/', name: 'workbench', component: Workbench, meta: { auth: true } },
    { path: '/tasks/new', name: 'task-new', component: TaskNew, meta: { auth: true } },
    { path: '/tasks/:id', name: 'task-detail', component: TaskDetail, props: true, meta: { auth: true } },
  ],
})

router.beforeEach((to) => {
  const loggedIn = isLoggedIn()
  if (to.meta.auth && !loggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && loggedIn) {
    return { path: '/' }
  }
  return true
})

export default router
