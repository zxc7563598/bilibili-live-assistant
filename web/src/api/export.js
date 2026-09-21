import { request } from '@/utils'

export default {
  /**
   * 领取数据导出下载凭证。
   *
   * 所有校验（模块、列、行数上限、并发数）都在这一步完成，因此会把失败原因
   * 正常返回给前端；后续的下载接口走浏览器原生下载，出错信息无法展示。
   *
   * 单独覆盖 12 秒超时：百万行表上的命中行数统计可能更久。
   */
  createTicket: data => request.post('/export/ticket', data, { timeout: 60000 }),
}
