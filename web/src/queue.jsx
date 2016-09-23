import React from 'react'
import { render } from 'react-dom'
import { observable, computed } from 'mobx'
import { observer } from 'mobx-react'
import CSSModules from 'react-css-modules'
import styles from '~/styles/queue.css'
import * as util from '~/js/util.jsx'

class Queue {
	@ observable data
	@ observable connected
	@ observable match
	constructor() {
		this._reset()
	}

	_reset() {
		this.sock = null
		this.state = false
		this.data = null
		this.match = null
	}

	connect() {
		if (this.sock) {
			this.sock.close()
			this._reset()
			return
		}
		let uri = util.wsAddressWithPath('ws')
		let sock = new WebSocket(uri)
		sock.onopen = () => {
			let data = {
				cmd: 'init',
				ID: 'queue',
				TYPE: '8',
			}
			sock.send(JSON.stringify(data))
		}
		sock.onclose = (e) => {
			this._reset()
		}
		sock.onmessage = (e) => {
			this.onMessage(e.data)
		}
		this.sock = sock
	}

	onMessage(msg) {
		let json = JSON.parse(msg)
		switch (json.cmd) {
			case 'init':
				this.connected = true
				break
			case 'matchData':
				if (json.data != null && this.connected) {
					this.data = json.data
				}
				break
			case 'matchStop':
				this.match = null
			case 'updateMatch':
				if (json.data != null && this.connected) {
					this.match = JSON.parse(json.data)
				}
		}
	}

	send(data) {
		if (this.sock) {
			let d = JSON.stringify(data)
			this.sock.send(d)
		}
	}
}

const HistoryCellView = CSSModules(React.createClass({
	render() {
		let matchData = this.props.data
		let idx = this.props.idx
		let top = (idx * 58) / 10.8 + 'vw'
		if (matchData.mode == 'g') {
			var modeImg = require('./assets/g_icon.png')
			var resultTitle = '获得:'
			var result = `${matchData.gold}[G]`
		} else {
			var modeImg = require('./assets/s_icon.png')
			var resultTitle = '生存:'
			var result = `${matchData.elasped.toFixed(2)}[S]`
		}
		let playerStr = matchData.member.map((player, idx) => {
			if (player.name) {
				return player.name
			} else {
				return util.playerStr(player.cid)
			}
		}).join(' ')
		let style = {
			position: 'absolute',
			width: '100%',
			height: '5.185vw',
			top: top,
		}
		return (
			<div style={style}>
				<DelayView count={0} num={0} left={'7.87vw'} top={'0.926vw'} />
				<div styleName='historyNumber'>{matchData.teamID}</div>
				<div styleName='historyPlayer'>{playerStr}</div>
				<img src={modeImg} styleName='historyIcon' />
				<div styleName='historyResultTitle'>{resultTitle}</div>
				<div styleName='historyResult'>{result}</div>
			</div>
		)
	}
}), styles)

const HistoryView = CSSModules(React.createClass({
	render() {
		let history = this.props.history
		if (history == null || history.length == 0) {
			return null
		}
		return (
			<div styleName='history'>
        {
        history.sort((a, b) => {
          return a.id > b.id
        }).map((m, idx) => {
          return <HistoryCellView data={m} idx={idx} key={idx}/>
        })
        }
      </div>
		)
	}
}), styles)

const CurrentMatchView = CSSModules(React.createClass({
	render() {
		let match = this.props.match
		if (match == null) {
			return null
		}
		let bg = match.mode == 'g' ? require('./assets/g_game_bg.png') : require('./assets/s_game_bg.png')
		let color = match.mode == 'g' ? '#dc8524' : '#03dceb'
		let time = util.timeStr(match.elasped)
		return (
			<div styleName='matchInfo'>
        <img src={bg}/>
        <div style={{color:color}}>
          <div styleName='matchTimeLabel'>游戏已开始：</div>
          <div styleName='matchGoldLabel'>当前金币数：</div>
          <div styleName='matchTimeValue'>{time}</div>
          <div styleName='matchGoldValue'>{match.gold + 'G'}</div>
        </div>
      </div>
		)
	}
}), styles)

const CurrentMatchCellView = CSSModules(React.createClass({
	render() {
		let match = this.props.match
		if (match == null) {
			return null
		}
		if (match.mode == 'g') {
			var bgImg = require('./assets/g_cell_bg.png')
			var modeImg = require('./assets/g_icon.png')
			var color = '#dc8524'
		} else {
			var bgImg = require('./assets/s_cell_bg.png')
			var modeImg = require('./assets/s_icon.png')
			var color = '#03dceb'
		}
		let playerStr = match.member.map((player, idx) => {
			return util.playerStr(player.cid)
		}).join(' ')
		return (
			<div styleName='matchCell'>
				<img src={bgImg}/>
				<div style={{color:color}}>
					<DelayView count={0} num={0} left={'7.87vw'} top={'1.99vw'} />
					<div styleName='currentNumber'>{match.teamID}</div>
					<div styleName='currentPlayer'>{playerStr}</div>
					<img src={modeImg} styleName='currentIcon' />
					<div styleName='currentResultTitle'>进行中...</div>
				</div>
			</div>
		)
	}
}), styles)


