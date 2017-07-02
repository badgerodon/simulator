import {
  FETCH_PROJECT_LIST_FAILURE,
  FETCH_PROJECT_LIST_SUCCESS,
  FETCH_PROCESS_LIST_FAILURE,
  FETCH_PROCESS_LIST_SUCCESS
} from '../actions';
import { Project, Process } from '../models';

interface ProjectListState {
  projects: Array<Project>;
  error?: string;
}
interface ProjectListAction {
  type: string;
  error?: string;
  projects?: Array<Project>;
}
export const projectListReducer = (
  state: ProjectListState,
  action: ProjectListAction
): ProjectListState => {
  state = state || {
    projects: [],
    error: null
  };

  switch (action.type) {
    case FETCH_PROJECT_LIST_SUCCESS:
      return {
        projects: action.projects
      };
    case FETCH_PROJECT_LIST_FAILURE:
      return {
        projects: [],
        error: action.error
      };
  }
  return state;
};

interface ProcessListState {
  processes: Array<Process>;
  error?: string;
}
interface ProcessListAction {
  type: string;
  error?: string;
  processes?: Array<Process>;
}
export const processListReducer = (
  state: ProcessListState,
  action: ProcessListAction
): ProcessListState => {
  state = state || {
    processes: [],
    error: null
  };

  console.log('ACTION!', action);

  switch (action.type) {
    case FETCH_PROCESS_LIST_SUCCESS:
      return {
        processes: action.processes
      };
    case FETCH_PROCESS_LIST_FAILURE:
      return {
        processes: [],
        error: action.error
      };
  }
  return state;
};
