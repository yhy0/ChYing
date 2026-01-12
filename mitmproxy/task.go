package mitmproxy

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/task"
)

/**
  @author: yhy
  @since: 2023/10/10
  @desc: 判断是否扫描过，不能在这里进行，防止一开始某些插件没开，运行中开启导致无法扫描，判断是否扫描过的逻辑放到每个插件内部中
  @update: 2025/7/10 - 迁移到 proxify 基于插件的架构
**/

var passiveTask *task.Task

// NewPassiveTask 初始化被动扫描任务
// 此函数保持原有的接口，但内部实现改为使用新的插件架构
func NewPassiveTask(_t *task.Task) {
	passiveTask = _t

	defer passiveTask.Pool.Release() // 释放协程池

	// 先加一，这里会一直阻塞，这样就不会马上退出, 这里要的就是一直阻塞，所以不使用 wg.Done()
	passiveTask.WG.Add(1)

	passiveTask.WG.Wait()
}
