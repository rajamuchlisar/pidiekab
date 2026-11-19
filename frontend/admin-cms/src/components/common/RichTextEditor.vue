<script setup lang="ts">
import { ref, watch, onMounted, computed, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: string
    placeholder?: string
    minHeight?: string
  }>(),
  {
    modelValue: '',
    placeholder: 'Tuliskan narasi konten lengkap di sini...',
    minHeight: '360px',
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

const editorRef = ref<HTMLDivElement | null>(null)
const sourceAreaRef = ref<HTMLTextAreaElement | null>(null)
const isSourceMode = ref(false)
const isFocused = ref(false)
const currentFormat = ref('p')

// Link modal state
const showLinkModal = ref(false)
const linkUrl = ref('')
const linkText = ref('')
let savedSelection: Range | null = null

// Word & Character count
const charCount = computed(() => {
  const text = (props.modelValue || '').replace(/<[^>]*>/g, '').trim()
  return text.length
})

const wordCount = computed(() => {
  const text = (props.modelValue || '').replace(/<[^>]*>/g, '').trim()
  if (!text) return 0
  return text.split(/\s+/).filter(Boolean).length
})

watch(
  () => props.modelValue,
  (val) => {
    if (editorRef.value && !isSourceMode.value) {
      if (editorRef.value.innerHTML !== val) {
        editorRef.value.innerHTML = val || ''
      }
    }
  }
)

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = props.modelValue || ''
  }
})

function onEditorInput() {
  if (!editorRef.value) return
  const html = editorRef.value.innerHTML
  emit('update:modelValue', html)
  emit('change', html)
}

function onSourceInput(e: Event) {
  const val = (e.target as HTMLTextAreaElement).value
  emit('update:modelValue', val)
  emit('change', val)
  if (editorRef.value) {
    editorRef.value.innerHTML = val
  }
}

function exec(command: string, value: string | undefined = undefined) {
  if (isSourceMode.value) return
  editorRef.value?.focus()
  document.execCommand(command, false, value)
  onEditorInput()
}

function setBlockFormat(tag: string) {
  if (isSourceMode.value) return
  currentFormat.value = tag
  exec('formatBlock', tag ? `<${tag}>` : '<p>')
}

function toggleSourceMode() {
  isSourceMode.value = !isSourceMode.value
  if (isSourceMode.value) {
    nextTick(() => {
      sourceAreaRef.value?.focus()
    })
  } else {
    nextTick(() => {
      if (editorRef.value) {
        editorRef.value.innerHTML = props.modelValue || ''
        editorRef.value.focus()
      }
    })
  }
}

function openLinkModal() {
  if (isSourceMode.value) return
  const sel = window.getSelection()
  if (sel && sel.rangeCount > 0) {
    savedSelection = sel.getRangeAt(0).cloneRange()
    linkText.value = sel.toString()
  } else {
    savedSelection = null
    linkText.value = ''
  }
  linkUrl.value = 'https://'
  showLinkModal.value = true
}

function applyLink() {
  if (!linkUrl.value || linkUrl.value === 'https://') {
    showLinkModal.value = false
    return
  }

  editorRef.value?.focus()
  if (savedSelection) {
    const sel = window.getSelection()
    sel?.removeAllRanges()
    sel?.addRange(savedSelection)
  }

  if (linkText.value && (!savedSelection || savedSelection.collapsed)) {
    // Insert new link with custom text
    const a = document.createElement('a')
    a.href = linkUrl.value
    a.textContent = linkText.value
    a.target = '_blank'
    a.rel = 'noopener noreferrer'
    savedSelection?.insertNode(a)
  } else {
    document.execCommand('createLink', false, linkUrl.value)
  }

  showLinkModal.value = false
  linkUrl.value = ''
  linkText.value = ''
  savedSelection = null
  onEditorInput()
}

function insertImagePrompt() {
  if (isSourceMode.value) return
  const url = prompt('Masukkan URL Gambar (https://...):')
  if (url && url.trim()) {
    exec('insertImage', url.trim())
  }
}

function cleanPastedHTML() {
  setTimeout(() => {
    onEditorInput()
  }, 10)
}
</script>