const DelayView = CSSModules(React.createClass({
	render() {
		let delayImg = (delay) => {
			switch (delay) {
				case 0:
					return require('./assets/late0.png')
				case 1:
					return require('./assets/late1.png')
				case 2:
					return require('./assets/late2.png')
				case 3:
					return require('./assets/late3.png')
				case 4:
					return require('./assets/late4.png')
			}
		}
		let count = this.props.count
		let style = {
			position: 'absolute',
			left: this.props.left,
			top: this.props.top,
			width: '3.333vw',
			height: '3.333vw',
		}
		return (
			<div style={style}>
				<img src={delayImg(count)} styleName='delayImg'/>
				<div styleName='delayText'>{this.props.num == 0 ? '--' : this.props.num + ''}</div>
			</div>
		)
	}
}), styles)

const PrepareCellView = CSSModules(React.createClass({
	render() {
		let team = this.props.team
		let empty = util.isEmpty(team)
		if (team == null) {
			return null
		}
		let bg = team.mode == 'g' ? require('./assets/g_p_bg.png') : require('./assets/s_p_bg.png')
		return (
			<div styleName='prepareCell'>
				<DelayView count={0} num={0} left={'8.333vw'} top={'3.7vw'} />
				<div styleName='prepareNumber'>{team.id}</div>
				<img src={bg} styleName='prepareBg'/>
				<div styleName='prepareText'>进入等待区...</div>
			</div>
		)
	}
}), styles)

const WaitingCellView = CSSModules(React.createClass({
	render() {
		let team = this.props.team
		let style = {
			position: 'absolute',
			left: this.props.left,
			top: this.props.top,
			height: '5.37vw',
			width: '45.37vw',
			color: '#89b2e8',
		}
		return (
			<div style={style}>
        <DelayView count={team ? team.delayCount : 0} num={this.props.num} left={'3.7vw'} top={'1.02vw'} />
        <div styleName='waitingNumber'>{team ? team.id : '--'}</div>
        <div styleName='waitingText'>{team ? '预计等待：' : '--'}</div>
        <div styleName='waitingTime'>{team ? this.props.t: '--'}</div>
      </div>
		)
	}
}), styles)

const QueueView = CSSModules(observer(React.createClass({
	render() {
		if (this.props.queue.data == null) {
			return (
				<div styleName='root'>
        <div styleName='container'>
          <img styleName='rootImg' src={require('./assets/qbg.png')} />
        </div>
      </div>
			)
		}
		let history = this.props.queue.data.history
		let queue = this.props.queue.data.queue
		let match = this.props.queue.match
		var count = queue.length
		let preparing = null
		let callingTeam = null
		let waiting = []
		for (let team of queue) {
			if (team.status != 0) {
				count--
				if (team.status == 1) {
					preparing = team
				}
			} else {
				if (team.calling > 0) {
					callingTeam = team
				}
				break
			}
		}
		for (let team of queue) {
			if (team.status == 0) {
				waiting.push(team)
			}
		}
		let ar = []
		for (var i = 0; i < 20; i++) {
			ar.push(i)
		}
		return (
			<div styleName='root'>
        <div styleName='container'>
          <img styleName='rootImg' src={require('./assets/qbg.png')} />
          <div styleName='timeLabel'>最长等待时间：</div>
          <div styleName='groupLabel'>当前排队组数：</div>
          <div styleName='timeValue'>{count * 5}</div>
          <div styleName='groupValue'>{count}</div>
          <div styleName='timeUnit'>分钟</div>
          <div styleName='groupUnit'>组</div>
          <CurrentMatchView match={match} />
          <HistoryView history={history} />
          <CurrentMatchCellView match={match} />
          <PrepareCellView team={preparing} />
          {
			  ar.map(i => {
				  if (i < waiting.length) {
					  var team = waiting[i]
					  var num = i + 1
				  } else {
					  var team = null
					  var num = 0
				  }
				  let t = 5 *(i + 1) + '分钟'
				  if (i < 10) {
					  let left = '4.259vw'
					  let top = (938 + i * 60) / 10.8 + 'vw'
					  return <WaitingCellView team={team} num={num} left={left} top={top} t={t} key={i} />
					  } else if (i < 20) {
						  let left = 544/10.8 + 'vw'
						  let top = (938 + (i -10) * 60) / 10.8 + 'vw'
						  return <WaitingCellView team={team} num={num} left={left} top={top} t={t} key={i} />
						  }
			  })
          }
		  <img styleName='log' src={require('./assets/queue_log.png')} />
		  <img styleName='wechat' src={require('./assets/wechat.png')} />
		  {
			  callingTeam ?  <img styleName='callingImg' src={require('./assets/c_bg.png')} /> : null
		  }
		  {
			  callingTeam ?  <div styleName='callingText'>{callingTeam.id}</div> : null
		  }
        </div>
      </div>
		)
	},
	componentDidMount() {
		this.props.queue.connect()
	}
})), styles)

var queue = new Queue()

render(<QueueView queue={queue} />, document.getElementById('queue'))
