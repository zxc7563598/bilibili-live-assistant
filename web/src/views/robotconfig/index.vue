<template>
  <AppPage show-footer full>
    <div class="flex gap-10">
      <div class="min-w-0 flex-1 space-y-10">
        <n-card v-show="roomConfigShow" id="module-basic" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                基础设置
              </div>
              <div class="mt-2 text-13 text-gray-400">
                配置机器人基础设置
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「启动时自动监听」
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                配置开启后, 项目因任何原因重启时, 若已经配置了登录信息与直播间, 则自动进行监听
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                启动时自动监听
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="roomConfigForm.is_listening === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="roomConfigForm.is_listening = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                用户名最大长度
              </div>
              <n-input v-model:value="roomConfigForm.max_name_length" type="number" placeholder="0 为不做任何限制" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                裁剪模式
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in nameTrimModeEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="roomConfigForm.name_trim_mode === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="roomConfigForm.name_trim_mode = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                消费奖励
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in consumeRewardEnabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="roomConfigForm.consume_reward_enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="roomConfigForm.consume_reward_enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                奖励类型
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in rewardTypeEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="roomConfigForm.reward_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="roomConfigForm.reward_type = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div v-show="roomConfigForm.consume_reward_enabled === '1'" class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                转换倍率
              </div>
              <n-input-group>
                <n-input v-model:value="roomConfigForm.consume_battery_rate" type="number" placeholder="设置为 2 则代表用户消耗 1 电池会得到 2 点奖励" />
                <n-input-group-label v-show="roomConfigForm.reward_type === '0'">
                  星光
                </n-input-group-label>
                <n-input-group-label v-show="roomConfigForm.reward_type === '1'">
                  积分
                </n-input-group-label>
              </n-input-group>
            </div>
            <div v-show="roomConfigForm.consume_reward_enabled === '2'" class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                舰长奖励
              </div>
              <n-input-group>
                <n-input v-model:value="roomConfigForm.captain_reward_amount" type="number" placeholder="奖励设置为按开通航海类型发放时生效" />
                <n-input-group-label v-show="roomConfigForm.reward_type === '0'">
                  星光
                </n-input-group-label>
                <n-input-group-label v-show="roomConfigForm.reward_type === '1'">
                  积分
                </n-input-group-label>
              </n-input-group>
            </div>
            <div v-show="roomConfigForm.consume_reward_enabled === '2'" class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                提督奖励
              </div>
              <n-input-group>
                <n-input v-model:value="roomConfigForm.commander_reward_amount" type="number" placeholder="奖励设置为按开通航海类型发放时生效" />
                <n-input-group-label v-show="roomConfigForm.reward_type === '0'">
                  星光
                </n-input-group-label>
                <n-input-group-label v-show="roomConfigForm.reward_type === '1'">
                  积分
                </n-input-group-label>
              </n-input-group>
            </div>
            <div v-show="roomConfigForm.consume_reward_enabled === '2'" class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                总督奖励
              </div>
              <n-input-group>
                <n-input v-model:value="roomConfigForm.governor_reward_amount" type="number" placeholder="奖励设置为按开通航海类型发放时生效" />
                <n-input-group-label v-show="roomConfigForm.reward_type === '0'">
                  星光
                </n-input-group-label>
                <n-input-group-label v-show="roomConfigForm.reward_type === '1'">
                  积分
                </n-input-group-label>
              </n-input-group>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="roomLoading" @click="apply('basic')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="signShow" id="module-sign" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                签到配置模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                签到相关配置, 用户可以发送特定指令进行签到, 并得到相对应的奖励
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="signForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="signForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="signForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="signForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="signForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="signForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                奖励类型
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in rewardTypeEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="signForm.reward_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="signForm.reward_type = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                奖励数量
              </div>
              <n-input v-model:value="signForm.reward_amount" type="number" placeholder="0 则不进行任何奖励" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                签到关键词
              </div>
              <n-input v-model:value="signForm.keyword" placeholder="用户触发签到的词, 建议增加符号以避免错误触发" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                查询关键词
              </div>
              <n-input v-model:value="signForm.query_keyword" placeholder="用户触发查询的词, 建议增加符号以避免错误触发" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                签到配置相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                用户名：<b>@name@</b><br>
                总签到天数：<b>@days@</b><br>
                连续签到天数：<b>@streak@</b><br>
                用户积分：<b>@points@</b><br>
                用户星光：<b>@stars@</b><br>
                用户航海类型：<b>@guard@</b><br>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                签到成功回复
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in signForm.success_reply" v-show="signForm.success_reply.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="signForm.success_reply[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="signForm.success_reply.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="signForm.success_reply.push('')">
                  添加
                </n-button>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                签到失败回复
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in signForm.fail_reply" v-show="signForm.fail_reply.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="signForm.fail_reply[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="signForm.fail_reply.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="signForm.fail_reply.push('')">
                  添加
                </n-button>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                重复签到回复
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in signForm.repeat_reply" v-show="signForm.repeat_reply.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="signForm.repeat_reply[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="signForm.repeat_reply.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="signForm.repeat_reply.push('')">
                  添加
                </n-button>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                查询成功回复
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in signForm.query_reply" v-show="signForm.query_reply.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="signForm.query_reply[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="signForm.query_reply.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="signForm.query_reply.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="signLoading" @click="apply('sign')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="adShow" id="module-ad" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                定时广告模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                定时广告配置, 机器人会在指定时间间隔向直播间发送相应的内容
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="adForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="adForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="adForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="adForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送间隔
              </div>
              <n-input-group>
                <n-input v-model:value="adForm.interval" type="number" placeholder="发送间隔" />
                <n-input-group-label>秒</n-input-group-label>
              </n-input-group>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送方式
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sendModeEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="adForm.send_mode === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="adForm.send_mode = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in adForm.content" v-show="adForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="adForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="adForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="adForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="adLoading" @click="apply('ad')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="giftShow" id="module-gift" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                礼物答谢模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                礼物答谢配置, 用户赠送礼物后由机器人发送弹幕进行感谢
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于展示数量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                开启后, 礼物信息动态变量中会携带数量信息, 类似于: 「x个礼物名称」
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                展示数量
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.show_count === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.show_count = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于礼物合并
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                开启后, 礼物信息动态变量中会携带多个礼物信息, 类似于: 「x个礼物名称、x个礼物名称、x个礼物名称」<br>
                主要用于处理用户短时间赠送大量盲盒产生多种礼物的情况, 开启后可以防止机器人感谢刷屏
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                礼物合并
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.merge_gift === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.merge_gift = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于盲盒统计
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                开启后, 机器人在答谢用户赠送的盲盒礼物信息时会在感谢信息最后追加本次盲盒盈亏信息, 类似于：「 | 赚了1.5元」、「 | 亏了3元」<br>
                需要注意答谢内容, 过长的内容将会被拆成多条发送
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                盲盒统计
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="giftForm.include_blindbox === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="giftForm.include_blindbox = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                起始感谢电池数
              </div>
              <n-input-group>
                <n-input v-model:value="giftForm.min_battery" type="number" placeholder="低于此电池数的礼物不触发感谢" />
                <n-input-group-label>电池</n-input-group-label>
              </n-input-group>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                礼物答谢相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                送礼人名称：<b>@name@</b><br>
                礼物信息：<b>@gift@</b><br>
                礼物总价(人民币)：<b>@price@</b><br>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in giftForm.content" v-show="giftForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="giftForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="giftForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="giftForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="giftLoading" @click="apply('gift')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="pkShow" id="module-pk" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                PK播报模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                PK播报配置, 机器人在PK开始前向直播间内广播PK对手信息
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于PK播报
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                可以设置多条信息, 进行 PK 播报时将会按照顺序发送内容
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                PK 播报相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                对方主播名称：<b>@anchor@</b><br>
                对方高能榜人数：<b>@online_num@</b><br>
                对方高能榜总计贡献度：<b>@online_score@</b><br>
                对方高能榜前三名贡献度：<b>@top3_score@</b><br>
                对方舰长总数：<b>@vip_num@</b><br>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="pkForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="pkForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in pkForm.content" v-show="pkForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="pkForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="pkForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="pkForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="pkLoading" @click="apply('pk')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="welcomeShow" id="module-welcome" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                进房欢迎模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                进房欢迎配置, 用户进入直播间后由机器人发送弹幕进行欢迎
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="welcomeForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="welcomeForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="welcomeForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="welcomeForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="welcomeForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="welcomeForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                进房欢迎相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                用户名称：<b>@name@</b><br>
                用户航海类型：<b>@guard@</b><br>
                用户累计进房次数：<b>@total_times@</b><br>
                用户累计进房天数：<b>@total_days@</b><br>
                用户连续进房天数：<b>@streak@</b><br>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in welcomeForm.content" v-show="welcomeForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="welcomeForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="welcomeForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="welcomeForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="welcomeLoading" @click="apply('welcome')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="followShow" id="module-follow" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                感谢关注模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                感谢关注配置, 用户关注直播间后由机器人发送弹幕进行感谢
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="followForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="followForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="followForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="followForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="followForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="followForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                感谢关注相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                用户名称：<b>@name@</b><br>
                用户航海类型：<b>@guard@</b><br>
                用户累计关注次数：<b>@total_times@</b><br>
                用户累计关注天数：<b>@total_days@</b><br>
                用户连续关注天数：<b>@streak@</b><br>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in followForm.content" v-show="followForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="followForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="followForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="followForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="followLoading" @click="apply('follow')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="shareShow" id="module-share" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                感谢分享模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                感谢分享配置, 用户分享直播间后由机器人发送弹幕进行感谢
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="shareForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="shareForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="shareForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="shareForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="shareForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="shareForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                感谢分享相关占位变量
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                用户名称：<b>@name@</b><br>
                用户航海类型：<b>@guard@</b><br>
                用户累计分享次数：<b>@total_times@</b><br>
                用户累计分享天数：<b>@total_days@</b><br>
                用户连续分享天数：<b>@streak@</b><br>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                发送内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(_reply, index) in shareForm.content" v-show="shareForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input v-model:value="shareForm.content[index]" placeholder="支持占位符变量" show-count :maxlength="99" />
                  <n-button secondary type="error" size="medium" @click="shareForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="shareForm.content.push('')">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="shareLoading" @click="apply('share')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
        <n-card v-show="replyShow" id="module-reply" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                自动回复模块
              </div>
              <div class="mt-2 text-13 text-gray-400">
                自动回复配置, 用户发送弹幕命中关键词后由机器人进行回复
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                是否开启
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyForm.enabled === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyForm.enabled = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                可用场景
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in sceneEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyForm.scene === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyForm.scene = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                触发门槛
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in requirementEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyForm.requirement === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyForm.requirement = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                配置内容
              </div>
              <div class="w-full space-y-5">
                <div v-for="(reply, index) in replyForm.content" v-show="replyForm.content.length > 0" :key="index" class="w-full flex gap-5">
                  <n-input :value="reply.keyword.join(', ')" placeholder="" disabled />
                  <n-button secondary type="primary" size="medium" @click="replyConfig(index)">
                    配置
                  </n-button>
                  <n-button secondary type="error" size="medium" @click="replyForm.content.splice(index, 1)">
                    删除
                  </n-button>
                </div>
                <n-button secondary type="primary" size="medium" class="w-full" @click="replyConfig(replyForm.content.length)">
                  添加
                </n-button>
              </div>
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="replyLoading" @click="apply('reply')">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
      </div>
      <aside class="hidden w-200 shrink-0 lg:block">
        <div class="sticky top-5 border border-gray-200 rounded-3 bg-white p-10 dark:border-gray-700 dark:bg-gray-900">
          <div class="mb-3 text-center text-15 text-gray-500 font-medium">
            配置模块
          </div>
          <div class="mt-10 space-y-5">
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-basic')">
              基础设置
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-sign')">
              签到配置
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-ad')">
              定时广告
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-gift')">
              礼物答谢
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-pk')">
              PK播报
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-welcome')">
              进房欢迎
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-follow')">
              感谢关注
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-share')">
              感谢分享
            </div>
            <div class="cursor-pointer border border-gray-200 rounded-3 px-3 py-2 text-center text-13 text-gray-600 transition dark:border-gray-700 hover:border-primary dark:text-gray-300 hover:text-primary" @click="scrollTo('module-reply')">
              自动回复
            </div>
          </div>
        </div>
      </aside>
    </div>
    <n-modal v-model:show="replyConfigModel" title="自动回复内容配置" preset="card" style="width: 640px" :mask-closable="false">
      <div class="space-y-10">
        <div class="flex items-start gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            关键词
          </div>
          <div class="w-full space-y-5">
            <div v-for="(_reply, index) in replyConfigForm.keyword" v-show="replyConfigForm.keyword.length > 0" :key="index" class="w-full flex gap-5">
              <n-input v-model:value="replyConfigForm.keyword[index]" placeholder="设置多个时只要触发一个就会触发" />
              <n-button secondary type="error" size="medium" @click="replyConfigForm.keyword.splice(index, 1)">
                删除
              </n-button>
            </div>
            <n-button secondary type="primary" size="medium" class="w-full" @click="replyConfigForm.keyword.push('')">
              添加
            </n-button>
          </div>
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            关键词匹配方式
          </div>
          <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
            <div v-for="item in matchPolicyEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyConfigForm.keyword_match_policy === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyConfigForm.keyword_match_policy = item.value">
              {{ item.label }}
            </div>
          </div>
        </div>
        <div class="flex items-start gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            安全词
          </div>
          <div class="w-full space-y-5">
            <div v-for="(_reply, index) in replyConfigForm.safe_word" v-show="replyConfigForm.safe_word.length > 0" :key="index" class="w-full flex gap-5">
              <n-input v-model:value="replyConfigForm.safe_word[index]" placeholder="设置多个时只要触发一个就会触发" />
              <n-button secondary type="error" size="medium" @click="replyConfigForm.safe_word.splice(index, 1)">
                删除
              </n-button>
            </div>
            <n-button secondary type="primary" size="medium" class="w-full" @click="replyConfigForm.safe_word.push('')">
              添加
            </n-button>
          </div>
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            安全词匹配方式
          </div>
          <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
            <div v-for="item in matchPolicyEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyConfigForm.safe_word_match_policy === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyConfigForm.safe_word_match_policy = item.value">
              {{ item.label }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            命中关键词禁言
          </div>
          <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
            <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="replyConfigForm.mute_sender === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="replyConfigForm.mute_sender = item.value">
              {{ item.label }}
            </div>
          </div>
        </div>
        <div v-show="replyConfigForm.mute_sender === '1'" class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            禁言时长
          </div>
          <n-input-group>
            <n-input v-model:value="replyConfigForm.mute_duration" type="number" placeholder="0 则表示永久" />
            <n-input-group-label>分钟</n-input-group-label>
          </n-input-group>
        </div>
        <div v-show="replyConfigForm.mute_sender === '1'" class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            赎回金额
          </div>
          <n-input-group>
            <n-input v-model:value="replyConfigForm.ransom_amount" type="number" placeholder="0 则表示不可赎回" />
            <n-input-group-label>电池</n-input-group-label>
          </n-input-group>
        </div>
        <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
          <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
            自动回复相关占位变量
          </div>
          <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
            发送人昵称：<b>@name@</b><br>
            用户航海类型：<b>@guard@</b><br>
            用户本日盲盒盈利：<b>@daily_net@</b><br>
            用户本周盲盒盈利：<b>@weekly_net@</b><br>
            用户本月盲盒盈利：<b>@monthly_net@</b><br>
            用户总计盲盒盈利：<b>@total_net@</b><br>
            直播间本日盲盒盈利：<b>@room_daily_net@</b><br>
            直播间本周盲盒盈利：<b>@room_weekly_net@</b><br>
            直播间本月盲盒盈利：<b>@room_monthly_net@</b><br>
            直播间总计盲盒盈利：<b>@room_total_net@</b><br>
          </div>
        </div>
        <div class="flex items-start gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            回复内容
          </div>
          <div class="w-full space-y-5">
            <div v-for="(_reply, index) in replyConfigForm.reply_content" v-show="replyConfigForm.reply_content.length > 0" :key="index" class="w-full flex gap-5">
              <n-input v-model:value="replyConfigForm.reply_content[index]" placeholder="设置多个时随机抽取一条进行回复, 支持占位符变量" show-count :maxlength="99" />
              <n-button secondary type="error" size="medium" @click="replyConfigForm.reply_content.splice(index, 1)">
                删除
              </n-button>
            </div>
            <n-button secondary type="primary" size="medium" class="w-full" @click="replyConfigForm.reply_content.push('')">
              添加
            </n-button>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <n-button size="small" type="primary" @click="saveReplyConfig">
            保存设置
          </n-button>
        </div>
      </template>
    </n-modal>
  </AppPage>
</template>

<script setup>
import api from './api'

// 基本配置模块
const roomLoading = ref(false)
const roomConfigShow = ref(false)
const roomConfigForm = ref({
  is_listening: '',
  max_name_length: '',
  name_trim_mode: '',
  consume_reward_enabled: '',
  reward_type: '',
  consume_battery_rate: '',
  captain_reward_amount: '',
  commander_reward_amount: '',
  governor_reward_amount: '',
})

// 签到配置模块
const signLoading = ref(false)
const signShow = ref(false)
const signForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  reward_type: '',
  reward_amount: '',
  keyword: '',
  query_keyword: '',
  success_reply: [],
  fail_reply: [],
  repeat_reply: [],
  query_reply: [],
})

