import * as React from 'react';
import { RouterState, Link } from 'react-router';
import { connect } from 'react-redux';

import { UIState } from '../redux/state';
import { listProjects } from '../redux/actions/project';
import { Project } from '../models';

interface State {
  projects: Array<Project>;
  error: string;
}
interface Props extends RouterState, State {
  listProjects: typeof listProjects;
}

class ProjectList extends React.Component<Props> {
  componentWillMount() {
    this.props.listProjects();
  }

  render() {
    return (
      <div>
        <h2>Projects</h2>
        {this.props.error
          ? <div className="error">
              Failed to fetch projects: {this.props.error}.
            </div>
          : null}
        <table>
          <thead>
            <tr>
              <th>Name</th>
            </tr>
          </thead>
          <tbody>
            {this.props.projects.map(project =>
              <tr key={project.id}>
                <td>
                  <Link to={'/projects/' + project.id}>
                    {project.name}
                  </Link>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

const connected = connect(
  (state: UIState): State => {
    return {
      projects: state.project.list || [],
      error: state.project.error
    };
  },
  {
    listProjects: listProjects
  }
)(ProjectList);

export { connected as ProjectList };
