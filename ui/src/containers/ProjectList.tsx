import * as React from 'react';
import { RouterState, Link } from 'react-router';
import { connect } from 'react-redux';

import { fetchProjectList } from '../actions';
import { Project } from '../models';

interface ProjectListProps {
  projects: Array<Project>;
  error: string;

  fetchProjectList: typeof fetchProjectList;
}

class ProjectList extends React.Component<ProjectListProps, {}> {
  componentWillMount() {
    this.props.fetchProjectList();
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
  state => {
    return {
      projects: state.projectList.projects,
      error: state.projectList.error
    };
  },
  {
    fetchProjectList: fetchProjectList
  }
)(ProjectList);

export { connected as ProjectList };
