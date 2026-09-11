// 商品规格 / SKU 组合的纯函数工具集

// 组合内 pair 之间的分隔符
export const SEP = '\u0001'
// 单个 pair 的 key 与 value 之间的分隔符
export const PAIR_SEP = '\u0002'
// 组合数上限，超过就没法在页面上正常编辑了
export const MAX_COMBOS = 200
// 规格类型数量上限，与后端 binding:"max=10" 保持一致
export const MAX_SPECS = 10
// 商品图片总数上限（轮播图 + 详情图），与后端 binding:"max=200" 保持一致
export const MAX_IMAGES = 200
// orphanStash 中按规格值 uid 兜底查找的键前缀；
// 组合键一定以规格名开头，用 PAIR_SEP 打头不会与组合键撞上
const STASH_UID_PREFIX = `${PAIR_SEP}uid${PAIR_SEP}`

let uidSeed = 0

// 生成页面内唯一的前端身份
export function nextUid(prefix) {
  uidSeed += 1
  return `${prefix}_${uidSeed.toString(36)}`
}

// 把一组 [{键:值}] 压成可比较的字符串键
export function propsKey(props) {
  return props
    .map((p) => {
      const k = Object.keys(p)[0]
      return k + PAIR_SEP + p[k]
    })
    .join(SEP)
}

// 组合的可读标签，如 '镭射款 / wink小蓝'；无规格时为「默认规格」
export function propsLabel(props) {
  return props.map(p => Object.values(p)[0]).join(' / ') || '默认规格'
}

/**
 * 解析 spec_properties。兼容 ''、'[]'、'null'、畸形 JSON 以及已经是数组的情况。
 * @returns {Array<object>} 永远是数组，绝不抛错
 */
export function parseSpecProperties(raw) {
  if (Array.isArray(raw))
    return raw
  if (typeof raw !== 'string' || !raw)
    return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  }
  catch {
    return []
  }
}

/**
 * 把任意顺序的 spec_properties 归一化成「按当前规格类型顺序」的对象数组。
 * 不在 specKeys 里的键会被丢弃（对应当前已不存在的规格类型）。
 */
export function normalizeProps(raw, specKeys) {
  const map = {}
  for (const pair of raw) {
    if (pair && typeof pair === 'object' && !Array.isArray(pair)) {
      for (const k of Object.keys(pair))
        map[k] = String(pair[k])
    }
  }
  return specKeys
    .filter(k => map[k] !== undefined)
    .map(k => ({ [k]: map[k] }))
}

/**
 * 笛卡尔积。空规格名 / 空规格值会被跳过，交由校验层报具体问题。
 * @returns {Array<{ props: Array<object>, uids: string[] }>} 无规格时返回单个空组合；某规格类型没有值时返回空数组
 */
export function buildCombos(specs) {
  const lists = specs
    .filter(s => s.key_name.trim() !== '')
    .map(s => s.values
      .filter(v => v.value_name.trim() !== '')
      .map(v => ({ key: s.key_name.trim(), value: v.value_name.trim(), uid: v.uid })))

  // 完全没有规格：给一条默认组合
  if (lists.length === 0)
    return [{ props: [], uids: [] }]
  // 有规格类型却没有值：不生成组合
  if (lists.some(l => l.length === 0))
    return []

  return lists.reduce(
    (rows, list) => rows.flatMap(r => list.map(item => ({
      props: [...r.props, { [item.key]: item.value }],
      uids: [...r.uids, item.uid],
    }))),
    [{ props: [], uids: [] }],
  )
}

/** 组合总数；返回 0 表示有规格类型没有值，1 表示无规格的默认组合 */
export function combosCount(specs) {
  const lens = specs
    .filter(s => s.key_name.trim() !== '')
    .map(s => s.values.filter(v => v.value_name.trim() !== '').length)
  if (lens.length === 0)
    return 1
  return lens.reduce((n, l) => n * l, 1)
}

/**
 * 按当前规格重新生成 SKU 行，并尽量保留上一轮已填的 price / stock / enabled。
 *
 * 两层匹配：
 * - key（名称字符串）——加载时与历史 SKU 对齐；
 * - uidKey（规格值 uid 排序集合）——规格值改名或规格类型换序后，仍能找回上一轮的行，
 *   避免用户刚填的价格库存被清空。
 *
 * @param {Array} combos       buildCombos(specs) 的结果
 * @param {Array} prevRows     当前 skuRows
 * @param {number} defaultPrice 新组合的价格预填值（商品展示价）
 * @param {Map} stash          可选。被移除过的行，同时按 key 与 uidKey 记录，供后续恢复
 * @returns {{ rows: Array, orphans: Array }} rows 对齐 combos 的新行；orphans 是被淘汰的旧行
 */
export function reconcileSkuRows(combos, prevRows, defaultPrice, stash) {
  const byKey = new Map()
  const byUid = new Map()
  for (const r of prevRows) {
    if (r.key)
      byKey.set(r.key, r)
    if (r.uidKey)
      byUid.set(r.uidKey, r)
  }

  const consumed = new Set()
  const rows = []
  const noSpecs = combos.length === 1 && combos[0].props.length === 0

  for (const { props, uids } of combos) {
    const key = propsKey(props)
    const uidKey = [...uids].sort().join(SEP)

    let hit = byKey.get(key) || (uidKey ? byUid.get(uidKey) : null) || null
    if (hit && consumed.has(hit))
      hit = null
    if (hit)
      consumed.add(hit)

    if (hit) {
      rows.push({ ...hit, key, uidKey, specProperties: props })
      continue
    }

    // 该组合之前被移除过（例如删掉规格值又加回来），优先恢复用户填过的值。
    // key 对不上时按 uidKey 兜底：规格值改名 / 规格类型改名都会让 key 变化，但 uid 不变
    const stashed = stash?.get(key)
      || (uidKey ? stash?.get(`${STASH_UID_PREFIX}${uidKey}`) : null)
      || null
    rows.push({
      key,
      uidKey,
      specProperties: props,
      price: stashed ? stashed.price : Number(defaultPrice) || 0,
      costPrice: stashed ? stashed.costPrice ?? 0 : 0,
      stock: stashed ? stashed.stock : 0,
      // 恢复出来的行沿用原来的库存基线；新建的行没有基线，保存时会照实提交库存
      originalStock: stashed ? stashed.originalStock ?? null : null,
      enabled: stashed ? stashed.enabled : noSpecs,
    })
  }

  const orphans = prevRows.filter(r => !consumed.has(r))
  if (stash) {
    for (const o of orphans) {
      stash.set(o.key, o)
      if (o.uidKey)
        stash.set(`${STASH_UID_PREFIX}${o.uidKey}`, o)
    }
  }

  return { rows, orphans }
}
