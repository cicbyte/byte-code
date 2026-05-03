import { Alova } from '@/utils/http/alova/index';

export interface TypeConsole {
  userCount: number;
  roleCount: number;
  menuCount: number;
  onlineUser: number;
}

export function getConsoleInfo() {
  return Alova.Get<TypeConsole>('/dashboard/console');
}
