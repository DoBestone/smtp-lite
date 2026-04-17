<template>
  <div class="docs-page">
    <div class="docs-page__toolbar">
      <p class="docs-page__sub">{{ t('docs.subtitle') }}</p>
      <el-button :icon="DocumentCopy" @click="copyMarkdown">
        {{ copyMdLabel }}
      </el-button>
    </div>

    <div class="docs-page__layout">
      <!-- 左侧目录 -->
      <aside class="docs-toc">
        <div class="docs-toc__label">目录</div>
        <ul>
          <li v-for="s in sections" :key="s.id">
            <a :href="`#${s.id}`" @click.prevent="scrollTo(s.id)">{{ s.label }}</a>
          </li>
        </ul>
        <div class="docs-toc__foot">
          <div class="docs-toc__label">当前域名</div>
          <code class="docs-toc__origin mono">{{ origin }}</code>
        </div>
      </aside>

      <!-- 主体 -->
      <div class="docs-body">
        <!-- 认证方式 -->
        <section id="auth" class="doc-sec">
          <h2>{{ t('docs.authTitle') }}</h2>
          <p class="doc-sec__desc">{{ t('docs.authDesc') }}</p>

          <div class="doc-grid">
            <div class="doc-auth">
              <div class="doc-auth__head">
                <span class="doc-auth__tag">推荐</span>
                <h3>{{ t('docs.authKey') }}</h3>
              </div>
              <p>{{ t('docs.authKeyDesc') }}</p>
              <pre class="doc-code" v-text="'X-API-Key: sk_xxxxxxxxxxxxxxxxxxxx'"></pre>
            </div>
            <div class="doc-auth">
              <div class="doc-auth__head">
                <h3>{{ t('docs.authJwt') }}</h3>
              </div>
              <p>{{ t('docs.authJwtDesc') }}</p>
              <pre class="doc-code" v-text="'Authorization: Bearer <jwt-token>'"></pre>
            </div>
          </div>
        </section>

        <!-- 单发 -->
        <section id="send" class="doc-sec">
          <h2>{{ t('docs.sendTitle') }}</h2>
          <EndpointRow method="POST" :path="`${origin}/api/v1/send`" />
          <CodeBlock :code="curlSend" />
        </section>

        <!-- 批量 -->
        <section id="batch" class="doc-sec">
          <h2>{{ t('docs.batchTitle') }}</h2>
          <EndpointRow method="POST" :path="`${origin}/api/v1/send/batch`" />
          <CodeBlock :code="curlBatch" />
        </section>

        <!-- 定时 -->
        <section id="scheduled" class="doc-sec">
          <h2>{{ t('docs.scheduledTitle') }}</h2>
          <EndpointRow method="POST" :path="`${origin}/api/v1/send/scheduled`" />
          <CodeBlock :code="curlScheduled" />
        </section>

        <!-- 代码示例 -->
        <section id="examples" class="doc-sec">
          <h2>{{ t('docs.examplesTitle') }}</h2>
          <div class="doc-lang-tabs">
            <button
              v-for="l in langs"
              :key="l.value"
              class="doc-lang"
              :class="{ 'is-active': lang === l.value }"
              @click="lang = l.value"
            >
              {{ l.label }}
            </button>
          </div>
          <CodeBlock :code="examples[lang]" />
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { DocumentCopy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import CodeBlock from './components/CodeBlock.vue'
import EndpointRow from './components/EndpointRow.vue'

const { t } = useI18n()

type Lang = 'curl' | 'python' | 'nodejs' | 'go' | 'php'

const lang = ref<Lang>('curl')
const origin = ref(location.origin || 'https://your-domain.com')

const sections = [
  { id: 'auth', label: '认证方式' },
  { id: 'send', label: '发送邮件' },
  { id: 'batch', label: '批量发送' },
  { id: 'scheduled', label: '定时发送' },
  { id: 'examples', label: '代码示例' }
]

const langs: Array<{ value: Lang; label: string }> = [
  { value: 'curl', label: 'cURL' },
  { value: 'python', label: 'Python' },
  { value: 'nodejs', label: 'Node.js' },
  { value: 'go', label: 'Go' },
  { value: 'php', label: 'PHP' }
]

function scrollTo(id: string) {
  const el = document.getElementById(id)
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const curlSend = computed(
  () => `curl -X POST ${origin.value}/api/v1/send \\
  -H "X-API-Key: sk_xxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "to": "user@example.com",
    "subject": "Hello",
    "body": "邮件内容",
    "is_html": false,
    "from_name": "发件人名称",
    "track_enabled": false
  }'`
)

const curlBatch = computed(
  () => `curl -X POST ${origin.value}/api/v1/send/batch \\
  -H "X-API-Key: sk_xxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "emails": ["user1@example.com", "user2@example.com"],
    "subject": "系统通知",
    "body": "<p>内容</p>",
    "is_html": true,
    "track_enabled": true
  }'`
)

const curlScheduled = computed(
  () => `curl -X POST ${origin.value}/api/v1/send/scheduled \\
  -H "X-API-Key: sk_xxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "to": "user@example.com",
    "subject": "预约提醒",
    "body": "您的会议 30 分钟后开始",
    "scheduled_at": "2026-01-01T10:00:00Z"
  }'`
)

const examples = computed<Record<Lang, string>>(() => {
  const base = origin.value
  return {
    curl: curlSend.value,
    python: `import requests

resp = requests.post(
    "${base}/api/v1/send",
    headers={"X-API-Key": "sk_xxxxxxxxxx"},
    json={
        "to": "user@example.com",
        "subject": "Hello",
        "body": "邮件内容",
    },
)
print(resp.json())`,
    nodejs: `import axios from 'axios'

const { data } = await axios.post(
  '${base}/api/v1/send',
  {
    to: 'user@example.com',
    subject: 'Hello',
    body: '邮件内容',
  },
  {
    headers: { 'X-API-Key': 'sk_xxxxxxxxxx' },
  },
)
console.log(data)`,
    go: `package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    body, _ := json.Marshal(map[string]any{
        "to":      "user@example.com",
        "subject": "Hello",
        "body":    "邮件内容",
    })
    req, _ := http.NewRequest("POST", "${base}/api/v1/send", bytes.NewReader(body))
    req.Header.Set("X-API-Key", "sk_xxxxxxxxxx")
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println(resp.Status)
}`,
    php: `<?php
$ch = curl_init('${base}/api/v1/send');
curl_setopt_array($ch, [
    CURLOPT_POST => true,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER => [
        'X-API-Key: sk_xxxxxxxxxx',
        'Content-Type: application/json',
    ],
    CURLOPT_POSTFIELDS => json_encode([
        'to' => 'user@example.com',
        'subject' => 'Hello',
        'body' => '邮件内容',
    ]),
]);
$resp = curl_exec($ch);
curl_close($ch);
echo $resp;`
  }
})

const copyMdLabel = ref(t('docs.copyMd'))

function buildMarkdown(): string {
  return `# SMTP Lite API

## 认证方式

- X-API-Key: \`sk_xxx\`（推荐，用于发信集成）
- Authorization: \`Bearer <jwt>\`（后台管理接口）

## 发送邮件

\`\`\`bash
${curlSend.value}
\`\`\`

## 批量发送

\`\`\`bash
${curlBatch.value}
\`\`\`

## 定时发送

\`\`\`bash
${curlScheduled.value}
\`\`\`

## Python

\`\`\`python
${examples.value.python}
\`\`\`
`
}

async function copyMarkdown() {
  try {
    await navigator.clipboard.writeText(buildMarkdown())
    ElMessage.success(t('common.success'))
  } catch {
    ElMessage.error(t('common.failed'))
  }
}
</script>

<style scoped>
.docs-page { display: flex; flex-direction: column; gap: 16px; }

.docs-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.docs-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }

.docs-page__layout {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

/* ---------- 目录 ---------- */
.docs-toc {
  position: sticky;
  top: 76px;
  padding: 14px 14px;
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: var(--shadow-xs);
}

.docs-toc__label {
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-placeholder);
  margin-bottom: 8px;
}

.docs-toc ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.docs-toc a {
  display: block;
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 12.5px;
  color: var(--color-text-regular);
  transition: background var(--transition-fast), color var(--transition-fast);
}

.docs-toc a:hover {
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
}

.docs-toc__foot {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px dashed var(--color-border);
}

.docs-toc__origin {
  display: block;
  padding: 6px 8px;
  background: var(--color-bg-subtle);
  border-radius: 6px;
  font-size: 11px;
  color: var(--color-text-regular);
  word-break: break-all;
}

/* ---------- 主体 ---------- */
.docs-body {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.doc-sec {
  background: #fff;
  padding: 20px 22px;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: var(--shadow-xs);
  scroll-margin-top: 76px;
}

.doc-sec h2 {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 10px;
  letter-spacing: -0.01em;
}

.doc-sec__desc {
  font-size: 12.5px;
  color: var(--color-text-secondary);
  margin-bottom: 14px;
}

/* Auth 双卡 */
.doc-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.doc-auth {
  padding: 14px 16px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-bg-subtle);
}

.doc-auth__head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.doc-auth__head h3 {
  font-family: var(--font-display);
  font-size: 13.5px;
  font-weight: 600;
}

.doc-auth__tag {
  display: inline-block;
  padding: 1px 8px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
}

.doc-auth p {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-bottom: 10px;
  line-height: 1.55;
}

.doc-code {
  margin: 0;
  padding: 8px 10px;
  background: #0f172a;
  border-radius: 6px;
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: #7dd3fc;
  overflow: auto;
  white-space: pre;
}

/* 语言 tabs */
.doc-lang-tabs {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--color-bg-subtle);
  border-radius: 8px;
  margin-bottom: 12px;
  width: fit-content;
}

.doc-lang {
  padding: 5px 12px;
  background: transparent;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.doc-lang:hover { color: var(--color-primary); }
.doc-lang.is-active {
  background: #fff;
  color: var(--color-primary);
  box-shadow: var(--shadow-xs);
}

/* ---------- 响应式 ---------- */
@media (max-width: 992px) {
  .docs-page__layout { grid-template-columns: 1fr; }
  .docs-toc { position: static; }
  .doc-grid { grid-template-columns: 1fr; }
}
</style>
