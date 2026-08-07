import { createRouter, createWebHistory } from 'vue-router'
import Workbench from './views/Workbench.vue'
import TaskNew from './views/TaskNew.vue'
import TaskDetail from './views/TaskDetail.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'workbench', component: Workbench },
    { path: '/tasks/new', name: 'task-new', component: TaskNew },
    { path: '/tasks/:id', name: 'task-detail', component: TaskDetail, props: true },
  ],
})

export default router
