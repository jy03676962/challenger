import React from 'react'
import { render } from 'react-dom'
import Login from '~/js/login.jsx'
import Hall from '~/js/hall.jsx'
import Arena from '~/js/arena.jsx'
import Board from '~/js/board.jsx'
import Game from '~/js/game.jsx'
import { observer } from 'mobx-react'
import CSSModules from 'react-css-modules'
import styles from '~/styles/base.css'

const App = CSSModules(observer(React.createClass({
  render() {
    var element
    const game = this.props.game
    console.log('game stage is ' + game.stage)
    switch (game.stage) {
      case 'login':
        element = <Login game={game} />
        break
      case 'hall':
        element = <Hall game={game} />
        break
      case 'arena':
        element = <Arena game={game} />
        break
      case 'board':
        element = <Board game={game} />
        break
    }
    let resetStyle = {
      position: 'fixed',
      top: '0',
      right: '0',
      border: '2px solid black',
      cursor: 'pointer',
      zIndex: '1',
    }
    return (
      <div id='app' styleName='base-div'>
        <div onClick={this.reset} style={resetStyle}>重置游戏</div>
        {element}
      </div>
    )
  },
  reset: function(e) {
    this.props.game.resetMatch()
  }
})), styles);

var game = new Game()

document.onkeydown = function(e) {
  game.onKeyDown(e)
};

document.onkeyup = function(e) {
  game.onKeyUp(e)
}

render(<App game={game} />, document.getElementById('root'))
