import { LIST_SUCCESS, LIST_FAILURE } from '../actions/project';
import { Project } from '../../models';

export interface State {
  error?: string;
  item?: Project;
  list?: Array<Project>;
}
export interface Action extends State {
  type: string;
}
export const reducer = (state: State, action: Action): State => {
  state = state || {};

  switch (action.type) {
    case LIST_SUCCESS:
      return { ...state, list: action.list };
    case LIST_FAILURE:
      return { ...state, error: action.error };
  }

  return state;
};
