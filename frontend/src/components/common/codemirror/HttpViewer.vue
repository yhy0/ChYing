<script setup lang="ts">
import { ref, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { EditorState } from '@codemirror/state';
import type { Extension } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { search, searchKeymap, openSearchPanel } from '@codemirror/search';
import { syntaxHighlighting } from '@codemirror/language';
import { 
  createHighlightStyle,
  createEditorTheme,
  createHttpMixedEditor
} from '../../../utils';

const props = defineProps<{
  data: string;
  readOnly?: boolean;
  isResponse?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:data', data: string): void
}>();

const editorRef = ref<HTMLElement | null>(null);
let editorView: EditorView | null = null;

// 创建编辑器
const createEditor = (
  element: HTMLElement, 
  content: string, 
  readOnly: boolean = false,
  isDarkMode: boolean = false
): EditorView => {
  // 确保content是字符串
  content = String(content || '');
  
  // 创建高亮样式
  const highlightStyle = createHighlightStyle(isDarkMode);
  
  const extensions: Extension[] = [
    lineNumbers(),
    highlightActiveLine(),
    history(),
    keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
    EditorState.readOnly.of(readOnly),
    // 应用语法高亮
    syntaxHighlighting(highlightStyle),
    // 添加搜索功能
    search({
      top: false,
      caseSensitive: false,
      wholeWord: false,
      regexp: false
    }),
    EditorView.theme({
      ".cm-panel.cm-search": {
        backgroundColor: isDarkMode ? 'var(--surface-color-dark, rgba(17, 24, 39, 0.95))' : 'var(--surface-color, rgba(255, 255, 255, 0.95))',
        color: isDarkMode ? 'var(--text-color-dark, #e5e7eb)' : 'var(--text-color, #334155)',
        border: 'none',
        borderTop: isDarkMode ? '1px solid var(--border-color-dark, #374151)' : '1px solid var(--border-color, #e2e8f0)',
        padding: '0 10px',
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 10,
        display: 'flex',
        alignItems: 'center',
        height: '36px',
        boxShadow: isDarkMode ? '0 -1px 3px rgba(0, 0, 0, 0.2)' : '0 -1px 3px rgba(0, 0, 0, 0.05)',
        backdropFilter: 'blur(8px)',
        WebkitBackdropFilter: 'blur(8px)',
        boxSizing: 'border-box',
      },
      // 简化隐藏替换相关UI的样式
      ".cm-search-replace, .cm-panel.cm-search button[name='select'], .cm-panel.cm-search button[name='replace'], .cm-panel.cm-search button[name='replaceAll'], .cm-panel.cm-search input[name='replace']": {
        display: "none !important",
      },
      // 简化隐藏关闭按钮和全选按钮的样式
      ".cm-panel.cm-search button[name='close'], .cm-panel.cm-search button[name='select-all']": {
        display: "none !important",
      },
      // 搜索输入框样式
      ".cm-textfield": {
        backgroundColor: isDarkMode ? '#1f2937' : 'white',
        color: isDarkMode ? '#e5e7eb' : '#334155',
        border: isDarkMode ? '1px solid #4b5563' : '1px solid #c7d2fe',
        borderRadius: '4px',
        padding: '3px 8px',
        fontSize: '13px',
        height: '26px',
        flexGrow: 1,
        minWidth: '150px',
        transition: 'all 0.2s ease',
        boxShadow: isDarkMode ? 'none' : 'inset 0 1px 2px rgba(0, 0, 0, 0.02)',
        '&:focus': {
          outline: 'none',
          borderColor: isDarkMode ? '#818cf8' : '#6366f1',
          boxShadow: isDarkMode
            ? '0 0 0 2px rgba(99, 102, 241, 0.2)'
            : '0 0 0 2px rgba(99, 102, 241, 0.15)',
        }
      },
      // 搜索按钮基础样式
      ".cm-button": {
        height: '28px',
        minWidth: '28px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: '4px',
        color: isDarkMode ? '#e5e7eb' : '#6366f1',
        background: 'transparent',
        border: 'none',
        padding: '0',
        margin: '0 2px',
        cursor: 'pointer',
        fontSize: '0', /* 隐藏原始文本 */
        fontWeight: '500',
        transition: 'all 0.15s ease',
        position: 'relative',
        '&:hover': {
          backgroundColor: isDarkMode ? 'rgba(55, 65, 81, 0.8)' : 'rgba(99, 102, 241, 0.1)',
          color: isDarkMode ? '#a5b4fc' : '#4f46e5',
        },
        '&:active': {
          backgroundColor: isDarkMode ? 'rgba(55, 65, 81, 0.9)' : 'rgba(99, 102, 241, 0.15)',
        }
      },
      // 隐藏按钮文字
      ".cm-button span": {
        fontSize: '0 !important',
        display: "none !important",
      },
      // 为 next 按钮添加图标 (向下箭头)
      ".cm-button[name='next']": {
        backgroundImage: isDarkMode
          ? "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24'%3E%3Cpath fill='%23e5e7eb' d='M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6l-6-6l1.41-1.41z'/%3E%3C/svg%3E\")"
          : "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24'%3E%3Cpath fill='%236366f1' d='M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6l-6-6l1.41-1.41z'/%3E%3C/svg%3E\")",
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center',
        backgroundSize: '18px 18px',
      },
      // 为 prev 按钮添加图标 (向上箭头)
      ".cm-button[name='prev']": {
        backgroundImage: isDarkMode
          ? "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24'%3E%3Cpath fill='%23e5e7eb' d='M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6l-6 6l1.41 1.41z'/%3E%3C/svg%3E\")"
          : "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24'%3E%3Cpath fill='%236366f1' d='M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6l-6 6l1.41 1.41z'/%3E%3C/svg%3E\")",
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center',
        backgroundSize: '18px 18px',
      },
      // 复选框标签样式 - 显示文本而不是图标
      ".cm-panel.cm-search label": {
        display: 'flex !important',
        alignItems: 'center !important',
        fontSize: '12px !important',
        margin: '0 4px !important',
        padding: '2px 6px !important',
        borderRadius: '4px !important',
        whiteSpace: 'nowrap !important',
        userSelect: 'none !important',
        cursor: 'pointer !important',
        color: isDarkMode ? '#9ca3af !important' : '#6366f1 !important',
        transition: 'all 0.15s ease !important',
        fontWeight: '500 !important',
        '&:hover': {
          color: isDarkMode ? '#e5e7eb !important' : '#4f46e5 !important',
          backgroundColor: isDarkMode ? 'rgba(55, 65, 81, 0.5) !important' : 'rgba(99, 102, 241, 0.1) !important',
        }
      },
      // 复选框样式优化
      ".cm-panel.cm-search input[type='checkbox']": {
        width: '14px !important',
        height: '14px !important',
        marginRight: '4px !important',
        cursor: 'pointer !important',
        borderRadius: '3px !important',
        accentColor: isDarkMode ? '#818cf8 !important' : '#6366f1 !important',
        '&:checked': {
          backgroundColor: isDarkMode ? '#818cf8 !important' : '#6366f1 !important',
        }
      },
      // 匹配计数样式
      ".cm-count": {
        fontSize: '12px !important',
        color: isDarkMode ? '#9ca3af !important' : '#6366f1 !important',
        marginLeft: '8px !important',
        minWidth: '3rem !important',
        fontFeatureSettings: '"tnum" !important',
        fontVariantNumeric: 'tabular-nums !important',
        fontWeight: '500 !important',
      },
      // 搜索面板布局
      ".cm-panel.cm-search form": {
        display: 'flex !important',
        alignItems: 'center !important',
        gap: '6px !important',
        width: '100%',
      },
      // 确保内容不被搜索框遮挡
      ".cm-content": {
        paddingBottom: "32px !important",
      },
      ".cm-scroller": {
        paddingBottom: "0 !important",
      }
    }),
    EditorView.updateListener.of(update => {
      if (update.docChanged && !readOnly) {
        // 文档内容变更时触发更新
        emit('update:data', update.state.doc.toString());
      }
    }),
    EditorView.theme(createEditorTheme(isDarkMode)),
    // 添加HTTP混合语法高亮
    ...createHttpMixedEditor(content, props.isResponse)
  ];

  // 创建编辑器实例
  const view = new EditorView({
    state: EditorState.create({
      doc: content,
      extensions: extensions
    }),
    parent: element
  });
  
  // 默认显示搜索面板
  setTimeout(() => {
    openSearchPanel(view);
  }, 100);
  
  return view;
};

// 上一次的内容，用于比较变化
let lastContent: string | null = null;

// 更新编辑器内容
const updateContent = () => {
  const content = props.data ?? '';
  
  // 如果编辑器不存在但DOM引用存在，创建它（优先检查编辑器是否存在）
  if (!editorView && editorRef.value) {
    editorView = createEditor(
      editorRef.value, 
      content, 
      props.readOnly,
      document.documentElement.classList.contains('dark')
    );
    // 缓存当前内容
    lastContent = content;
    return;
  }
  
  // 如果内容没有变化，不处理
  if (content === lastContent) return;

  // 更新已存在的编辑器内容
  if (editorView) {
    const currentContent = editorView.state.doc.toString();
    if (currentContent !== content) {
      // 防止循环更新：检查是否只是格式变化导致的微小差异
      // 比如，只是换行符数量或类型不同 (\n vs \r\n)
      const normalizedCurrent = currentContent.replace(/\r\n/g, '\n');
      const normalizedContent = content.replace(/\r\n/g, '\n');
      
      if (normalizedCurrent === normalizedContent) {
        // 内容实质相同，只是格式不同，直接更新缓存避免循环
        lastContent = currentContent;
        return;
      }
      
      // 保存当前光标位置和滚动位置
      const selection = editorView.state.selection;
      const scrollTop = editorView.scrollDOM.scrollTop;
      
      try {
        // 使用简单替换更新内容，避免复杂的差异计算
        editorView.dispatch({
          changes: {
            from: 0,
            to: currentContent.length,
            insert: content
          },
          selection: {
            anchor: Math.min(selection.main.anchor, content.length),
            head: Math.min(selection.main.head, content.length)
          }
        });
        
        // 恢复滚动位置
        editorView.scrollDOM.scrollTop = scrollTop;
      } catch (error) {
        console.error('更新编辑器时出错:', error);
        // 出错时，尝试重新创建编辑器
        if (editorRef.value) {
          editorView.destroy();
          editorView = createEditor(
            editorRef.value, 
            content, 
            props.readOnly,
            document.documentElement.classList.contains('dark')
          );
        }
      }
      
      // 缓存更新后的内容
      lastContent = editorView.state.doc.toString();
    }
  }
};

// 触发搜索功能
const triggerSearch = () => {
  if (editorView !== null) {
    openSearchPanel(editorView);
  }
};

// 以更干净的方式销毁编辑器实例
onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy();
    editorView = null;
  }
});

