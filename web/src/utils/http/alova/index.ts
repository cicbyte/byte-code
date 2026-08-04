import { createAlova } from 'alova';
import VueHook from 'alova/vue';
import adapterFetch from 'alova/fetch';
import { isString } from 'lodash-es';
import { useUser } from '@/store/modules/user';
import { storage } from '@/utils/Storage';
import { useGlobSetting } from '@/hooks/setting';
import { ResultEnum } from '@/enums/httpEnum';
import { isUrl } from '@/utils';

const { apiUrl, urlPrefix } = useGlobSetting();

// 会话过期跳转中标记：并发多个 401 时只清理/跳转一次（整页跳转会自动重置）
let sessionExpiredRedirecting = false;

function handleSessionExpired(message?: string) {
  const Message = window.$message;
  Message?.error(message || '登录已过期，请重新登录');
  if (sessionExpiredRedirecting || window.location.pathname.startsWith('/login')) {
    return;
  }
  sessionExpiredRedirecting = true;
  const userStore = useUser();
  userStore.logout();
  // 整页跳转保证 Pinia 与本地缓存彻底重置，并携带回跳地址
  const redirect = encodeURIComponent(window.location.pathname + window.location.search);
  window.location.href = `/login?redirect=${redirect}`;
}

function normalizeResponse(res: any) {
  if (res.result !== undefined) {
    return res;
  }
  if (res.data !== undefined) {
    return {
      code: res.code === 0 ? ResultEnum.SUCCESS : res.code,
      result: res.data,
      message: res.message || 'ok',
    };
  }
  return res;
}

export const Alova = createAlova({
  baseURL: apiUrl,
  statesHook: VueHook,
  cacheFor: 0,
  cacheLogger: process.env.NODE_ENV === 'development',
  requestAdapter: adapterFetch(),
  beforeRequest(method) {
    const userStore = useUser();
    const token = userStore.getToken;
    if (!method.meta?.ignoreToken && token) {
      method.config.headers['token'] = token;
    }
    const isUrlStr = isUrl(method.url as string);
    if (!isUrlStr && urlPrefix) {
      method.url = `${urlPrefix}${method.url}`;
    }
    if (!isUrlStr && apiUrl && isString(apiUrl)) {
      method.url = `${apiUrl}${method.url}`;
    }
  },
  responded: {
    onSuccess: async (response, method) => {
      const raw = (response.json && (await response.json())) || response.body;
      const res = normalizeResponse(raw);

      const { message, code, result } = res;

      if (code === 401 || code === 912) {
        handleSessionExpired(message);
        throw new Error(message || '登录已过期');
      }

      // 强制改密：账号仍在使用初始密码，跳转到账号安全页完成修改
      if (code === ResultEnum.MUST_CHANGE_PASSWORD) {
        const Message = window.$message;
        Message?.error(message || '首次登录请先修改初始密码');
        if (!window.location.pathname.startsWith('/setting/account')) {
          window.location.href = '/setting/account';
        }
        throw new Error(message || '首次登录请先修改初始密码');
      }

      if (method.meta?.isReturnNativeResponse) {
        return res;
      }

      if (method.meta?.isTransformResponse === false) {
        return result;
      }

      const Message = window.$message;

      if (ResultEnum.SUCCESS === code) {
        return result;
      }

      Message?.error(message);
      throw new Error(message);
    },
  },
});
