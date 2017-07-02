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

import { projectListReducer, processListReducer } from './reducers';

export const store = createStore(
  combineReducers({
    routing: routerReducer,
    projectList: projectListReducer,
    processList: processListReducer
  }),
  compose(applyMiddleware(thunk))
);
export const history = syncHistoryWithStore(browserHistory, store);
