import { Alova } from '@/utils/http/alova/index';

export function getProfile() {
  return Alova.Get('/account/profile');
}

export function updateProfile(data) {
  return Alova.Put('/account/profile', data);
}

export function changePassword(data) {
  return Alova.Put('/account/password', data);
}
