import { createRouter, createWebHistory } from 'vue-router'

const defaultMeta = [
  // { name: "theme-color", content: "#ffffff" }
]

const defaultTitle = import.meta.env.VITE_APP_NAME
const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../pages/login/index.vue'),
    meta: {
      title: defaultTitle,
      metaTags: defaultMeta,
    },
  },
  {
    path: '/',
    name: 'home',
    component: () => import('../pages/home/index.vue'),
    meta: {
      title: defaultTitle,
      metaTags: defaultMeta,
    },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('../pages/profile/index.vue'),
    meta: {
      title: defaultTitle,
      metaTags: defaultMeta,
    },
  },
  {
  // 匹配为定义路由然后重定向到404页面
    path: '/:pathMath(.*)',
    redirect: '/',
  },
]

// 设置路由
const router = createRouter({
  routes,
  history: createWebHistory('/shop'),
})

// 导出路由
export default router