// 使用常量优化重复字符串的使用
const DARK_CLASS = 'dark';

// 添加防抖函数提高性能
let updateTimeout: number | null = null;
const debouncedUpdateContent = () => {
  if (updateTimeout) {
    clearTimeout(updateTimeout);
  }
  updateTimeout = window.setTimeout(() => {
    updateContent();
    updateTimeout = null;
  }, 10);
};

// 监听数据变化，使用防抖优化性能
watch(() => props.data, debouncedUpdateContent, { deep: true });

// 监听主题变化
watch(
  () => document.documentElement.classList.contains(DARK_CLASS),
  (isDark) => {
    if (editorView && editorRef.value) {
      // 保存当前内容和光标位置
      const content = editorView.state.doc.toString();
      const scrollTop = editorView.scrollDOM.scrollTop;
      
      // 销毁旧实例并创建新实例
      editorView.destroy();
      editorView = createEditor(editorRef.value, content, props.readOnly, isDark);
      
      // 恢复滚动位置
      editorView.scrollDOM.scrollTop = scrollTop;
    }
  }
);

// 立即更新方法，供外部组件调用
const forceUpdate = () => {
  if (updateTimeout) {
    clearTimeout(updateTimeout);
    updateTimeout = null;
  }
  updateContent();
};

// 在挂载时添加立即更新
onMounted(() => {
  nextTick(() => {
    updateContent();
  });

  // 监听主题变化
  const darkModeObserver = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.attributeName === 'class') {
        const isDarkMode = document.documentElement.classList.contains('dark');
        
        // 重新创建编辑器以应用新主题
        if (editorView && editorRef.value) {
          const content = editorView.state.doc.toString();
          editorView.destroy();
          editorView = createEditor(editorRef.value, content, props.readOnly, isDarkMode);
        }
      }
    });
  });
  
  darkModeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  });
});

// 暴露方法给父组件
defineExpose({
  triggerSearch,
  forceUpdate
});
</script>

<template>
  <div class="http-viewer h-full">
    <div ref="editorRef" class="h-full"></div>
  </div>
</template>

<style scoped>
.http-viewer {
  height: 100%;
  position: relative;
}

:deep(.cm-editor) {
  height: 100%;
}

:deep(.cm-scroller) {
  overflow: auto;
}

.cm-editor {
  height: 100%;
  width: 100%;
  overflow: auto !important;
}

.cm-scroller {
  overflow: auto !important;
}

.word-wrap .cm-content {
  white-space: pre-wrap !important;
}
</style> 