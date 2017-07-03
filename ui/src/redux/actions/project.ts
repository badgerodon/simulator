import * as db from '../../db';
import { Project } from '../../models';

// actions

export const DELETE_REQUEST = 'DELETE_REQUEST';
export const DELETE_SUCCESS = 'DELETE_SUCCESS';
export const DELETE_FAILURE = 'DELETE_FAILURE';
export const GET_REQUEST = 'GET_REQUEST';
export const GET_SUCCESS = 'GET_SUCCESS';
export const GET_FAILURE = 'GET_FAILURE';
export const LIST_REQUEST = 'LIST_REQUEST';
export const LIST_SUCCESS = 'LIST_SUCCESS';
export const LIST_FAILURE = 'LIST_FAILURE';
export const SAVE_REQUEST = 'SAVE_REQUEST';
export const SAVE_SUCCESS = 'SAVE_SUCCESS';
export const SAVE_FAILURE = 'SAVE_FAILURE';

// creators

const deleteRequestCreator = (id: number) => ({
  type: DELETE_REQUEST,
  id: id
});
const deleteSuccessCreator = () => ({
  type: DELETE_SUCCESS
});
const deleteFailureCreator = (error: string) => ({
  type: DELETE_FAILURE,
  error: error
});
const getRequestCreator = (id: number) => ({
  type: GET_REQUEST,
  id: id
});
const getSuccessCreator = (item: Project) => ({
  type: GET_SUCCESS,
  item: item
});
const getFailureCreator = (error: string) => ({
  type: GET_FAILURE,
  error: error
});
const listRequestCreator = () => ({
  type: LIST_REQUEST
});
const listSuccessCreator = (list: Array<Project>) => ({
  type: LIST_SUCCESS,
  list: list
});
const listFailureCreator = (error: string) => ({
  type: LIST_FAILURE,
  error: error
});
const saveRequestCreator = (item: Project) => ({
  type: SAVE_REQUEST,
  item: item
});
const saveSuccessCreator = (item: Project) => ({
  type: SAVE_SUCCESS,
  item: item
});
const saveFailureCreator = (error: string) => ({
  type: SAVE_FAILURE,
  error: error
});

export const deleteProject = (id: number) => (dispatch: any) => {
  dispatch(deleteRequestCreator(id));
  db
    .deleteProject(id)
    .then((res: any) => dispatch(deleteSuccessCreator()))
    .catch(error => dispatch(deleteFailureCreator(error)));
};
export const getProject = (id: number) => (dispatch: any) => {
  dispatch(getRequestCreator(id));
  db
    .getProject(id)
    .then(res => dispatch(getSuccessCreator(res)))
    .catch(error => dispatch(getFailureCreator(error.message)));
};
export const listProjects = () => (dispatch: any) => {
  dispatch(listRequestCreator());
  db
    .getProjects()
    .then(res => dispatch(listSuccessCreator(res)))
    .catch(error => dispatch(listFailureCreator(error.message)));
};
export const saveProject = (item: Project) => (dispatch: any) => {
  dispatch(saveRequestCreator(item));
  db
    .putProject(item)
    .then(res => dispatch(saveSuccessCreator({ ...item, id: res })))
    .catch(error => dispatch(saveFailureCreator(error)));
};
