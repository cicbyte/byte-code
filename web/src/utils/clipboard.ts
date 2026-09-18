/**
 * 剪贴板复制（安全上下文降级）。
 *
 * navigator.clipboard 仅在安全上下文（https / localhost）可用——
 * 内网 http 部署（如 http://host:port）下是 undefined，且可选链调用
 * 会静默跳过、连失败分支都不触发（表现为点复制毫无反应）。
 * 降级路径：隐藏 textarea + document.execCommand('copy')，
 * 需在用户手势同步栈内调用（各调用点均在 click handler 里，天然满足）。
 */
export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text);
  }
  return new Promise((resolve, reject) => {
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    ta.style.pointerEvents = 'none';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    try {
      // execCommand 已废弃但仍是最可靠的 http 降级；UA 策略变化时再评估
      document.execCommand('copy') ? resolve() : reject(new Error('execCommand returned false'));
    } catch (e) {
      reject(e);
    } finally {
      document.body.removeChild(ta);
    }
  });
}
