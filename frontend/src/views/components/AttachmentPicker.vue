<template>
  <div class="ap">
    <div
      class="ap__drop"
      :class="{ 'is-drag': dragOver }"
      @click="input?.click()"
      @dragover.prevent="dragOver = true"
      @dragleave.prevent="dragOver = false"
      @drop.prevent="onDrop"
    >
      <el-icon :size="20"><UploadFilled /></el-icon>
      <span>{{ hintText }}</span>
      <input
        ref="input"
        type="file"
        multiple
        hidden
        @change="onPickFiles"
      />
    </div>

    <div v-if="modelValue.length" class="ap__list">
      <div v-for="(att, i) in modelValue" :key="i" class="ap__item">
        <el-icon><Document /></el-icon>
        <span class="ap__name">{{ att.filename }}</span>
        <span class="ap__size">{{ formatSize(att.size ?? 0) }}</span>
        <button class="ap__x" type="button" @click="remove(i)">
          <el-icon><Close /></el-icon>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Close, Document, UploadFilled } from '@element-plus/icons-vue'

interface Attachment {
  filename: string
  content: string
  size?: number
}

const props = defineProps<{ modelValue: Attachment[] }>()
const emit = defineEmits<{ 'update:modelValue': [Attachment[]] }>()

const { t } = useI18n()
const input = ref<HTMLInputElement | null>(null)
const dragOver = ref(false)

const hintText = t('send.attachmentDrop')

async function readAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      resolve(result.split(',')[1] ?? '')
    }
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

async function addFiles(files: FileList | File[]) {
  const out: Attachment[] = [...props.modelValue]
  for (const f of Array.from(files)) {
    const content = await readAsBase64(f)
    out.push({ filename: f.name, content, size: f.size })
  }
  emit('update:modelValue', out)
}

function onPickFiles(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) addFiles(target.files).finally(() => (target.value = ''))
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  if (e.dataTransfer?.files?.length) addFiles(e.dataTransfer.files)
}

function remove(i: number) {
  emit('update:modelValue', props.modelValue.filter((_, idx) => idx !== i))
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let n = bytes
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return `${n.toFixed(n < 10 ? 1 : 0)} ${units[i]}`
}
</script>

<style scoped>
.ap {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.ap__drop {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 16px;
  border: 1.5px dashed var(--color-border);
  border-radius: 10px;
  background: #fff;
  color: var(--color-text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.ap__drop:hover,
.ap__drop.is-drag {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
}

.ap__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ap__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: var(--color-bg-subtle);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 12px;
}

.ap__name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-primary);
}

.ap__size {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--color-text-placeholder);
  font-variant-numeric: tabular-nums;
}

.ap__x {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px;
  color: var(--color-text-placeholder);
  display: inline-flex;
  border-radius: 4px;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.ap__x:hover {
  color: var(--color-danger);
  background: var(--color-danger-bg);
}
</style>
