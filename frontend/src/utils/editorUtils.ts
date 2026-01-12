import { LanguageSupport, StreamLanguage, HighlightStyle } from '@codemirror/language';
import { Tag, tags } from '@lezer/highlight';
import { json } from '@codemirror/lang-json';
import { xml } from '@codemirror/lang-xml';
import { html } from '@codemirror/lang-html';

// 定义HTTP特定标签
const httpMethod = Tag.define();
const httpStatus = Tag.define();
const headerName = Tag.define();
const headerValue = Tag.define();
const httpHeader = Tag.define(); // 为兼容旧代码

// 创建自定义高亮样式
export const createHighlightStyle = (isDarkMode: boolean) => {
  return HighlightStyle.define([
    // 基本语法元素
    { tag: tags.keyword, color: isDarkMode ? "#93c5fd" : "#3b82f6", fontWeight: "bold" },
    { tag: tags.propertyName, color: isDarkMode ? "#93c5fd" : "#3b82f6", fontWeight: "bold" },
    { tag: tags.attributeName, color: isDarkMode ? "#93c5fd" : "#3b82f6", fontWeight: "bold" },
    { tag: tags.string, color: isDarkMode ? "#d1d5db" : "#374151" }, // 保持普通颜色
    { tag: tags.attributeValue, color: isDarkMode ? "#d1d5db" : "#374151" },
    { tag: tags.number, color: isDarkMode ? "#c084fc" : "#8b5cf6" },
    { tag: tags.comment, color: isDarkMode ? "#9ca3af" : "#6b7280", fontStyle: "italic" }, // 用于空行
    { tag: tags.heading, color: isDarkMode ? "#f472b6" : "#ec4899", fontWeight: "bold" },
    { tag: tags.punctuation, color: isDarkMode ? "#e5e7eb" : "#4b5563" },
    
    // HTTP 特定样式
    { tag: httpMethod, color: isDarkMode ? "#60a5fa" : "#2563eb", fontWeight: "bold" }, // HTTP 方法 - 蓝色
    { tag: httpStatus, color: isDarkMode ? "#f87171" : "#dc2626", fontWeight: "bold" }, // HTTP 状态码 - 红色
    { tag: headerName, color: isDarkMode ? "#a78bfa" : "#8b5cf6", fontWeight: "bold" }, // HTTP 头部名称 - 紫色
    { tag: headerValue, color: isDarkMode ? "#d1d5db" : "#374151" }, // HTTP 头部值 - 普通颜色
    { tag: httpHeader, color: isDarkMode ? "#f87171" : "#dc2626", fontWeight: "bold" }, // 为兼容旧代码
  ]);
};

// 创建 HTTP 语法高亮（为兼容旧代码）
export function httpHeadersLanguage() {
  // 修改方法让它返回字符串标识符，然后在 props 中映射到 Tag
  const languageDef = StreamLanguage.define({
    name: "http",
    tokenTable: {
      method: httpMethod,
      header: httpHeader
    },
    startState() { return { }; },
    token(stream, _state) {
      // 如果在行首
      if (stream.sol()) {
        // 跳过前导空格
        while (stream.peek() === ' ' || stream.peek() === '\t') {
          stream.next();
        }
        
        // 检查是否匹配 HTTP 方法（请求行）
        if (stream.match(/GET /i) || stream.match(/POST /i) || 
            stream.match(/PUT /i) || stream.match(/DELETE /i) || 
            stream.match(/HEAD /i) || stream.match(/OPTIONS /i) || 
            stream.match(/PATCH /i) || stream.match(/CONNECT /i) || 
            stream.match(/TRACE /i)) {
          stream.skipToEnd();
          return "method";
        }
        
        // 检查是否匹配 HTTP 响应状态行
        if (stream.match(/HTTP\/[0-9.]+\s+\d+/i)) {
          stream.skipToEnd();
          return "method";
        }
        
        // 检查是否匹配 HTTP 头部名称
        const headerRegex = /^[A-Za-z0-9-]+:/i;
        const headerText = stream.string.slice(stream.pos);
        if (headerRegex.test(headerText)) {
          const colonPos = headerText.indexOf(':');
          stream.pos += colonPos + 1;
          return "header";
        }
        
        // 空行
        if (stream.eol()) {
          return "comment";
        }
      }
      
      // 对于其他所有内容（包括头部值和请求体），不特殊高亮
      stream.skipToEnd();
      return null;
    }
  });
  
  return new LanguageSupport(languageDef);
}