<template>
  <div class="rich-editor-container" :class="{ focused: isFocused }">
    <!-- Toolbar -->
    <div class="editor-toolbar">
      <!-- Format Block Selector -->
      <div class="toolbar-group">
        <select
          v-model="currentFormat"
          class="toolbar-select"
          :disabled="isSourceMode"
          @change="setBlockFormat(currentFormat)"
        >
          <option value="p">Paragraf Normal</option>
          <option value="h2">Heading 2 (Sub-Judul)</option>
          <option value="h3">Heading 3 (Bagian)</option>
          <option value="h4">Heading 4 (Kecil)</option>
          <option value="blockquote">Kutipan (Blockquote)</option>
          <option value="pre">Kode Preformatted</option>
        </select>
      </div>

      <div class="toolbar-divider"></div>

      <!-- Inline Styles -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn"
          title="Tebal (Bold)"
          :disabled="isSourceMode"
          @click="exec('bold')"
        >
          <strong>B</strong>
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Miring (Italic)"
          :disabled="isSourceMode"
          @click="exec('italic')"
        >
          <em>I</em>
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Garis Bawah (Underline)"
          :disabled="isSourceMode"
          @click="exec('underline')"
        >
          <u>U</u>
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Coret (Strikethrough)"
          :disabled="isSourceMode"
          @click="exec('strikeThrough')"
        >
          <s>S</s>
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <!-- Lists -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn"
          title="Daftar Butir (Bullet List)"
          :disabled="isSourceMode"
          @click="exec('insertUnorderedList')"
        >
          ☰
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Daftar Nomor (Numbered List)"
          :disabled="isSourceMode"
          @click="exec('insertOrderedList')"
        >
          1.
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Garis Pemisah (Horizontal Rule)"
          :disabled="isSourceMode"
          @click="exec('insertHorizontalRule')"
        >
          —
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <!-- Alignment -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn"
          title="Rata Kiri"
          :disabled="isSourceMode"
          @click="exec('justifyLeft')"
        >
          ⇤
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Rata Tengah"
          :disabled="isSourceMode"
          @click="exec('justifyCenter')"
        >
          ≡
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Rata Kanan"
          :disabled="isSourceMode"
          @click="exec('justifyRight')"
        >
          ⇥
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <!-- Links & Media -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn"
          title="Sisipkan Tautan (Link)"
          :disabled="isSourceMode"
          @click="openLinkModal"
        >
          🔗
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Sisipkan Gambar dari URL"
          :disabled="isSourceMode"
          @click="insertImagePrompt"
        >
          🖼️
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Hapus Format (Clean Formatting)"
          :disabled="isSourceMode"
          @click="exec('removeFormat')"
        >
          🧹
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <!-- Undo / Redo -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn"
          title="Undo"
          :disabled="isSourceMode"
          @click="exec('undo')"
        >
          ↶
        </button>
        <button
          type="button"
          class="tb-btn"
          title="Redo"
          :disabled="isSourceMode"
          @click="exec('redo')"
        >
          ↷
        </button>
      </div>

      <div class="toolbar-spacer"></div>

      <!-- Source Toggle -->
      <div class="toolbar-group">
        <button
          type="button"
          class="tb-btn btn-mode"
          :class="{ active: isSourceMode }"
          :title="isSourceMode ? 'Kembali ke Editor Visual (WYSIWYG)' : 'Lihat / Edit Kode HTML Langsung'"
          @click="toggleSourceMode"
        >
          {{ isSourceMode ? '👁️ Mode Visual' : '&lt;/&gt; Kode HTML' }}
        </button>
      </div>
    </div>

    <!-- Visual Editor Area -->
    <div
      v-show="!isSourceMode"
      ref="editorRef"
      class="editor-content-area"
      contenteditable="true"
      :style="{ minHeight }"
      :data-placeholder="placeholder"
      @input="onEditorInput"
      @paste="cleanPastedHTML"
      @focus="isFocused = true"
      @blur="isFocused = false"
    ></div>

    <!-- HTML Source Code Mode Area -->
    <textarea
      v-show="isSourceMode"
      ref="sourceAreaRef"
      :value="modelValue"
      class="editor-source-area"
      :style="{ minHeight }"
      placeholder="Masukkan kode HTML langsung di sini..."
      @input="onSourceInput"
      @focus="isFocused = true"
      @blur="isFocused = false"
    ></textarea>

    <!-- Footer Stats -->
    <div class="editor-footer">
      <div class="footer-hint">
        <span v-if="!isSourceMode">Mode: <strong>WYSIWYG Visual</strong> (Kompatibel dengan Go Template)</span>
        <span v-else class="text-amber">Mode: <strong>HTML Source Editor</strong></span>
      </div>
      <div class="footer-stats">
        <span><strong>{{ wordCount }}</strong> kata</span>
        <span class="dot">•</span>
        <span><strong>{{ charCount }}</strong> karakter</span>
      </div>
    </div>

    <!-- Link Modal -->
    <div v-if="showLinkModal" class="link-modal-backdrop" @click.self="showLinkModal = false">
      <div class="link-modal">
        <h4 class="modal-title">Sisipkan Tautan (Link)</h4>
        <div class="modal-body">
          <div class="form-group mb-2">
            <label class="modal-label">Teks Tautan</label>
            <input
              v-model="linkText"
              type="text"
              class="modal-input"
              placeholder="Contoh: Kunjungi Situs Resmi"
            />
          </div>
          <div class="form-group mb-3">
            <label class="modal-label">URL Tujuan *</label>
            <input
              v-model="linkUrl"
              type="url"
              class="modal-input"
              placeholder="https://..."
              required
              @keydown.enter.prevent="applyLink"
            />
          </div>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary btn-sm" @click="showLinkModal = false">
            Batal
          </button>
          <button type="button" class="btn btn-primary btn-sm" @click="applyLink">
            Terapkan Tautan
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rich-editor-container {
  display: flex;
  flex-direction: column;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background-color: #ffffff;
  overflow: hidden;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.rich-editor-container.focused {
  border-color: #0d9488;
  box-shadow: 0 0 0 3px rgba(13, 148, 136, 0.12);
}

/* Toolbar */
.editor-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  background-color: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  user-select: none;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
}

