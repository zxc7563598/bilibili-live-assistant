# CLAUDE.md — 前端开发指南

## 技术栈

- **框架**：Vue 3（Composition API）
- **UI 库**：Naive UI
- **构建工具**：Vite
- **状态管理**：Pinia
- **HTTP 请求**：Axios（封装在 `src/utils/http`）
- **图标**：Iconify（`@iconify/vue`）+ 自定义 SVG

本项目基于 [vue-naive-admin](https://github.com/zclzone/vue-naive-admin) 二次开发。

## 目录结构

```
web/src
├── api/              # 接口请求封装
├── assets/           # 静态资源（图标、图片）
├── components/       # 公共组件
│   ├── common/       # 通用组件（AppPage、CommonPage、AppCard 等）
│   └── me/           # 业务模板组件（MeCrud、MeModal、MeQueryItem）
├── composables/      # 组合式函数（useAliveData、useCrud、useForm、useModal）
├── directives/       # 自定义指令
├── layouts/          # 布局系统（full/normal/simple/empty）
├── router/           # 路由配置 + 守卫
├── store/            # Pinia 状态管理
├── styles/           # 全局样式
├── utils/            # 工具方法（HTTP 封装、存储等）
└── views/            # 页面视图（按模块分目录）
```

## 核心开发模式

### 1. 页面容器选择

**优先使用 `CommonPage`**（标准业务页面），仅特殊场景用 `AppPage`：

| 场景                           | 容器                               |
| ------------------------------ | ---------------------------------- |
| 列表页、表单页、标准 CRUD 页面 | `CommonPage`（内置标题、卡片包裹） |
| 仪表盘、大屏、高度自定义布局   | `AppPage`（空白画布，完全自控）    |

```html
<template>
  <CommonPage>
    <!-- 页面内容 -->
  </CommonPage>
</template>
```

### 2. 标准 CRUD 页面模板

页面 = CommonPage + MeCrud + MeModal + composables：

```html
<template>
  <CommonPage>
    <!-- 列表 -->
    <MeCrud ref="crudRef" :columns="columns" :getData="api.read" v-model:queryItems="query">
      <MeQueryItem label="用户名">
        <n-input v-model:value="query.username" />
      </MeQueryItem>
    </MeCrud>

    <!-- 弹窗 -->
    <MeModal ref="modalRef">
      <n-form ref="modalFormRef" :model="modalForm">
        <!-- 表单字段 -->
      </n-form>
    </MeModal>
  </CommonPage>
</template>

<script setup>
  import { useAliveData, useCrud } from '@/composables'

  // 查询条件（自动缓存）
  const { aliveData: query } = useAliveData({ username: null })

  // 表格列定义
  const columns = [
    { title: '用户名', key: 'username' },
    { title: '状态', key: 'status' },
    {
      title: '操作',
      key: 'actions',
      render(row) {
        /* ... */
      },
    },
  ]

  // CRUD 封装（自动处理新增/编辑/删除/loading）
  const crudRef = ref(null)
  const { modalRef, modalFormRef, modalForm, handleAdd, handleEdit, handleDelete } = useCrud({
    name: '用户',
    initForm: { username: '', status: 1 },
    doCreate: api.create,
    doUpdate: api.update,
    doDelete: api.delete,
    refresh: () => crudRef.value?.handleSearch(),
  })
</script>
```

### 3. 页面开发步骤

1. 在 `views/<模块>/` 下创建页面目录和 `index.vue`
2. 使用 `CommonPage` 作为外层容器
3. 用 `MeCrud` + `MeQueryItem` 搭建列表
4. 用 `MeModal` + Naive UI 表单组件搭建弹窗
5. 用 `useCrud` / `useForm` / `useModal` / `useAliveData` 管理状态
6. 接口封装放在同目录的 `api.js` 或 `src/api/` 中
7. **不需要手动配前端路由** — 在后台「菜单管理」中配置路径即可

### 4. Composables 用法速查

| Composable                     | 用途                         | 关键返回值                                                          |
| ------------------------------ | ---------------------------- | ------------------------------------------------------------------- |
| `useAliveData(initData, key?)` | 缓存查询条件，页面返回时恢复 | `aliveData`（响应式）, `reset()`                                    |
| `useForm(initForm?)`           | 表单数据 + 校验              | `[formRef, formModel, validation, rules]`                           |
| `useModal()`                   | 弹窗控制 + loading           | `[modalRef, okLoading]`                                             |
| `useCrud({...})`               | 完整 CRUD 编排               | `{ modalRef, modalForm, handleAdd, handleEdit, handleDelete, ... }` |

`useCrud` 内部已整合 `useModal` + `useForm`，自动处理新增/编辑/查看模式切换、弹窗标题、loading 状态。

## 核心组件 Props 速查

### MeCrud

| Prop             | 说明                                                                                               |
| ---------------- | -------------------------------------------------------------------------------------------------- |
| `columns` (必传) | Naive UI 表格列配置                                                                                |
| `getData` (必传) | 数据请求方法，自动传 `{ ...queryItems, pageNo, pageSize }`，需返回 `{ data: { pageData, total } }` |
| `queryItems`     | 查询参数（v-model）                                                                                |
| `remote`         | 是否后端分页（默认 true）                                                                          |

暴露方法（通过 ref）：`handleSearch()`、`handleReset()`、`handleExport()`

### MeModal

通过 `modalRef.value.open({ title, onOk, ... })` 打开，`onOk` 返回 `false` 可阻止关闭。

## 路由与菜单

项目采用**动态路由**：页面开发完成后，在后台「菜单管理」中配置路径与 Vue 页面的绑定即可，无需手动维护前端路由表。

## 新增模块检查清单

- [ ] 在 `views/<模块>/` 创建页面文件
- [ ] 使用 CommonPage + MeCrud + MeModal 模式
- [ ] 封装 API 请求（在各自模块的`api.js`）
- [ ] 在后台菜单管理中配置路由
