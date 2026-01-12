import type { AttackResult, IntruderTab, IntruderResult } from '../../../types';

// 攻击控制相关的组合式API
export function useIntruderAttackController() {
  // Start an attack
  const startAttack = (tab: IntruderTab) => {
    if (!tab) return;

    tab.isRunning = true;
    tab.progress = {
      total: calculateTotalRequests(tab),
      current: 0,
      startTime: Date.now(),
      endTime: null,
    };

    // 清空之前的结果
    tab.results = [];

    // 实际攻击由后端处理，前端通过事件接收结果
    // 后端会通过 Attack-Data 和动态 UUID 事件发送攻击结果
    
    return tab;
  };

  // Stop an attack
  const stopAttack = (tab: IntruderTab) => {
    if (!tab) return;

    tab.isRunning = false;
    tab.progress.endTime = Date.now();
    
    return tab;
  };

  // Calculate total requests for an attack
  const calculateTotalRequests = (tab: IntruderTab): number => {
    // This calculation would depend on the attack type and payload sets
    // For simplicity, we'll just return a number
    switch (tab.attackType) {
      case 'sniper':
        // Number of positions × number of payloads in set 1
        return (tab.payloadPositions.length || 1) * (tab.payloadSets[0]?.items.length || 10);
      case 'battering-ram':
        // Number of payloads in set 1
        return tab.payloadSets[0]?.items.length || 10;
      case 'pitchfork':
        // Number of payloads in smallest set
        return Math.min(...tab.payloadSets.map(set => set.items.length || 10));
      case 'cluster-bomb':
        // Product of the sizes of all payload sets
        return tab.payloadSets.reduce((product, set) => product * (set.items.length || 10), 1);
      default:
        return 10;
    }
  };

  // 添加攻击结果（由后端事件触发调用）
  const addAttackResult = (tab: IntruderTab, result: IntruderResult) => {
    if (!tab) return;
    
    // 更新进度
    tab.progress.current++;
    
    // 使用数组方法触发响应式更新
    tab.results = [...tab.results, result];

    // 检查攻击是否完成
    if (tab.progress.current >= tab.progress.total && tab.isRunning) {
      tab.isRunning = false;
      tab.progress.endTime = Date.now();
    }
    
    return result;
  };

  // 攻击类型选项
  const attackTypes = [
    { value: 'sniper', label: 'Sniper' },
    { value: 'battering-ram', label: 'Battering Ram' },
    { value: 'pitchfork', label: 'Pitchfork' },
    { value: 'cluster-bomb', label: 'Cluster Bomb' }
  ];

  return {
    startAttack,
    stopAttack,
    calculateTotalRequests,
    addAttackResult,
    attackTypes
  };
} 