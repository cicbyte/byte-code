import { Alova } from '@/utils/http/alova/index';

export function getSystemConfig() {
  return Alova.Get('/system/config');
}

export function updateSystemConfig(data) {
  return Alova.Put('/system/config', data);
}
