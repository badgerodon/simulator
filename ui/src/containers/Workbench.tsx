import * as React from 'react';
import { RouterState, Link } from 'react-router';
import { connect } from 'react-redux';

import { fetchProcessList } from '../actions';
import { Process } from '../models';

interface WorkbenchProps extends RouterState {
  processes: Array<Process>;
  error?: string;
  fetchProcessList: typeof fetchProcessList;
}
class Workbench extends React.Component<WorkbenchProps, {}> {
  componentDidMount() {
    this.props.fetchProcessList(+this.props.params.projectID);
  }
  render() {
    return (
      <div>
        <h2>Workbench</h2>
        <h3>Processes</h3>
        {this.props.error
          ? <div className="error">
              Failed to fetch projects: {this.props.error}.
            </div>
          : null}
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
                <td>&nbsp;</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

const workbenchConnected = connect(
  (state: any) => {
    return {
      processes: state.processList.processes,
      error: state.processList.error
    };
  },
  {
    fetchProcessList: fetchProcessList
  }
)(Workbench);

export { workbenchConnected as Workbench };