// 创建混合HTTP语法高亮（头部+主体）
export function httpMixedLanguage(isResponse = false) {
  // 缓存正则表达式，避免重复创建
  const headerRegex = /^[A-Za-z0-9-]+:/i;
  const methodRegex = /(GET|POST|PUT|DELETE|HEAD|OPTIONS|PATCH|CONNECT|TRACE) /i;
  const statusRegex = /HTTP\/[0-9.]+\s+\d+/i;
  
  return new LanguageSupport(StreamLanguage.define({
    name: "http-mixed",
    tokenTable: {
      method: httpMethod,
      status: httpStatus,
      headerName: headerName,
      headerValue: headerValue
    },
    startState() { 
      return { inBody: false }; 
    },
    token(stream, state) {
      // 已经在请求体部分，不需要尝试高亮HTTP标记
      if (state.inBody) {
        stream.skipToEnd();
        return null;
      }
      
      // 如果在行首，处理HTTP方法/响应状态和头部
      if (stream.sol()) {
        // 跳过前导空格
        if (stream.eatSpace()) return null;
        // 空行表示请求体开始
        if (stream.eol()) {
          state.inBody = true;
          return null;
        }
        
        // 检查HTTP请求方法 - 使用缓存的正则
        if (!isResponse && stream.match(methodRegex)) {
          stream.skipToEnd();
          return "method";
        }
        
        // 检查HTTP响应状态 - 使用缓存的正则
        if (isResponse && stream.match(statusRegex)) {
          // 高亮状态码部分
          const content = stream.string;
          const match = /HTTP\/[0-9.]+\s+(\d+)/.exec(content);
          if (match && match[1]) {
            const pos = content.indexOf(match[1], stream.pos - 10);
            if (pos >= 0) {
              stream.pos = pos + match[1].length;
              return "status";
            }
          }
          stream.skipToEnd();
          return null;
        }
        
        // 检查HTTP头部 - 使用缓存的正则
        const headerText = stream.string;
        if (headerRegex.test(headerText)) {
          const colonPos = headerText.indexOf(':');
          if (colonPos > 0) {
            stream.pos = colonPos + 1;
            return "headerName";
          }
        }
      } 
      // 行中的内容 - 头部值
      else if (stream.string.indexOf(':') > 0 && stream.pos > stream.string.indexOf(':')) {
        stream.skipToEnd();
        return "headerValue";
      }
      
      // 其他内容不做特殊处理
      stream.skipToEnd();
      return null;
    }
  }));
}

// 为HTTP请求/响应创建语言支持，自动处理混合语法高亮
// 为HTTP请求/响应创建语言支持，自动处理混合语法高亮
export function createHttpMixedEditor(content: string, isResponse = false) {
  // 只有非空内容才检测类型，避免空内容时不必要的处理
  const detectedType = content.trim() ? detectContentType(content) : 'text';
  
  // 添加HTTP混合语法高亮支持 - 这是关键修改
  const extensions = [httpMixedLanguage(isResponse)];

  // 为主体部分添加适当的语言支持
  if (detectedType === 'json') {
    extensions.push(json());
  } else if (detectedType === 'xml') {
    extensions.push(xml());
  } else if (detectedType === 'html') {
    extensions.push(html());
  }

  return extensions;
}

