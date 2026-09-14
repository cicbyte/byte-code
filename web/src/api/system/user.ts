import { Alova } from '@/utils/http/alova/index';

/**
 * @description: 获取用户信息
 */
export function getUserInfo() {
  return Alova.Get<InResult>('/admin_info', {
    meta: {
      isReturnNativeResponse: true,
    },
  });
}

/**
 * @description: 用户登录
 */
export function login(params) {
  return Alova.Post<InResult>('/login', params, {
    meta: {
      isReturnNativeResponse: true,
    },
  });
}

/**
 * @description: 发起密码重置（防枚举：恒定成功）
 */
export function forgotPassword(account: string) {
  return Alova.Post<InResult>('/auth/forgot-password', { account });
}

/**
 * @description: 凭重置令牌设置新密码（单次/30min）
 */
export function resetPassword(token: string, newPassword: string) {
  return Alova.Post<InResult>('/auth/reset-password', { token, newPassword });
}

/**
 * @description: 用户修改密码
 */
export function changePassword(params, uid) {
  return Alova.Post(`/user/u${uid}/changepw`, { params });
}

/**
 * @description: 用户登出
 */
export function logout(params) {
  return Alova.Post('/login/logout', {
    params,
  });
}
