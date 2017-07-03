import * as React from 'react';
import { connect } from 'react-redux';
import { RouterState, Link } from 'react-router';

import { UIState } from '../redux/state';
import { deleteProcess, listProcesses } from '../redux/actions/process';
import { Process } from '../models';

interface State {
  processes: Array<Process>;
  error?: string;
}
interface Props extends RouterState, State {
  deleteProcess: typeof deleteProcess;
  listProcesses: typeof listProcesses;
}
class Workbench extends React.Component<Props, {}> {
  componentWillMount() {
    this.props.listProcesses(parseInt(this.props.params.projectID, 10));
  }

  componentWillReceiveProps(nextProps: Props) {
    if (nextProps.params.projectID !== this.props.params.projectID) {
      this.props.listProcesses(parseInt(this.props.params.projectID, 10));
    }
  }

  handleClickDelete(evt: React.MouseEvent<any>, processID: number) {
    this.props.deleteProcess(processID);
  }

  render() {
    return (
      <div>
        <h2>
          <Link to="/">Projects</Link>
        </h2>
        <h3>Processes</h3>
        {this.props.error
          ? <div className="error">
              Failed to fetch projects: {this.props.error}.
            </div>
          : null}
        <Link
          to={'/projects/' + this.props.params.projectID + '/processes/create'}
        >
          Create Process
        </Link>
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Import Path</th>
              <th>Branch</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {this.props.processes.map(process =>
              <tr>
                <td>
                  {process.config.name}
                </td>
                <td>
                  {process.config.importPath}
                </td>
                <td>
                  {process.config.branch}
                </td>
                <td>
                  <a
                    href="#delete"
                    onClick={evt => this.handleClickDelete(evt, process.id)}
                  >
                    Delete
                  </a>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

const workbenchConnected = connect(
  (state: UIState): State => {
    return {
      processes: state.process.list || [],
      error: state.process.error
    };
  },
  {
    deleteProcess,
    listProcesses
  }
)(Workbench);

export { workbenchConnected as Workbench };
