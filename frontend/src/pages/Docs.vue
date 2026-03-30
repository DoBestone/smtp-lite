<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>API 文档</h2>
        <p class="page-desc">完整的 API 接口文档与代码示例</p>
      </div>
      <button class="btn-copy-md" @click="copyMarkdown">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
          <rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="1.8"/>
          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="1.8"/>
        </svg>
        {{ copyBtnText }}
      </button>
    </div>
    
    <div class="docs-container">
      <!-- 基础认证 -->
      <section class="doc-section">
        <h3>认证方式</h3>
        <div class="doc-card">
          <p>支持两种认证方式：</p>
          <ul>
            <li><strong>JWT Token</strong> - 通过登录获取，用于管理操作</li>
            <li><strong>API Key</strong> - 在 API Key 页面创建，用于发送邮件</li>
          </ul>
        </div>
      </section>
      
      <!-- 发送邮件 -->
      <section class="doc-section">
        <h3>发送邮件</h3>
        <div class="doc-card">
          <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/v1/send</code>
          </div>
          <pre class="code-block">curl -X POST {{ baseUrl }}/api/v1/send \
  -H "X-API-Key: sk_xxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Hello",
    "body": "邮件内容",
    "is_html": false,
    "from_name": "发件人名称",
    "track_enabled": false
  }'</pre>
        </div>
      </section>
      
      <!-- 批量发送 -->
      <section class="doc-section">
        <h3>批量发送</h3>
        <div class="doc-card">
          <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/v1/send/batch</code>
          </div>
          <pre class="code-block">curl -X POST {{ baseUrl }}/api/v1/send/batch \
  -H "X-API-Key: sk_xxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "批量通知",
    "emails": ["user1@example.com", "user2@example.com"],
    "subject": "系统通知",
    "body": "内容",
    "is_html": true
  }'</pre>
        </div>
      </section>
      
      <!-- 代码示例 -->
      <section class="doc-section">
        <h3>代码示例</h3>
        <div class="tabs">
          <button v-for="t in ['python', 'nodejs', 'go']" :key="t" :class="['tab', { active: lang === t }]" @click="lang = t">{{ t.toUpperCase() }}</button>
        </div>
        <div class="doc-card">
          <pre class="code-block">{{ codeExamples[lang] }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'

export default {
  name: 'Docs',
  setup() {
    const lang = ref('python')
    const baseUrl = computed(() => window.location.origin)
    const copyBtnText = ref('复制 MD')
    
    const codeExamples = {
      python: `import requests

API_URL = "${window.location.origin}/api/v1/send"
API_KEY = "sk_xxxxxxxxxx"

def send_email(to, subject, body, is_html=False):
    resp = requests.post(API_URL,
        headers={"X-API-Key": API_KEY},
        json={"to": to, "subject": subject, "body": body, "is_html": is_html},
        timeout=30)
    return resp.json()

result = send_email("user@example.com", "验证码", "您的验证码是：123456")
print(result)`,
      nodejs: `const axios = require('axios');

const API_URL = '${window.location.origin}/api/v1/send';
const API_KEY = 'sk_xxxxxxxxxx';

async function sendEmail(to, subject, body, isHtml = false) {
  const resp = await axios.post(API_URL,
    { to, subject, body, is_html: isHtml },
    { headers: { 'X-API-Key': API_KEY } });
  return resp.data;
}

sendEmail('user@example.com', '验证码', '您的验证码是：123456')
  .then(console.log);`,
      go: `package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func sendEmail(to, subject, body string) {
    payload, _ := json.Marshal(map[string]interface{}{
        "to": to, "subject": subject, "body": body,
    })
    req, _ := http.NewRequest("POST", "${window.location.origin}/api/v1/send", bytes.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "sk_xxxxxxxxxx")
    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    fmt.Println(resp.Status)
}

func main() {
    sendEmail("user@example.com", "验证码", "您的验证码是：123456")
}`
    }
    
    const copyMarkdown = async () => {
      const url = baseUrl.value
      const md = `# SMTP Lite API 文档

## 认证方式

支持两种认证方式：

- **JWT Token** - 通过登录获取，用于管理操作
- **API Key** - 在 API Key 页面创建，用于发送邮件

## 发送邮件

\`POST\` \`/api/v1/send\`

\`\`\`bash
curl -X POST ${url}/api/v1/send \\
  -H "X-API-Key: sk_xxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "to": "user@example.com",
    "subject": "Hello",
    "body": "邮件内容",
    "is_html": false,
    "from_name": "发件人名称",
    "track_enabled": false
  }'
\`\`\`

## 批量发送

\`POST\` \`/api/v1/send/batch\`

\`\`\`bash
curl -X POST ${url}/api/v1/send/batch \\
  -H "X-API-Key: sk_xxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "批量通知",
    "emails": ["user1@example.com", "user2@example.com"],
    "subject": "系统通知",
    "body": "内容",
    "is_html": true
  }'
\`\`\`

## 代码示例

### Python

\`\`\`python
${codeExamples.python}
\`\`\`

### Node.js

\`\`\`javascript
${codeExamples.nodejs}
\`\`\`

### Go

\`\`\`go
${codeExamples.go}
\`\`\`
`
      try {
        await navigator.clipboard.writeText(md)
        copyBtnText.value = '已复制 ✓'
        setTimeout(() => { copyBtnText.value = '复制 MD' }, 2000)
      } catch {
        copyBtnText.value = '复制失败'
        setTimeout(() => { copyBtnText.value = '复制 MD' }, 2000)
      }
    }

    return { lang, baseUrl, codeExamples, copyBtnText, copyMarkdown }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

.btn-copy-md {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn-copy-md:hover {
  background: #2563eb;
}

.docs-container { max-width: 900px; }

.doc-section { margin-bottom: 32px; }
.doc-section h3 { font-size: 18px; margin: 0 0 16px 0; color: #1e293b; }

.doc-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.doc-card ul { margin: 8px 0; padding-left: 20px; }
.doc-card li { margin: 8px 0; color: #475569; }

.endpoint { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }

.method {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  font-family: monospace;
}
.method.post { background: #dcfce7; color: #166534; }
.method.get { background: #dbeafe; color: #1e40af; }

code {
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
  color: #475569;
}

.code-block {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre;
}

.tabs { display: flex; gap: 8px; margin-bottom: 16px; }

.tab {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}

.tab.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}
</style>