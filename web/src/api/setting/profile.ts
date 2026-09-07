import { Alova } from '@/utils/http/alova/index';

export interface ProfileResult {
  nickname: string;
  email: string;
  phone: string;
  address: string;
  avatar: string;
}

export function getProfile() {
  return Alova.Get<ProfileResult>('/account/profile');
}

export function updateProfile(data) {
  return Alova.Put('/account/profile', data);
}

export function changePassword(data) {
  return Alova.Put('/account/password', data);
}

/** 上传头像（png/jpg/jpeg/gif/webp，≤2MB），返回带版本号的头像 URL */
export function uploadAvatar(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return Alova.Put<{ avatar: string }>('/account/avatar', formData);
}
