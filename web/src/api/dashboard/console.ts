import { Alova } from '@/utils/http/alova/index';

export interface TypeConsole {
  userCount: number;
  roleCount: number;
  menuCount: number;
  onlineUser: number;
}

export interface DashboardStats {
  totalRequirements: number;
  totalTasks: number;
  inProgress: number;
  pendingReview: number;
  testPassRate: number;
  aiEfficiencyRank: AIRankItem[];
  recentTasks: RecentTask[];
}

export interface AIRankItem {
  name: string;
  value: number;
}

export interface RecentTask {
  id: number;
  title: string;
  status: string;
  assignee: string;
  updated_at: string;
}

export function getConsoleInfo() {
  return Alova.Get<TypeConsole>('/dashboard/console');
}

export function getDashboardStats() {
  return Alova.Get<DashboardStats>('/dashboard/stats');
}
