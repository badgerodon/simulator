import * as React from "react";
import { RouterState, Link } from "react-router";

interface DemoProps {
	name: string;
}

class DemoRunner extends React.Component<DemoProps, {}> {
	render() {
		return (
			<div>
				<p>
					{this.props.name}
				</p>
			</div>
		);
	}
}

class Demos extends React.Component<RouterState, {}> {
	render() {
		return (
			<div>
				<h1>Go Simulator Demos</h1>
				<ul>
					<li>
						<Link to={"/demos/hello"}>Hello World</Link>
					</li>
				</ul>
				{this.props.params.name
					? <DemoRunner name={this.props.params.name} />
					: null}
			</div>
		);
	}
}

export { Demos };
