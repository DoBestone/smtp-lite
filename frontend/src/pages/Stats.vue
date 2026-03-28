<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>统计信息</h2>
        <p class="page-desc">邮件发送情况概览与实时数据</p>
      </div>
      <button class="btn-secondary" @click="actions.loadStats()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2"/></svg>
        刷新
      </button>
    </div>
    
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon blue">📧</div>
        <div class="stat-value">{{ stats.total_sent || 0 }}</div>
        <div class="stat-label">累计发送</div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green">✓</div>
        <div class="stat-value">{{ stats.success || 0 }}</div>
        <div class="stat-label">发送成功</div>
      </div>
      <div class="stat-card">
        <div class="stat-icon red">✗</div>
        <div class="stat-value">{{ stats.failed || 0 }}</div>
        <div class="stat-label">发送失败</div>
      </div>
      <div class="stat-card">
        <div class="stat-icon cyan">🕐</div>
        <div class="stat-value">{{ stats.today_sent || 0 }}</div>
        <div class="stat-label">今日发送</div>
      </div>
      <div class="stat-card">
        <div class="stat-icon purple">%</div>
        <div class="stat-value">{{ (stats.success_rate || 0).toFixed(1) }}<span class="stat-unit">%</span></div>
        <div class="stat-label">成功率</div>
      </div>
      <div class="stat-card">
        <div class="stat-icon orange">👁</div>
        <div class="stat-value">{{ stats.opened || 0 }}</div>
        <div class="stat-label">已打开 <span v-if="stats.open_rate">({{ stats.open_rate.toFixed(1) }}%)</span></div>
      </div>
      <div class="stat-card">
        <div class="stat-icon teal">👆</div>
        <div class="stat-value">{{ stats.clicked || 0 }}</div>
        <div class="stat-label">已点击 <span v-if="stats.click_rate">({{ stats.click_rate.toFixed(1) }}%)</span></div>
      </div>
    </div>
    
    <!-- 队列状态 -->
    <div class="card">
      <div class="card-header">
        <h3>发送队列状态</h3>
      </div>
      <div class="queue-stats">
        <div class="queue-item">
          <span class="queue-label">待发送</span>
          <span class="queue-value pending">{{ queueStats.pending || 0 }}</span>
        </div>
        <div class="queue-item">
          <span class="queue-label">处理中</span>
          <span class="queue-value processing">{{ queueStats.processing || 0 }}</span>
        </div>
        <div class="queue-item">
          <span class="queue-label">已发送</span>
          <span class="queue-value success">{{ queueStats.sent || 0 }}</span>
        </div>
        <div class="queue-item">
          <span class="queue-label">发送失败</span>
          <span class="queue-value failed">{{ queueStats.failed || 0 }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, onMounted } from 'vue'
import { store, actions } from '@/store'

export default {
  name: 'Stats',
  setup() {
    const stats = computed(() => store.stats)
    const queueStats = computed(() => store.queueStats)
    
    onMounted(() => {
      actions.loadStats()
      actions.loadQueueStats()
    })
    
    return { stats, queueStats, actions }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.stat-card {
  position: relative;
}

.stat-icon {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}

.stat-icon.blue { background: #dbeafe; }
.stat-icon.green { background: #dcfce7; }
.stat-icon.red { background: #fee2e2; }
.stat-icon.cyan { background: #cffafe; }
.stat-icon.purple { background: #f3e8ff; }
.stat-icon.orange { background: #ffedd5; }
.stat-icon.teal { background: #ccfbf1; }

.stat-unit {
  font-size: 16px;
  margin-left: 2px;
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.queue-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 20px;
  padding: 20px;
}

.queue-item {
  text-align: center;
}

.queue-label {
  display: block;
  font-size: 13px;
  color: #64748b;
  margin-bottom: 8px;
}

.queue-value {
  font-size: 28px;
  font-weight: 700;
}

.queue-value.pending { color: #f59e0b; }
.queue-value.processing { color: #3b82f6; }
.queue-value.success { color: #22c55e; }
.queue-value.failed { color: #ef4444; }
</style>