.toolbar-divider {
  width: 1px;
  height: 20px;
  background-color: #cbd5e1;
  margin: 0 4px;
}

.toolbar-spacer {
  flex: 1;
}

.toolbar-select {
  height: 30px;
  padding: 2px 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #334155;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  outline: none;
  cursor: pointer;
}

.toolbar-select:focus {
  border-color: #0d9488;
}

.tb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
  font-size: 0.875rem;
  font-weight: 600;
  color: #475569;
  background-color: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s, color 0.15s;
}

.tb-btn:hover:not(:disabled) {
  background-color: #e2e8f0;
  color: #0f172a;
}

.tb-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.tb-btn.btn-mode {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 4px 10px;
  background-color: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  color: #334155;
}

.tb-btn.btn-mode:hover {
  background-color: #e2e8f0;
}

.tb-btn.btn-mode.active {
  background-color: #0d9488;
  border-color: #0f766e;
  color: #ffffff;
}

/* Content Area */
.editor-content-area {
  padding: 16px 18px;
  font-size: 0.95rem;
  line-height: 1.7;
  color: #1e293b;
  outline: none;
  overflow-y: auto;
  word-break: break-word;
}

.editor-content-area:empty:before {
  content: attr(data-placeholder);
  color: #94a3b8;
  pointer-events: none;
}

/* Typographic defaults inside the visual editor */
.editor-content-area :deep(h2) {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  margin: 1.25rem 0 0.5rem;
}

.editor-content-area :deep(h3) {
  font-size: 1.15rem;
  font-weight: 600;
  color: #1e293b;
  margin: 1rem 0 0.4rem;
}

.editor-content-area :deep(h4) {
  font-size: 1rem;
  font-weight: 600;
  color: #334155;
  margin: 0.85rem 0 0.35rem;
}

.editor-content-area :deep(p) {
  margin: 0 0 0.85rem;
}

.editor-content-area :deep(blockquote) {
  border-left: 4px solid #0d9488;
  padding-left: 14px;
  margin: 1rem 0;
  color: #475569;
  font-style: italic;
  background-color: #f8fafc;
  padding: 10px 14px;
  border-radius: 0 6px 6px 0;
}

.editor-content-area :deep(ul),
.editor-content-area :deep(ol) {
  padding-left: 24px;
  margin: 0 0 0.85rem;
}

.editor-content-area :deep(li) {
  margin-bottom: 0.25rem;
}

.editor-content-area :deep(a) {
  color: #0d9488;
  text-decoration: underline;
}

.editor-content-area :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 6px;
  margin: 0.75rem 0;
}

.editor-content-area :deep(hr) {
  border: 0;
  height: 1px;
  background-color: #e2e8f0;
  margin: 1.25rem 0;
}

.editor-content-area :deep(pre) {
  background-color: #0f172a;
  color: #f8fafc;
  padding: 12px 14px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.85rem;
  overflow-x: auto;
  margin: 0.85rem 0;
}

/* Source Code Mode Area */
.editor-source-area {
  width: 100%;
  padding: 16px 18px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.875rem;
  line-height: 1.6;
  color: #0f172a;
  background-color: #fafbfc;
  border: none;
  outline: none;
  resize: vertical;
}

/* Footer Stats */
.editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background-color: #f8fafc;
  border-top: 1px solid #e2e8f0;
  font-size: 0.75rem;
  color: #64748b;
}

.footer-stats {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot {
  color: #cbd5e1;
}

.text-amber {
  color: #b45309;
}

/* Link Modal */
.link-modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.link-modal {
  width: 90%;
  max-width: 440px;
  background-color: #ffffff;
  border-radius: 10px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.15), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  padding: 20px;
}

.modal-title {
  margin: 0 0 14px;
  font-size: 1rem;
  font-weight: 600;
  color: #0f172a;
}

.modal-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #475569;
  margin-bottom: 4px;
}

.modal-input {
  width: 100%;
  height: 36px;
  padding: 0 10px;
  font-size: 0.875rem;
  color: #0f172a;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  outline: none;
}

.modal-input:focus {
  border-color: #0d9488;
  box-shadow: 0 0 0 2px rgba(13, 148, 136, 0.15);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 14px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 0.8125rem;
}
</style>
