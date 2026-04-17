<template>
  <div class="stats">
    <!-- 顶部工具条 —— 裸放在背景上，不开单独卡片 -->
    <div class="stats__toolbar">
      <div class="stats__sub">
        <span class="stats__dot" :class="{ 'is-live': !loading }"></span>
        <span>{{ t('stats.subtitle') }}</span>
        <span class="stats__sep">·</span>
        <span class="stats__updated">{{ t('common.refresh') }} {{ updatedLabel }}</span>
      </div>
      <el-button :icon="Refresh" @click="refresh" :loading="loading">
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <!-- KPI 网格 · 5 + 2 布局 -->
    <div class="stats__grid">
      <StatCard
        :title="t('stats.totalSent')"
        :value="stats.total_sent"
        :icon="Message"
        tone="primary"
        :delay="0"
      />
      <StatCard
        :title="t('stats.success')"
        :value="stats.success"
        :icon="CircleCheck"
        tone="success"
        :delay="60"
      />
      <StatCard
        :title="t('stats.failed')"
        :value="stats.failed"
        :icon="CircleClose"
        tone="danger"
        :delay="120"
      />
      <StatCard
        :title="t('stats.todaySent')"
        :value="stats.today_sent"
        :icon="Timer"
        tone="info"
        :delay="180"
      />
      <StatCard
        :title="t('stats.successRate')"
        :value="stats.success_rate"
        suffix="%"
        :decimals="1"
        :icon="DataLine"
        tone="primary"
        :delay="240"
      />
      <StatCard
        :title="t('stats.opened')"
        :value="stats.opened"
        :sub="openSub"
        :icon="View"
        tone="warning"
        :delay="300"
      />
      <StatCard
        :title="t('stats.clicked')"
        :value="stats.clicked"
        :sub="clickSub"
        :icon="Pointer"
        tone="accent"
        :delay="360"
      />
    </div>

    <!-- 队列状态 -->
    <section class="stats__panel">
      <header class="stats__panel-head">
        <h3>{{ t('stats.queueTitle') }}</h3>
        <span class="stats__panel-meta">
          {{ queueTotal }}
          <span class="stats__panel-meta-unit">total</span>
        </span>
      </header>

      <div class="stats__queue">
        <QueueCell :label="t('stats.queuePending')" :value="queue.pending" tone="info" />
        <QueueCell :label="t('stats.queueProcessing')" :value="queue.processing" tone="primary" animate />
        <QueueCell :label="t('stats.queueSent')" :value="queue.sent" tone="success" />
        <QueueCell :label="t('stats.queueFailed')" :value="queue.failed" tone="danger" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CircleCheck,
  CircleClose,
  DataLine,
  Message,
  Pointer,
  Refresh,
  Timer,
  View
} from '@element-plus/icons-vue'
import { statsApi } from '@/api'
import type { QueueStats, Stats } from '@/api/types'
import StatCard from './components/StatCard.vue'
import QueueCell from './components/QueueCell.vue'

const { t } = useI18n()

const stats = ref<Partial<Stats>>({})
const queue = ref<Partial<QueueStats>>({})
const loading = ref(false)
const updatedAt = ref<Date | null>(null)

const queueTotal = computed(() =>
  ((queue.value.pending ?? 0) +
    (queue.value.processing ?? 0) +
    (queue.value.sent ?? 0) +
    (queue.value.failed ?? 0)).toLocaleString()
)

const openSub = computed(() =>
  stats.value.open_rate != null ? `${stats.value.open_rate.toFixed(1)}%` : ''
)
const clickSub = computed(() =>
  stats.value.click_rate != null ? `${stats.value.click_rate.toFixed(1)}%` : ''
)

const updatedLabel = computed(() => {
  if (!updatedAt.value) return '—'
  const d = updatedAt.value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
})

async function refresh() {
  loading.value = true
  try {
    const [s, q] = await Promise.all([statsApi.get(), statsApi.queue()])
    stats.value = s ?? {}
    queue.value = q ?? {}
    updatedAt.value = new Date()
  } finally {
    loading.value = false
  }
}

let timer: number | null = null

onMounted(() => {
  refresh()
  timer = window.setInterval(refresh, 30_000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<style scoped>
.stats {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stats__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.stats__sub {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--color-text-secondary);
}

.stats__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--color-text-placeholder);
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.15);
}

.stats__dot.is-live {
  background: var(--color-success);
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.18);
  animation: pulse 2.2s ease-out infinite;
}

@keyframes pulse {
  0%   { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.4); }
  70%  { box-shadow: 0 0 0 7px rgba(34, 197, 94, 0); }
  100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
}

.stats__sep { color: var(--color-text-placeholder); }
.stats__updated { font-family: var(--font-mono); font-size: 11.5px; }

/* ---------- KPI 网格 ---------- */
.stats__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 14px;
}

/* ---------- 队列面板 ---------- */
.stats__panel {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 18px 20px;
  box-shadow: var(--shadow-xs);
}

.stats__panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-bottom: 14px;
  margin-bottom: 14px;
  border-bottom: 1px dashed var(--color-border);
}

.stats__panel-head h3 {
  font-family: var(--font-display);
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.stats__panel-meta {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--color-text-regular);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.stats__panel-meta-unit {
  margin-left: 4px;
  font-size: 11px;
  font-weight: 400;
  color: var(--color-text-placeholder);
  letter-spacing: 0.04em;
}

.stats__queue {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
}

@media (max-width: 576px) {
  .stats__toolbar { flex-direction: column; align-items: flex-start; gap: 8px; }
  .stats__sub { flex-wrap: wrap; }
}
</style>
