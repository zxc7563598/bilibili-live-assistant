import { request } from '@/utils'

export default {
  importData: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request.post('/upload/import', fd, { timeout: 300000 })
  },
  getImportProgress: taskId => request.post('/upload/import/progress', { task_id: taskId }, { needTip: false }),
}
