import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { Router, Route, IndexRoute, IndexRedirect } from 'react-router';

import { Root } from './containers/root';
import { EditProcess } from './containers/EditProcess';
import { ProjectList } from './containers/ProjectList';
import { Workbench } from './containers/Workbench';
import { store, history } from './redux/state';

ReactDOM.render(
  <Provider store={store}>
    <Router history={history}>
      <Route path="/" component={Root}>
        <IndexRedirect to="projects" />
        <Route path="projects" component={ProjectList} />
        <Route path="projects/:projectID" component={Workbench} />
        <Route
          path="projects/:projectID/processes/create"
          component={EditProcess}
        />
      </Route>
    </Router>
  </Provider>,
  document.getElementById('root')
);
/*
           />*/
