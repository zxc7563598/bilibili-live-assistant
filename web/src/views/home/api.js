import { request } from '@/utils'

export default {
  // 检查更新不通时不弹提示，版本信息静默降级
  getVersionInfo: () => request.post('/version/info', {}, { needTip: false }),
  importData: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request.post('/upload/import', fd, { timeout: 300000 })
  },
  getImportProgress: taskId => request.post('/upload/import/progress', { task_id: taskId }, { needTip: false }),
}