// 定时广告配置模块
const adLoading = ref(false)
const adShow = ref(false)
const adForm = ref({
  enabled: '',
  scene: '',
  interval: '',
  send_mode: '',
  content: [],
})

// 礼物答谢配置模块
const giftLoading = ref(false)
const giftShow = ref(false)
const giftForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  show_count: '',
  merge_gift: '',
  include_blindbox: '',
  min_battery: '',
  content: [],
})

// PK配置模块
const pkLoading = ref(false)
const pkShow = ref(false)
const pkForm = ref({
  enabled: '',
  content: [],
})

// 进房欢迎模块
const welcomeLoading = ref(false)
const welcomeShow = ref(false)
const welcomeForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  content: [],
})

// 感谢关注配置模块
const followLoading = ref(false)
const followShow = ref(false)
const followForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  content: [],
})

// 感谢分享配置模块
const shareLoading = ref(false)
const shareShow = ref(false)
const shareForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  content: [],
})

// 自动回复配置模块
const replyLoading = ref(false)
const replyShow = ref(false)
const replyForm = ref({
  enabled: '',
  scene: '',
  requirement: '',
  content: [],
})
const replyConfigModel = ref(false)
const replyConfigIndex = ref(0)
const replyConfigForm = ref({
  keyword: [],
  keyword_match_policy: '0',
  safe_word: [],
  safe_word_match_policy: '0',
  mute_sender: '0',
  mute_duration: '0',
  ransom_amount: '0',
  reply_content: [],
})

