import * as React from 'react';
import { reduxForm, Field } from 'redux-form';

const ProcessForm = (props: any) =>
  <form onSubmit={props.handleSubmit}>
    <label htmlFor="name">Name</label>
    <Field
      name="name"
      component="input"
      type="text"
      placeholder="process name"
    />
  </form>;

const connectedProcessForm = reduxForm({
  form: 'process'
})(ProcessForm);

export { connectedProcessForm as ProcessForm };
