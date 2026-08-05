<template>
  <div ref="editorContainer" class="code-editor-container"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { EditorState } from '@codemirror/state'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection } from '@codemirror/view'
import { defaultKeymap, indentWithTab, history, historyKeymap } from '@codemirror/commands'
import { StreamLanguage, bracketMatching, foldGutter, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { go } from '@codemirror/legacy-modes/mode/go'
import { oneDark } from '@codemirror/theme-one-dark'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  readOnly: {
    type: Boolean,
    default: false
  },
  placeholder: {
    type: String,
    default: '// Write your Go code here...'
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const editorContainer = ref(null)
let editorView = null
let isUpdatingFromProp = false

onMounted(() => {
  if (!editorContainer.value) return

  const updateListener = EditorView.updateListener.of((update) => {
    if (update.docChanged && !isUpdatingFromProp) {
      const value = update.state.doc.toString()
      emit('update:modelValue', value)
      emit('change', value)
    }
  })

  const customTheme = EditorView.theme({
    '&': {
      height: '100%',
      fontSize: '14px',
      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
      borderRadius: '0.5rem',
      overflow: 'hidden'
    },
    '.cm-scroller': {
      fontFamily: 'inherit',
      overflow: 'auto',
      lineHeight: '1.6'
    },
    '.cm-content': {
      padding: '12px 0'
    },
    '.cm-line': {
      padding: '0 16px'
    },
    '.cm-gutters': {
      backgroundColor: '#111827',
      color: '#6b7280',
      borderRight: '1px solid #1f2937'
    },
    '&.cm-focused': {
      outline: 'none'
    }
  })

  const state = EditorState.create({
    doc: props.modelValue,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightActiveLine(),
      drawSelection(),
      history(),
      bracketMatching(),
      foldGutter(),
      StreamLanguage.define(go),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      oneDark,
      customTheme,
      keymap.of([
        ...defaultKeymap,
        ...historyKeymap,
        indentWithTab
      ]),
      EditorState.readOnly.of(props.readOnly),
      updateListener
    ]
  })

  editorView = new EditorView({
    state,
    parent: editorContainer.value
  })
})

watch(() => props.modelValue, (newValue) => {
  if (editorView && newValue !== editorView.state.doc.toString()) {
    isUpdatingFromProp = true
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: newValue
      }
    })
    isUpdatingFromProp = false
  }
})

onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy()
  }
})
</script>

<style scoped>
.code-editor-container {
  width: 100%;
  height: 100%;
  min-height: 350px;
  background-color: #0d1117;
  border-radius: 0.5rem;
  border: 1px solid #1f2937;
  overflow: hidden;
}

:deep(.cm-editor) {
  height: 100%;
}
</style>