// 通用部分
const enabledEnums = [
  {
    label: '关闭',
    value: '0',
  },
  {
    label: '开启',
    value: '1',
  },
]

const sceneEnums = [
  {
    label: '不限制',
    value: '0',
  },
  {
    label: '直播中',
    value: '1',
  },
  {
    label: '非直播中',
    value: '2',
  },
]

const requirementEnums = [
  {
    label: '不限制',
    value: '0',
  },
  {
    label: '带本直播间牌子',
    value: '1',
  },
  {
    label: '带本直播间大航海牌子',
    value: '2',
  },
]

const rewardTypeEnums = [
  {
    label: '星光',
    value: '0',
  },
  {
    label: '积分',
    value: '1',
  },
]

const nameTrimModeEnums = [
  {
    label: '省略后面',
    value: '0',
  },
  {
    label: '省略前面',
    value: '1',
  },
]

const sendModeEnums = [
  {
    label: '随机发送',
    value: '0',
  },
  {
    label: '顺序发送',
    value: '1',
  },
]

const matchPolicyEnums = [
  {
    label: '任意命中即触发',
    value: '0',
  },
  {
    label: '全部命中才触发',
    value: '1',
  },
]

const consumeRewardEnabledEnums = [
  {
    label: '不发放',
    value: '0',
  },
  {
    label: '按消费电池发放',
    value: '1',
  },
  {
    label: '按开通航海类型发放',
    value: '2',
  },
]

