import * as React from 'react';
import { RouterState } from 'react-router';

class Root extends React.Component<RouterState, {}> {
  render() {
    return (
      <div>
        <h1>gRPC Simulator</h1>
        {this.props.children}
      </div>
    );
  }
}

export { Root };
