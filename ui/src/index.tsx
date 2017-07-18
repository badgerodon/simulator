import * as React from "react";
import * as ReactDOM from "react-dom";
import { Provider } from "react-redux";
import { Router, Route, IndexRoute, IndexRedirect } from "react-router";

import { Root } from "./containers/root";
import { ProjectList } from "./containers/ProjectList";
import { Workbench } from "./containers/Workbench";
import { Demos } from "./containers/Demos";
import { store, history } from "./redux/state";

ReactDOM.render(
	<Provider store={store}>
		<Router history={history}>
			<Route path="/" component={Root}>
				<IndexRedirect to="demos" />
				<Route path="projects" component={ProjectList} />
				<Route path="projects/:projectID" component={Workbench} />
				{/* <Route
					path="projects/:projectID/processes/create"
					component={EditProcess}
				/> */}
				<Route path="demos" component={Demos} />
				<Route path="demos/:name" component={Demos} />
			</Route>
		</Router>
	</Provider>,
	document.getElementById("root")
);
/*
           />*/
