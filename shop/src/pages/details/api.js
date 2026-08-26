export default {
  getShopDetails: (id) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            id,
            name: '无线蓝牙耳机 主动降噪',
            slides: [
              { src: 'https://shub.points.xin/attachment/goods/cover_image/image_681ad8392a90b5.30457608.png', label: '轮播图 1' },
              { src: '', label: '轮播图 2' },
              { src: 'https://danmusuite.hejunjie.life/dist/avatar.jpg', label: '轮播图 3' },
            ],
            amount: 12800,
            type: 0,
            sold: 2341,
            stock: 99,
            tags: ['热卖', '包邮'],
            skuGroups: [
              { name: '颜色', options: ['星空灰', '月光白', '曜石黑'] },
              { name: '版本', options: ['标准版', 'Pro 版'] },
            ],
            desc: '一款主打舒适的主动降噪耳机，支持多设备连接。',
            detail_images: [
              { src: 'https://shub.points.xin/attachment/goods/cover_image/image_681ad8392a90b5.30457608.png', label: '商品说明图 1' },
              { src: 'https://shub.points.xin/attachment/goods/cover_image/image_681ad8392a90b5.30457608.png', label: '商品说明图 2' },
              { src: '', label: '商品说明图 3' },
            ],
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
}
