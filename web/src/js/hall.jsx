import React from 'react';
import { observer } from 'mobx-react'

const Hall = observer(React.createClass({
  render: function() {
    return (
      <div>
      <h1>欢迎，{this.props.game.playerName}</h1>
      <MatchView game={this.props.game} />
      </div>
    );
  },
}));

const MatchView = observer(React.createClass({
  render: function() {
    let controllers = this.props.game.controllers
    let mm = controllers.filter((c) => {
      return c.status == 1 && c.address.type == 2
    })
    return (
      <div>
        {
          mm.map((m) =>{
            return <div style={{color:'green'}} key={'player:'+m.address.id}>{m.address.id}</div>
            })
        }
        <button onClick={this.startGoldMode}>开始赏金模式</button>
        <button onClick={this.startSurvivalMode}>开始生存模式</button>
      </div>
    )
  },
  startGoldMode: function(e) {
    this.props.game.startMatch('g')
  },
  startSurvivalMode: function(e) {
    this.props.game.startMatch('s')
  },

}))

export default Hall
