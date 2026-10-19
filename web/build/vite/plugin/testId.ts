import type { Plugin } from 'vite';

/**
 * 生产构建剥离 data-test-id（集成测试锚点）：
 * 源码里保留供 DrissionPage 定位（dev / 测试模式），`vite build` 时从
 * .vue 模板与 .ts(x) 中整属性移除，产物不携带测试锚点。
 * 覆盖四种形态（enforce: pre 先于 vue 编译处理模板源码）：
 * - 静态 `data-test-id="..."` 与动态 `:data-test-id="expr"` 模板属性
 * - 对象属性 `"data-test-id": expr`（nodeProps 等运行时属性透传）：
 *   带尾逗号（首位/中间属性）、带前逗号（末位属性）两步剥离保证
 *   对象字面量逗号仍合法；唯一属性（无逗号）单独整对剥离
 */
const TEST_ID_ATTR = /\s+:?data-test-id="[^"]*"/g;
const TEST_ID_PROP_TRAILING = /['"]data-test-id['"]\s*:\s*(?:`[^`]*`|"[^"]*"|'[^']*')\s*,/g;
const TEST_ID_PROP_LEADING = /,\s*['"]data-test-id['"]\s*:\s*(?:`[^`]*`|"[^"]*"|'[^']*')/g;
const TEST_ID_PROP_SOLE = /\{\s*['"]data-test-id['"]\s*:\s*(?:`[^`]*`|"[^"]*"|'[^']*')\s*\}/g;

export function configTestIdStripPlugin(isBuild: boolean) {
  if (!isBuild) return [];
  return [
    {
      name: 'bytecode-strip-test-id',
      enforce: 'pre',
      transform(code: string, id: string) {
        if (!/\.(vue|tsx?|jsx?)($|\?)/.test(id)) return null;
        if (!code.includes('data-test-id')) return null;
        const stripped = code
          .replace(TEST_ID_ATTR, '')
          .replace(TEST_ID_PROP_TRAILING, '')
          .replace(TEST_ID_PROP_LEADING, '')
          .replace(TEST_ID_PROP_SOLE, '{}');
        return { code: stripped, map: null };
      },
    } as Plugin,
  ];
}