// 自动回复配置
function replyConfig(index) {
  try {
    const source = replyForm.value.content[index]
    if (source) {
      replyConfigForm.value = JSON.parse(JSON.stringify(source))
    }
    else {
      replyConfigForm.value = {
        keyword: [],
        keyword_match_policy: '0',
        safe_word: [],
        safe_word_match_policy: '0',
        mute_sender: '0',
        mute_duration: '0',
        ransom_amount: '0',
        reply_content: [],
      }
    }
    replyConfigIndex.value = index
    replyConfigModel.value = true
  }
  catch (error) {
    console.warn(error)
    replyConfigModel.value = false
  }
}

// 自动回复配置存储
function saveReplyConfig() {
  if (replyConfigForm.value.keyword.length === 0) {
    $message?.warning('关键词不可以为空')
    return false
  }
  if (replyConfigForm.value.reply_content.length === 0) {
    $message?.warning('回复内容不可以为空')
    return false
  }
  replyForm.value.content[replyConfigIndex.value] = { ...replyConfigForm.value }
  replyConfigModel.value = false
}

// 各模块配置存储
function apply(type) {
  switch (type) {
    case 'basic': // 基础设置
      if (roomConfigForm.value.is_listening.trim() === '') {
        return $message.warning('启动时自动监听不可以为空')
      }
      if (roomConfigForm.value.max_name_length.trim() === '') {
        return $message.warning('用户名最大长度不可以为空')
      }
      if (roomConfigForm.value.name_trim_mode.trim() === '') {
        return $message.warning('裁剪模式不可以为空')
      }
      if (roomConfigForm.value.consume_reward_enabled.trim() === '') {
        return $message.warning('消费奖励不可以为空')
      }
      if (roomConfigForm.value.reward_type.trim() === '') {
        return $message.warning('奖励类型不可以为空')
      }
      if (roomConfigForm.value.consume_reward_enabled === '1') {
        if (roomConfigForm.value.consume_battery_rate.trim() === '') {
          return $message.warning('转换倍率不可以为空')
        }
      }
      if (roomConfigForm.value.consume_reward_enabled === '2') {
        if (roomConfigForm.value.captain_reward_amount.trim() === '') {
          return $message.warning('舰长奖励不可以为空')
        }
        if (roomConfigForm.value.commander_reward_amount.trim() === '') {
          return $message.warning('提督奖励不可以为空')
        }
        if (roomConfigForm.value.governor_reward_amount.trim() === '') {
          return $message.warning('总督奖励不可以为空')
        }
      }
      roomLoading.value = true
      api.applyRoom(
        roomConfigForm.value.is_listening,
        roomConfigForm.value.max_name_length,
        roomConfigForm.value.name_trim_mode,
        roomConfigForm.value.consume_reward_enabled,
        roomConfigForm.value.reward_type,
        roomConfigForm.value.consume_battery_rate,
        roomConfigForm.value.captain_reward_amount,
        roomConfigForm.value.commander_reward_amount,
        roomConfigForm.value.governor_reward_amount,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        roomLoading.value = false
      })
      break
    case 'sign': // 签到配置
      if (signForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (signForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (signForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      if (signForm.value.reward_type.trim() === '') {
        return $message.warning('奖励类型不可以为空')
      }
      if (signForm.value.reward_amount.trim() === '') {
        return $message.warning('奖励数量不可以为空')
      }
      if (signForm.value.keyword.trim() === '') {
        return $message.warning('签到关键词不可以为空')
      }
      if (signForm.value.query_keyword.trim() === '') {
        return $message.warning('查询关键词不可以为空')
      }
      signForm.value.success_reply = signForm.value.success_reply.map(s => s.trim()).filter(s => s !== '')
      if (signForm.value.success_reply.length === 0) {
        return $message.warning('签到成功回复不可以为空')
      }
      signForm.value.fail_reply = signForm.value.fail_reply.map(s => s.trim()).filter(s => s !== '')
      if (signForm.value.fail_reply.length === 0) {
        return $message.warning('签到失败回复不可以为空')
      }
      signForm.value.repeat_reply = signForm.value.repeat_reply.map(s => s.trim()).filter(s => s !== '')
      if (signForm.value.repeat_reply.length === 0) {
        return $message.warning('重复签到回复不可以为空')
      }
      signForm.value.query_reply = signForm.value.query_reply.map(s => s.trim()).filter(s => s !== '')
      if (signForm.value.query_reply.length === 0) {
        return $message.warning('查询成功回复不可以为空')
      }
      signLoading.value = true
      api.applySign(
        signForm.value.enabled,
        signForm.value.scene,
        signForm.value.requirement,
        signForm.value.reward_type,
        signForm.value.reward_amount,
        signForm.value.keyword,
        signForm.value.query_keyword,
        signForm.value.success_reply,
        signForm.value.fail_reply,
        signForm.value.repeat_reply,
        signForm.value.query_reply,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        signLoading.value = false
      })
      break
    case 'ad': // 定时广告
      if (adForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (adForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (adForm.value.interval.trim() === '') {
        return $message.warning('发送间隔不可以为空')
      }
      if (adForm.value.send_mode.trim() === '') {
        return $message.warning('发送模式不可以为空')
      }
      adForm.value.content = adForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (adForm.value.content.length === 0) {
        return $message.warning('广告内容不可以为空')
      }
      adLoading.value = true
      api.applyAd(
        adForm.value.enabled,
        adForm.value.scene,
        adForm.value.interval,
        adForm.value.send_mode,
        adForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        adLoading.value = false
      })
      break
    case 'gift': // 礼物答谢
      if (giftForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (giftForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (giftForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      if (giftForm.value.show_count.trim() === '') {
        return $message.warning('显示计数不可以为空')
      }
      if (giftForm.value.merge_gift.trim() === '') {
        return $message.warning('合并礼物不可以为空')
      }
      if (giftForm.value.include_blindbox.trim() === '') {
        return $message.warning('包含盲盒不可以为空')
      }
      if (giftForm.value.min_battery.trim() === '') {
        return $message.warning('最低电池数不可以为空')
      }
      giftForm.value.content = giftForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (giftForm.value.content.length === 0) {
        return $message.warning('答谢内容不可以为空')
      }
      giftLoading.value = true
      api.applyGift(
        giftForm.value.enabled,
        giftForm.value.scene,
        giftForm.value.requirement,
        giftForm.value.show_count,
        giftForm.value.merge_gift,
        giftForm.value.include_blindbox,
        giftForm.value.min_battery,
        giftForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        giftLoading.value = false
      })
      break
    case 'pk': // PK播报
      if (pkForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      pkForm.value.content = pkForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (pkForm.value.content.length === 0) {
        return $message.warning('播报内容不可以为空')
      }
      pkLoading.value = true
      api.applyPk(
        pkForm.value.enabled,
        pkForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        pkLoading.value = false
      })
      break
    case 'welcome': // 进房欢迎
      if (welcomeForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (welcomeForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (welcomeForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      welcomeForm.value.content = welcomeForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (welcomeForm.value.content.length === 0) {
        return $message.warning('欢迎内容不可以为空')
      }
      welcomeLoading.value = true
      api.applyWelcome(
        welcomeForm.value.enabled,
        welcomeForm.value.scene,
        welcomeForm.value.requirement,
        welcomeForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        welcomeLoading.value = false
      })
      break
    case 'follow': // 感谢关注
      if (followForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (followForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (followForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      followForm.value.content = followForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (followForm.value.content.length === 0) {
        return $message.warning('感谢内容不可以为空')
      }
      followLoading.value = true
      api.applyFollow(
        followForm.value.enabled,
        followForm.value.scene,
        followForm.value.requirement,
        followForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        followLoading.value = false
      })
      break
    case 'share': // 感谢分享
      if (shareForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (shareForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (shareForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      shareForm.value.content = shareForm.value.content.map(s => s.trim()).filter(s => s !== '')
      if (shareForm.value.content.length === 0) {
        return $message.warning('感谢内容不可以为空')
      }
      shareLoading.value = true
      api.applyShare(
        shareForm.value.enabled,
        shareForm.value.scene,
        shareForm.value.requirement,
        shareForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        shareLoading.value = false
      })
      break
    case 'reply': // 自动回复
      if (replyForm.value.enabled.trim() === '') {
        return $message.warning('是否开启不可以为空')
      }
      if (replyForm.value.scene.trim() === '') {
        return $message.warning('可用场景不可以为空')
      }
      if (replyForm.value.requirement.trim() === '') {
        return $message.warning('触发门槛不可以为空')
      }
      replyLoading.value = true
      api.applyReply(
        replyForm.value.enabled,
        replyForm.value.scene,
        replyForm.value.requirement,
        replyForm.value.content,
      ).then(() => {
        $message.success('保存成功')
      }).finally(() => {
        replyLoading.value = false
      })
      break
  }
}

function scrollTo(id) {
  document.getElementById(id)?.scrollIntoView({
    behavior: 'smooth',
  })
}

onMounted(() => {
  api.getRoom().then((res) => {
    Object.assign(roomConfigForm.value, res.data)
    roomConfigShow.value = true
  }).catch(() => {
    roomConfigShow.value = false
  })
  api.getSign().then((res) => {
    Object.assign(signForm.value, res.data)
    signShow.value = true
  }).catch(() => {
    signShow.value = false
  })
  api.getAd().then((res) => {
    Object.assign(adForm.value, res.data)
    adShow.value = true
  }).catch(() => {
    adShow.value = false
  })
  api.getGift().then((res) => {
    Object.assign(giftForm.value, res.data)
    giftShow.value = true
  }).catch(() => {
    giftShow.value = false
  })
  api.getPk().then((res) => {
    Object.assign(pkForm.value, res.data)
    pkShow.value = true
  }).catch(() => {
    pkShow.value = false
  })
  api.getWelcome().then((res) => {
    Object.assign(welcomeForm.value, res.data)
    welcomeShow.value = true
  }).catch(() => {
    welcomeShow.value = false
  })
  api.getFollow().then((res) => {
    Object.assign(followForm.value, res.data)
    followShow.value = true
  }).catch(() => {
    followShow.value = false
  })
  api.getShare().then((res) => {
    Object.assign(shareForm.value, res.data)
    shareShow.value = true
  }).catch(() => {
    shareShow.value = false
  })
  api.getReply().then((res) => {
    Object.assign(replyForm.value, res.data)
    replyShow.value = true
  }).catch(() => {
    replyShow.value = false
  })
})
</script>
