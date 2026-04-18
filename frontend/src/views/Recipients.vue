<template>
  <div class="rcp-page">
    <div class="rcp-page__toolbar">
      <p class="rcp-page__sub">{{ t('recipients.subtitle') }}</p>
      <el-button :icon="Refresh" @click="loadGroups" :loading="loadingGroups" />
    </div>

    <div class="rcp-page__grid">
      <!-- 左栏：分组列表 -->
      <el-card shadow="never" class="rcp-page__side">
        <div class="rcp-side__head">
          <h3>{{ t('recipients.groupList') }}</h3>
          <el-button size="small" type="primary" :icon="Plus" @click="openGroupDialog()">
            {{ t('recipients.addGroup') }}
          </el-button>
        </div>

        <div v-if="groups.length" class="rcp-side__list">
          <div
            v-for="g in groups"
            :key="g.id"
            class="rcp-group"
            :class="{ 'is-active': activeGroupId === g.id }"
            role="button"
            tabindex="0"
            @click="selectGroup(g.id)"
            @keydown.enter.space.prevent="selectGroup(g.id)"
          >
            <div class="rcp-group__top">
              <span class="rcp-group__name">{{ g.name }}</span>
              <span class="rcp-group__count">{{ g.count ?? 0 }}</span>
            </div>
            <div v-if="g.description" class="rcp-group__desc">{{ g.description }}</div>
            <div class="rcp-group__ops">
              <button
                type="button"
                class="rcp-group__op"
                :title="t('common.edit')"
                @click.stop="openGroupDialog(g)"
              >
                <el-icon><Edit /></el-icon>
              </button>
              <button
                type="button"
                class="rcp-group__op rcp-group__op--danger"
                :title="t('common.delete')"
                @click.stop="onDeleteGroup(g)"
              >
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </div>
        </div>
        <EmptyHint v-else :icon="Folder" :text="t('recipients.empty')" />
      </el-card>

      <!-- 右栏：成员列表 -->
      <el-card shadow="never" class="rcp-page__main">
        <template v-if="activeGroup">
          <div class="rcp-main__head">
            <div class="rcp-main__title">
              <h3>{{ activeGroup.name }}</h3>
              <span class="rcp-main__count">
                {{ recipients.length }} {{ t('recipients.memberCount') }}
              </span>
            </div>
            <div class="rcp-main__actions">
              <el-button :icon="Upload" @click="importVisible = true">
                {{ t('recipients.batchImport') }}
              </el-button>
              <el-button type="primary" :icon="Plus" @click="openRecipientDialog()">
                {{ t('recipients.addRecipient') }}
              </el-button>
            </div>
          </div>

          <el-table
            v-loading="loadingRecipients"
            :data="recipients"
            class="rcp-table"
            row-key="id"
            max-height="600"
          >
            <el-table-column label="邮箱" prop="email" min-width="220">
              <template #default="{ row }">
                <span class="mono rcp-table__email">{{ row.email }}</span>
              </template>
            </el-table-column>
            <el-table-column label="名称" prop="name" min-width="160">
              <template #default="{ row }">
                {{ row.name || '—' }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.status')" width="110">
              <template #default="{ row }">
                <StatusPill :kind="row.status === 'active' ? 'success' : 'muted'">
                  {{ row.status === 'active' ? '正常' : '黑名单' }}
                </StatusPill>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.createdAt')" width="170">
              <template #default="{ row }">
                <span class="mono rcp-table__meta">{{ formatDateTime(row.created_at) }}</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.operations')" width="100" align="right" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteRecipient(row)">
                  {{ t('common.remove') }}
                </el-button>
              </template>
            </el-table-column>
            <template #empty>
              <EmptyHint :icon="User" :text="t('recipients.emptyMembers')" />
            </template>
          </el-table>
        </template>

        <div v-else class="rcp-main__placeholder">
          <EmptyHint :icon="Folder" :text="t('recipients.selectGroup')" />
        </div>
      </el-card>
    </div>

    <!-- 分组弹窗 -->
    <el-dialog
      v-model="groupDialogVisible"
      :title="editingGroup ? '编辑分组' : t('recipients.addGroup')"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('recipients.groupName')">
          <el-input v-model="groupForm.name" placeholder="如：客户名单 / 订阅用户" />
        </el-form-item>
        <el-form-item :label="t('recipients.groupDescription')">
          <el-input v-model="groupForm.description" placeholder="可选备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submittingGroup" @click="submitGroup">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加收件人弹窗 -->
    <el-dialog
      v-model="recipientDialogVisible"
      :title="t('recipients.addRecipient')"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item label="邮箱">
          <el-input v-model="recipientForm.email" type="email" placeholder="user@example.com" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="recipientForm.name" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recipientDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submittingRecipient" @click="submitRecipient">
          {{ t('common.add') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量导入弹窗 -->
    <el-dialog
      v-model="importVisible"
      :title="t('recipients.batchImport')"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <p class="rcp-import__desc">{{ t('recipients.batchImportDesc') }}</p>
      <el-input
        v-model="importText"
        type="textarea"
        :rows="10"
        placeholder="user1@example.com&#10;user2@example.com,张三&#10;user3@example.com"
        resize="vertical"
      />
      <template #footer>
        <el-button @click="importVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submittingImport" @click="submitImport">
          {{ t('common.import') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Delete,
  Edit,
  Folder,
  Plus,
  Refresh,
  Upload,
  User
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { recipientsApi } from '@/api'
import type { Recipient, RecipientGroup } from '@/api/types'
import StatusPill from './components/StatusPill.vue'
import EmptyHint from './components/EmptyHint.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

// ---------- 分组 ----------
const groups = ref<RecipientGroup[]>([])
const loadingGroups = ref(false)
const activeGroupId = ref<string>('')

const activeGroup = computed(() =>
  groups.value.find((g) => g.id === activeGroupId.value) ?? null
)

async function loadGroups() {
  loadingGroups.value = true
  try {
    groups.value = await recipientsApi.listGroups()
    if (!activeGroupId.value && groups.value.length) {
      selectGroup(groups.value[0].id)
    } else if (activeGroupId.value) {
      await loadRecipients()
    }
  } finally {
    loadingGroups.value = false
  }
}

function selectGroup(id: string) {
  activeGroupId.value = id
}

watch(activeGroupId, () => {
  loadRecipients()
})

// ---------- 成员 ----------
const recipients = ref<Recipient[]>([])
const loadingRecipients = ref(false)

async function loadRecipients() {
  if (!activeGroupId.value) {
    recipients.value = []
    return
  }
  loadingRecipients.value = true
  try {
    recipients.value = await recipientsApi.listByGroup(activeGroupId.value)
  } finally {
    loadingRecipients.value = false
  }
}

// ---------- 分组弹窗 ----------
const groupDialogVisible = ref(false)
const editingGroup = ref<RecipientGroup | null>(null)
const submittingGroup = ref(false)
const groupForm = reactive({ name: '', description: '' })

function openGroupDialog(group?: RecipientGroup) {
  editingGroup.value = group ?? null
  groupForm.name = group?.name ?? ''
  groupForm.description = group?.description ?? ''
  groupDialogVisible.value = true
}

async function submitGroup() {
  if (!groupForm.name.trim()) {
    ElMessage.error('请输入分组名称')
    return
  }
  submittingGroup.value = true
  try {
    if (editingGroup.value) {
      await recipientsApi.updateGroup(editingGroup.value.id, { ...groupForm })
    } else {
      await recipientsApi.createGroup({ ...groupForm })
    }
    ElMessage.success(t('common.success'))
    groupDialogVisible.value = false
    await loadGroups()
  } finally {
    submittingGroup.value = false
  }
}

async function onDeleteGroup(group: RecipientGroup) {
  const confirmed = await ElMessageBox.confirm(t('recipients.deleteGroupConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await recipientsApi.removeGroup(group.id)
  if (activeGroupId.value === group.id) activeGroupId.value = ''
  await loadGroups()
}

// ---------- 添加收件人弹窗 ----------
const recipientDialogVisible = ref(false)
const submittingRecipient = ref(false)
const recipientForm = reactive({ email: '', name: '' })

function openRecipientDialog() {
  recipientForm.email = ''
  recipientForm.name = ''
  recipientDialogVisible.value = true
}

async function submitRecipient() {
  if (!recipientForm.email.trim()) {
    ElMessage.error('请输入邮箱')
    return
  }
  if (!activeGroupId.value) return
  submittingRecipient.value = true
  try {
    await recipientsApi.create({
      group_id: activeGroupId.value,
      email: recipientForm.email.trim(),
      name: recipientForm.name.trim() || undefined
    })
    ElMessage.success(t('common.success'))
    recipientDialogVisible.value = false
    await Promise.all([loadRecipients(), loadGroups()])
  } finally {
    submittingRecipient.value = false
  }
}

async function onDeleteRecipient(row: Recipient) {
  const confirmed = await ElMessageBox.confirm(t('recipients.deleteRecipientConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.remove'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await recipientsApi.remove(row.id)
  await Promise.all([loadRecipients(), loadGroups()])
}

// ---------- 批量导入 ----------
const importVisible = ref(false)
const importText = ref('')
const submittingImport = ref(false)

async function submitImport() {
  if (!importText.value.trim()) {
    ElMessage.error('请输入要导入的邮箱')
    return
  }
  if (!activeGroupId.value) return
  submittingImport.value = true
  try {
    const res = await recipientsApi.batchImport(activeGroupId.value, importText.value)
    ElMessage.success(
      t('recipients.batchImportHint', {
        imported: res.imported,
        skipped: res.skipped
      })
    )
    importVisible.value = false
    importText.value = ''
    await Promise.all([loadRecipients(), loadGroups()])
  } finally {
    submittingImport.value = false
  }
}

onMounted(loadGroups)
</script>

<style scoped>
.rcp-page { display: flex; flex-direction: column; gap: 16px; }

.rcp-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.rcp-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }

.rcp-page__grid {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.rcp-page__side,
.rcp-page__main {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}

.rcp-page__side :deep(.el-card__body) { padding: 16px; }
.rcp-page__main :deep(.el-card__body) { padding: 16px; }

/* ---------- 侧栏 ---------- */
.rcp-side__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--color-border);
  margin-bottom: 12px;
}
.rcp-side__head h3 {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.rcp-side__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 620px;
  overflow-y: auto;
}

.rcp-group {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.rcp-group:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
}

.rcp-group.is-active {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
  box-shadow: var(--shadow-primary);
}

.rcp-group__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.rcp-group__name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.rcp-group__count {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
  padding: 1px 7px;
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 4px;
}

.rcp-group.is-active .rcp-group__count {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.rcp-group__desc {
  font-size: 11px;
  color: var(--color-text-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rcp-group__ops {
  position: absolute;
  top: 8px;
  right: 8px;
  display: none;
  gap: 2px;
}

.rcp-group:hover .rcp-group__ops,
.rcp-group.is-active .rcp-group__ops { display: flex; }

.rcp-group__op {
  width: 22px;
  height: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  color: var(--color-text-secondary);
}

.rcp-group__op:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.rcp-group__op--danger:hover {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

/* ---------- 主栏 ---------- */
.rcp-main__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--color-border);
  margin-bottom: 12px;
}

.rcp-main__title { display: flex; align-items: baseline; gap: 10px; }
.rcp-main__title h3 {
  font-family: var(--font-display);
  font-size: 14px;
  font-weight: 600;
}

.rcp-main__count {
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--color-text-secondary);
}

.rcp-main__actions { display: flex; gap: 8px; }

.rcp-main__placeholder { padding: 24px 0; }

.rcp-table { font-size: 13px; }
.rcp-table__email { color: var(--color-text-primary); }
.rcp-table__meta { font-size: 11.5px; color: var(--color-text-secondary); }

.rcp-import__desc {
  margin-bottom: 10px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

@media (max-width: 992px) {
  .rcp-page__grid { grid-template-columns: 1fr; }
}
</style>
