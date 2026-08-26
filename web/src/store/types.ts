import { IAsyncRouteState } from '@/store/modules/asyncRoute';
import { IUserState } from '@/store/modules/user';

export interface IStore {
  asyncRoute: IAsyncRouteState;
  user: IUserState;
  count: number;
}
