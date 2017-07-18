import { isEqual, merge } from "lodash";
import * as React from "react";
import { connect } from "react-redux";
import { RouterState, Link } from "react-router";

import { UIState } from "../redux/state";
import { saveProcess } from "../redux/actions/process";
import { Process } from "../models";

interface State {
	name: string;
	importPath: string;
	branch: string;
}
interface Props extends RouterState {
	process?: Process;
	error?: string;
	saveProcess: typeof saveProcess;
}
class EditProcess extends React.Component<Props, State> {
	componentWillMount() {
		this.resetState();
	}
	componentWillReceiveProps(nextProps: Props) {
		if (!isEqual(this.props.process, nextProps.process)) {
			this.resetState();
		}
	}
	resetState() {
		this.setState({
			name: "",
			importPath: "",
			branch: ""
		});
	}

	handleSave(evt: React.FormEvent<any>) {
		evt.preventDefault();

		saveProcess({
			id: this.props.process.id,
			projectID: this.props.process.projectID,
			config: {
				name: this.state.name,
				importPath: this.state.importPath,
				branch: this.state.branch
			}
		});
	}

	render() {
		return (
			<div>
				<form onSubmit={evt => this.handleSave(evt)}>
					<fieldset>
						<label>Name</label>
						<input
							type="text"
							name="name"
							value={this.state.name}
							onChange={evt => this.setState({ name: evt.target.value })}
						/>
					</fieldset>

					<fieldset>
						<label>Import Path</label>
						<input
							type="text"
							name="importPath"
							value={this.state.importPath}
							onChange={evt => this.setState({ importPath: evt.target.value })}
						/>
					</fieldset>

					<fieldset>
						<label>Branch</label>
						<input
							type="text"
							name="branch"
							value={this.state.branch}
							onChange={evt => this.setState({ branch: evt.target.value })}
						/>
					</fieldset>

					<button>Save</button>
				</form>
			</div>
		);
	}
}

// const connected = connect(
// 	(state: UIState): State => {
// 		return {
// 			process: state.process.item,
// 			error: state.process.error,
// 			saveProcess: null
// 		};
// 	},
// 	{
// 		saveProcess
// 	}
// )(EditProcess);

export { EditProcess };
