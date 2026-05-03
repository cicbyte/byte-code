import { IAsyncRouteState } from '@/store/modules/asyncRoute';
import { IUserState } from '@/store/modules/user';
import { IScreenLockState } from '@/store/modules/screenLock';

export interface IStore {
  asyncRoute: IAsyncRouteState;
  user: IUserState;
  screenLock: IScreenLockState;
  count: number;
}
