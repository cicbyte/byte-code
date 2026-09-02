import { ref } from 'vue';

// 全局强制改密弹窗的显隐状态（单例）：
// 登录页检测到 mustChangePassword、或任意接口收到 code 1001 时置 true
export const forcePwdVisible = ref(false);
