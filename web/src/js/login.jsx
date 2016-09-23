import React from 'react';
import { observer } from 'mobx-react'

const Login = observer(React.createClass({
  render: function() {
    return (
      <div>
      <h1>登陆</h1>
      <form onSubmit={this.onSubmit}>
      <label>昵称</label>
      <input type='text' ref='userName'/><br/><br/>
      </form>
      <button onClick={this.onSubmit}>登陆</button>
      </div>
    );
  },
  onSubmit: function(e) {
    this.props.game.connectServer(this.refs.userName.value)
  },
}));

export default Login;
