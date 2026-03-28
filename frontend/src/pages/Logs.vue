<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>发送日志</h2>
        <p class="page-desc">查看所有邮件发送记录与状态详情</p>
      </div>
      <button class="btn-secondary" @click="loadLogs">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2"/></svg>
        刷新
      </button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>收件人</th>
            <th>主题</th>
            <th>状态</th>
            <th>追踪</th>
            <th>发送时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td class="cell-main">{{ log.to_email }}</td>
            <td>{{ log.subject || '-' }}</td>
            <td>
              <span :class="['badge', log.status]">{{ log.status === 'success' ? '成功' : '失败' }}</span>
            </td>
            <td>
              <span v-if="log.opened" class="track-badge opened">已打开</span>
              <span v-if="log.clicked" class="track-badge clicked">已点击</span>
              <span v-if="!log.opened && !log.clicked" class="text-muted">-</span>
            </td>
            <td>{{ formatDate(log.created_at) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="5" class="empty-cell">
              <div class="empty-state">暂无发送记录</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { store, actions } from '@/store'

export default {
  name: 'Logs',
  setup() {
    const logs = computed(() => store.logs)
    
    const loadLogs = () => actions.loadLogs()
    
    const formatDate = (date) => {
      if (!date) return '-'
      return new Date(date).toLocaleString('zh-CN')
    }
    
    onMounted(loadLogs)
    
    return { logs, loadLogs, formatDate }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.track-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  margin-right: 4px;
}

.track-badge.opened { background: #fef3c7; color: #92400e; }
.track-badge.clicked { background: #dbeafe; color: #1e40af; }

.text-muted { color: #94a3b8; }
</style>