import {
  GET_SUCCESS,
  GET_FAILURE,
  SAVE_SUCCESS,
  SAVE_FAILURE
} from '../actions/process';
import { Process } from '../../models';

export interface State {
  error?: string;
  item?: Process;
  list?: Array<Process>;
}
export interface Action extends State {
  type: string;
}
export const reducer = (state: State, action: Action): State => {
  state = state || {};

  switch (action.type) {
    case GET_SUCCESS:
      return { ...state, item: action.item };
    case GET_FAILURE:
      return { ...state, error: action.error };
    case SAVE_SUCCESS:
      return { ...state, item: action.item };
    case SAVE_FAILURE:
      return { ...state, error: action.error };
  }

  return state;
};
