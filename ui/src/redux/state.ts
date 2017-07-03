import {
  createStore,
  combineReducers,
  applyMiddleware,
  compose,
  GenericStoreEnhancer
} from 'redux';
import thunk from 'redux-thunk';
import {
  syncHistoryWithStore,
  routerReducer,
  RouterState
} from 'react-router-redux';
import { browserHistory } from 'react-router';

import {
  State as ProcessState,
  reducer as processReducer
} from './reducers/process';
import {
  State as ProjectState,
  reducer as projectReducer
} from './reducers/project';

export interface UIState {
  routing: RouterState;
  process: ProcessState;
  project: ProjectState;
}

export const store = createStore(
  combineReducers({
    routing: routerReducer,
    process: processReducer,
    project: projectReducer
  }),
  compose(applyMiddleware(thunk))
);
export const history = syncHistoryWithStore(browserHistory, store);