// 检测内容类型（统一处理函数，同时支持分离和混合内容）
export function detectContentType(content: string, body?: string): string {
  // 避免处理空内容
  if (!content || typeof content !== 'string' || !content.trim()) {
    return 'text';
  }
  
  // 如果提供了单独的body参数，使用这个分支
  if (body !== undefined) {
    if (typeof body !== 'string') body = String(body || '');
    if (!body || !body.trim()) return 'text';
    
    // 从content作为headers中获取Content-Type
    const headers = typeof content !== 'string' ? String(content || '') : content;
    const contentTypeMatch = headers.match(/content-type:\s*([^\r\n]+)/i);
    const contentType = contentTypeMatch ? contentTypeMatch[1].toLowerCase() : '';
    
    // 根据Content-Type或内容特征检测类型
    try {
      // 检测JSON
      if (contentType.includes('application/json') || 
          (body.trim().startsWith('{') && body.trim().endsWith('}')) || 
          (body.trim().startsWith('[') && body.trim().endsWith(']'))) {
        JSON.parse(body);
        return 'json';
      }
    } catch (e) {
      // JSON解析失败，继续检查其他类型
    }
    
    // 检测XML/HTML
    if (contentType.includes('application/xml') || 
        contentType.includes('text/xml') || 
        contentType.includes('text/html') || 
        body.includes('<?xml') || 
        body.includes('<!DOCTYPE') || 
        body.includes('<html')) {
      
      if (body.includes('<!DOCTYPE html>') || body.includes('<html')) {
        return 'html';
      }
      
      return 'xml';
    }
    
    return 'text';
  }
  
  // 处理完整HTTP内容
  // 查找空行后的主体部分
  const parts = content.split(/\r?\n\r?\n/);
  if (parts.length < 2) return 'text';
  
  const bodyContent = parts.slice(1).join('\n\n').trim();
  if (!bodyContent) return 'text';
  
  // 从头部获取Content-Type
  const headers = parts[0];
  const contentTypeMatch = headers.match(/content-type:\s*([^\r\n]+)/i);
  const contentType = contentTypeMatch ? contentTypeMatch[1].toLowerCase() : '';
  
  // 根据Content-Type或内容特征检测类型
  try {
    // 检测JSON
    if (contentType.includes('application/json') || 
        (bodyContent.trim().startsWith('{') && bodyContent.trim().endsWith('}')) || 
        (bodyContent.trim().startsWith('[') && bodyContent.trim().endsWith(']'))) {
      
      // 尝试解析JSON以验证格式
      JSON.parse(bodyContent);
      return 'json';
    }
  } catch (e) {
    // JSON解析失败，继续检查其他类型
  }
  
  // 检测XML/HTML
  if (contentType.includes('application/xml') || 
      contentType.includes('text/xml') || 
      contentType.includes('text/html') || 
      bodyContent.includes('<?xml') || 
      bodyContent.includes('<!DOCTYPE') || 
      bodyContent.includes('<html')) {
    
    if (bodyContent.includes('<!DOCTYPE html>') || bodyContent.includes('<html')) {
      return 'html';
    }
    
    return 'xml';
  }
  
  return 'text';
}

// 编辑器主题创建
export const createEditorTheme = (isDarkMode: boolean) => {
  return {
    "&": {
      fontSize: "12px",
      height: "100%"
    },
    ".cm-scroller": {
      fontFamily: "monospace",
      overflow: "auto"
    },
    ".cm-content": {
      minHeight: "100px",
      color: isDarkMode ? "#f3f4f6" : "#111827",
      padding: "8px",
      width: "max-content",
      minWidth: "100%"
    },
    ".cm-line": {
      padding: "0 4px"
    },
    ".cm-gutters": {
      backgroundColor: isDarkMode ? "#282838" : "#f3f4f6",
      color: isDarkMode ? "#9ca3af" : "#6b7280",
      border: "none",
      minWidth: "40px",     // 增加最小宽度确保行号显示
      paddingRight: "8px",  // 增加右侧内边距
      paddingLeft: "8px"    // 增加左侧内边距
    },
    ".cm-gutter": {
      minWidth: "28px"      // 设置单个gutter的最小宽度
    },
    ".cm-lineNumbers": {
      minWidth: "28px"      // 专门为行号设置最小宽度
    },
    ".cm-gutterElement": {
      textAlign: "right",    // 确保行号右对齐
      paddingRight: "4px"    // 小的右侧内边距让数字不贴边
    },
    ".cm-activeLineGutter": {
      backgroundColor: isDarkMode ? "#3e3e5e" : "#e5e7eb"
    }
  };
}; 