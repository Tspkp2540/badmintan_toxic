import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { guest: true },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/DashboardView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/ranking',
    name: 'Ranking',
    component: () => import('@/views/RankingView.vue'),
    // ranking เป็น public route ตาม BE (ไม่ต้อง auth)
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/court',
    redirect: '/courts',
  },
  {
    path: '/courts',
    name: 'Courts',
    component: () => import('@/views/CourtListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/court/:id',
    name: 'Court',
    component: () => import('@/views/CourtView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/skill-guide',
    name: 'SkillGuide',
    component: () => import('@/views/SkillGuideView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/manage-users',
    name: 'ManageUsers',
    component: () => import('@/views/ManageUsersView.vue'),
    meta: { requiresAuth: true, requiredRoles: ['admin', 'leader', 'vice_leader'] },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let sessionInitialized = false

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // ครั้งแรกที่เข้า app ให้ restore session ก่อน
  if (!sessionInitialized) {
    sessionInitialized = true
    await authStore.initSession()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.meta.guest && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
  } else if (to.meta.requiredRoles) {
    const roles = to.meta.requiredRoles as string[]
    if (!roles.includes(authStore.userRole)) {
      next({ name: 'Dashboard' })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
