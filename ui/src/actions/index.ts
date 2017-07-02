import { Project, Process } from '../models';
import * as db from '../db';

import 'isomorphic-fetch';

export const FETCH_PROJECT_LIST_REQUEST = 'FETCH_PROJECT_LIST_REQUEST';
export const FETCH_PROJECT_LIST_SUCCESS = 'FETCH_PROJECT_LIST_SUCCESS';
export const FETCH_PROJECT_LIST_FAILURE = 'FETCH_PROJECT_LIST_FAILURE';

interface FetchProjectListResponse {
  projects: Array<Project>;
}
export const receiveProjectListSuccess = (projects: Array<Project>) => ({
  type: FETCH_PROJECT_LIST_SUCCESS,
  projects: projects
});
export const receiveProjectListFailure = (error: string) => ({
  type: FETCH_PROJECT_LIST_FAILURE,
  error: error
});
export const fetchProjectList = () => (dispatch: any) =>
  db
    .getProjects()
    .then(res => dispatch(receiveProjectListSuccess(res)))
    .catch(error => dispatch(receiveProjectListFailure(error.message)));

export const FETCH_PROCESS_LIST_REQUEST = 'FETCH_PROCESS_LIST_REQUEST';
export const FETCH_PROCESS_LIST_SUCCESS = 'FETCH_PROCESS_LIST_SUCCESS';
export const FETCH_PROCESS_LIST_FAILURE = 'FETCH_PROCESS_LIST_FAILURE';

interface FetchProcessListResponse {
  processes: Array<Process>;
}
export const receiveProcessListSuccess = (processes: Array<Process>) => ({
  type: FETCH_PROCESS_LIST_SUCCESS,
  processes: processes
});
export const receiveProcessListFailure = (error: string) => ({
  type: FETCH_PROCESS_LIST_FAILURE,
  error: error
});
export const fetchProcessList = (projectID: number) => (dispatch: any) => {
  db
    .getProcesses(projectID)
    .then(res => dispatch(receiveProcessListSuccess(res)))
    .catch(error => dispatch(receiveProcessListFailure(error.message)));
};
