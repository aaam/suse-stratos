import { UPDATE_HELM_RELEASE } from './workloads.actions';

const defaultState = {};

export function helmReleaseReducer(state: any = defaultState, action) {
  switch (action.type) {
    case UPDATE_HELM_RELEASE:
      return {
        ...state,
      };
  }
}
