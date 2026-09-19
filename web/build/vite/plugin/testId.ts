import type { Plugin } from 'vite';

/**
 * 生产构建剥离 data-test-id（集成测试锚点）：
 * 源码里保留供 DrissionPage 定位（dev / 测试模式），`vite build` 时从
 * .vue 模板与 .ts(x) 中整属性移除，产物不携带测试锚点。
 * 同时覆盖静态 `data-test-id="..."` 与动态 `:data-test-id="expr"` 两种
 * 形态（enforce: pre 先于 vue 编译处理模板源码）。
 */
const TEST_ID_ATTR = /\s+:?data-test-id="[^"]*"/g;

export function configTestIdStripPlugin(isBuild: boolean) {
  if (!isBuild) return [];
  return [
    {
      name: 'bytecode-strip-test-id',
      enforce: 'pre',
      transform(code: string, id: string) {
        if (!/\.(vue|tsx?|jsx?)($|\?)/.test(id)) return null;
        if (!code.includes('data-test-id')) return null;
        const stripped = code.replace(TEST_ID_ATTR, '');
        return { code: stripped, map: null };
      },
    } as Plugin,
  ];
}
