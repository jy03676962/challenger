import React from 'react';
import { observer } from 'mobx-react'
import Player from '~/js/player.jsx'
import Laser from '~/js/laser.jsx'
import Scheme from '~/js/scheme.jsx'
import styles from '~/styles/info.css'
import CSSModules from 'react-css-modules'

const Arena = observer(React.createClass({
  render() {
    let opt = this.props.game.options
    let arenaWidth = (opt.arenaCellSize + opt.arenaBorder) * opt.arenaWidth * opt.webScale
    let arenaHeight = (opt.arenaCellSize + opt.arenaBorder) * opt.arenaHeight * opt.webScale
    const infoHeight = 60
    let infoStyle = {
      marginTop: '10px',
      width: arenaWidth + opt.arenaBorder * opt.webScale + 'px',
      height: infoHeight + 'px',
      marginLeft: 'auto',
      marginRight: 'auto',
      textAlign: 'center',
      backgroundColor: '#CCCCCC'
    }
    let bgStyle = {
      width: arenaWidth + 'px',
      fontSize: '0',
      margin: 'auto',
      position: 'relative',
      border: opt.arenaBorder / 2 * opt.webScale + 'px solid ' + Scheme.wall
    }
    let gStyle = {
      position: 'absolute',
      top: infoHeight + 'px',
      left: '0',
      right: '0',
      margin: 'auto',
      width: arenaWidth + 'px',
      height: arenaHeight + 'px',
      border: opt.arenaBorder / 2 * opt.webScale + 'px solid ' + Scheme.wall
    }
    return (
      <div style={{position:'relative'}}>
      <ArenaInfoBar game={this.props.game} rootStyle={infoStyle}/>
      <ArenaBackground opt={opt} rootStyle={bgStyle} />
      <ArenaButtonLayer game={this.props.game} rootStyle={gStyle} />
      <ArenaGround game={this.props.game} rootStyle={gStyle} />
      </div>
    )
  }
}))

const ArenaInfoBar = CSSModules(observer(React.createClass({
  render() {
    let game = this.props.game
    let opt = game.options
    let match = game.match
    let msg, timeText
    let player = null
    for (let p of game.match.member) {
      if (p.cid == game.playerName) {
        player = p
      }
    }
    let combo = player ? player.combo : 0
    if (match.stage.startsWith('warmup')) {
      let left = match.warmupTime.toFixed(1)
      msg = '预热阶段'
      timeText = `还剩${left}秒`
    } else if (match.rampageTime > 0) {
      msg = '暴走阶段'
      let left = match.rampageTime.toFixed(1)
      timeText = `还剩${left}秒`
    } else {
      msg = '游戏阶段'
      if (match.mode == 'g') {
        let left = (match.totalTime).toFixed(1)
        timeText = `还剩${left}秒`
      }
    }
    let gold = match.gold.toFixed(0)
    let goldText = `当前金币: ${gold}`
    let p = (match.energy / opt.maxEnergy) * 100 + '%'
    let energyText = `能量(${match.energy.toFixed(1)}/${opt.maxEnergy}):`
    let energyTextColor, energyBarColor
    if (match.rampageTime > 0) {
      p = '100%'
      energyTextColor = 'red'
      energyBarColor = 'red'
    } else {
      energyTextColor = 'black'
      energyBarColor = '#66CC00'
    }
    return (
      <div style={this.props.rootStyle}>
      <div styleName='leftBar'>
      <div styleName='message'>{msg}</div>
      <div styleName='timer'>{timeText}</div>
      </div>
      <div styleName='centerBar'>{`当前连击${combo}`}</div>
      <div styleName='rightBar'>
      <div styleName='gold'>{goldText}</div>
      <div styleName='energyBg'>
      <div styleName='energy' style={{width:p, backgroundColor:energyBarColor}}></div>
      </div>
      <div styleName='energyText' style={{color:energyTextColor}}>{energyText}</div>
      </div>
      </div>
    )
  }
})), styles)

const ArenaBackground = ({ opt, rootStyle }) => {
  let size = opt.arenaCellSize * opt.webScale + 'px'
  let elements = []
  for (let i = 0; i < opt.arenaHeight; i++) {
    for (let j = 0; j < opt.arenaWidth; j++) {
      let cellStyle = {
        width: size,
        height: size,
        display: 'inline-block',
        border: opt.arenaBorder / 2 * opt.webScale + 'px solid ' + Scheme.border,
        backgroundColor: Scheme.normalTile
      }
      if (j == opt.arenaEntrance.X && i == opt.arenaEntrance.Y) {
        cellStyle.backgroundColor = Scheme.entranceTile
      } else if (j == opt.arenaExit.X && i == opt.arenaExit.Y) {
        cellStyle.backgroundColor = Scheme.exitTile
      } else {
        cellStyle.backgroundColor = Scheme.normalTile
      }
      elements.push(<div style={cellStyle} key={'cell:'+i * opt.arenaWidth + j}></div>)
    }
  }
  for (let [index, wall] of opt.walls.entries()) {
    let wallStyle = {
      position: 'absolute',
      backgroundColor: Scheme.wall,
      left: wall.X * opt.webScale + 'px',
      top: wall.Y * opt.webScale + 'px',
      width: wall.W * opt.webScale + 'px',
      height: wall.H * opt.webScale + 'px',
    }
    elements.push(<div style={wallStyle} key={'wall:' + index}></div>)
  }
  return (
    <div style={rootStyle}>
  {elements}
  </div>
  );
}

const ArenaButtonLayer = observer(React.createClass({
  render() {
    let match = this.props.game.match
    let opt = this.props.game.options
    return (
      <div style={this.props.rootStyle}>
      {
        opt.buttons.filter((button) =>{
          return match.onButtons && (button.id in match.onButtons)
        }).map((button) => {
          let color = match.rampageTime > 0 ? Scheme.buttonRampage : Scheme.buttonInit
          let border = null
          for (let player of match.member) {
            if (player.button == button.id) {
              border = `2px solid ${Scheme.buttonPressing}`
              if (match.rampageTime > 0) {
                color = Scheme.buttonRampageLevel[player.buttonLevel]
              } else {
                color = Scheme.buttonLevel[player.buttonLevel]
              }
            }
          }
          let r = button.r
          let buttonStyle = {
            position: 'absolute',
            border: border,
            backgroundColor: color,
            left: r.X * opt.webScale + 'px',
            top: r.Y * opt.webScale + 'px',
            width: r.W * opt.webScale + 'px',
            height: r.H * opt.webScale + 'px',
          }
          return <div style={buttonStyle} key={'button:' + button.id}></div>
        })
      }
      </div>
    )
  }
}))

const ArenaGround = observer(React.createClass({
  render() {
    let match = this.props.game.match
    let opt = this.props.game.options
    let objects = []
    match.member.forEach(function(member, idx) {
      objects.push(<Player player={member} options={opt} key={'player:'+idx} idx={idx}/>)
    })
    if (match.lasers) {
      match.lasers.forEach(function(laser, idx) {
        objects.push(<Laser laser={laser} options={opt} key={'laser:'+idx}/>)
      })
    }
    return (
      <div style={this.props.rootStyle}>
      {objects}
      </div>
    )
  }
}))

export default Arena
