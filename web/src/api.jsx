import React from 'react'
import { render } from 'react-dom'
import { observable, computed } from 'mobx'
import { observer } from 'mobx-react'
import CSSModules from 'react-css-modules'
import styles from '~/styles/api.css'
import { wsAddressWithPath } from '~/js/util.jsx'

class Api {
	@ observable state
	@ observable addr
	@ observable log
	constructor() {
		this._reset()
	}

	_reset() {
		this.addr = ''
		this.output = ''
		this.state = 'not connected'
		this.sock = null
		this.log = ''
	}
	connect() {
		if (this.sock) {
			this.sock.close()
			this._reset()
			return
		}
		let uri = wsAddressWithPath('ws')
		let sock = new WebSocket(uri)
		this.state = 'connecting...'
		sock.onopen = () => {
			console.log('connected to ' + uri)
			let data = {
				cmd: 'init',
				ID: 'api_test',
				TYPE: '3',
			}
			sock.send(JSON.stringify(data))
		}
		sock.onclose = (e) => {
			console.log('connection closed (' + e.code + ')')
			this._reset()
		}
		sock.onmessage = (e) => {
			this.onMessage(e.data)
		}
		this.sock = sock
	}
	onMessage(msg) {
		console.log('got socket message: ' + msg)
		let data = JSON.parse(msg)
		switch (data.cmd) {
			case 'init':
				this.state = 'connected'
			case 'addTCP':
				this.addr = data.data
				//this.log = `新连接: ${this.addr.id}\n` + this.log
				break
			case 'removeTCP':
				this._reset()
				break
			default:
				this.addr = data.addr
				//this.log = `收到:${msg}\n` + this.log
		}
	}
	send(data) {
		let d = JSON.stringify(data)
		this.log = `发送:${d}\n` + this.log
		this.sock.send(d)
	}
}

const ApiView = CSSModules(observer(React.createClass({
	render() {
		let c = this.props.api.state == 'connected' ? '断开' : '连接'
		let a = this.props.api.addr == null ? '' : this.props.api.addr.id
		return (
			<div styleName='root'>
				<div styleName='left'>
					<div styleName='block'>
						<label styleName='content'>地址</label>
						<label>{a}</label>
						<label styleName='content'>状态</label>
						<label>{this.props.api.state}</label>
						<button onClick={this.connect}>{c}</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>ArduinoID</label>
						<input type='text' ref='addr'></input><br/>
					</div>
					<div styleName='block'>
						<label styleName='title'>灯带效果</label><br/>
						<label styleName='content'>wall</label>
						<input type='text' ref='wall'></input><br/>
						<label styleName='content'>led_t</label>
						<input type='text' ref='led_t'></input><br/>
						<label styleName='content'>mode</label>
						<input type='text' ref='mm'></input><br/>
						<button onClick={this.ledCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>激光控制</label><br/>
						<label styleName='content'>laser_n</label>
						<input type='text' ref='laser_n'></input><br/>
						<label styleName='content'>laser_s</label>
						<input type='text' ref='laser_s'></input><br/>
						<button onClick={this.laserCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>按键</label><br/>
						<label styleName='content'>useful</label>
						<input type='text' ref='useful'></input><br/>
						<label styleName='content'>mode</label>
						<input type='text' ref='btn_mode'></input><br/>
						<label styleName='content'>stage</label>
						<input type='text' ref='btn_stage'></input><br/>
						<button onClick={this.btnCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>播放音乐</label>
						<input type='text' ref='music'></input><br/>
						<button onClick={this.musicCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>指示灯</label>
						<input type='text' ref='light'></input><br/>
						<button onClick={this.lightCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>模式选择</label>
						<input type='text' ref='mode'></input><br/>
						<button onClick={this.modeCtrl}>发送</button>
					</div>
					<div styleName='block'>
						<label styleName='title'>分数设置</label><br/>
						<label styleName='content'>T1</label>
						<input type='text' ref='t1'></input><br/>
						<label styleName='content'>T2</label>
						<input type='text' ref='t2'></input><br/>
						<label styleName='content'>T3</label>
						<input type='text' ref='t3'></input><br/>
						<label styleName='content'>暴走</label>
						<input type='text' ref='tr'></input><br/>
						<button onClick={this.scoreCtrl}>发送</button>
					</div>
				</div>
				<div styleName='right'>
					<textarea ref='console' readOnly value={this.props.api.log}></textarea>
				</div>
			</div>
		)
	},
	connect: function(e) {
		this.props.api.connect()

	},
	ledCtrl: function(e) {
		let d = {
			cmd: 'led_ctrl',
			led: [{
				wall: this.refs.wall.value,
				led_t: this.refs.led_t.value,
				mode: this.refs.mm.value
			}],
			addr: this.refs.addr.value,
		}
		this.props.api.send(d)
	},
	btnCtrl: function(e) {
		let d = {
			cmd: 'btn_ctrl',
			useful: this.refs.useful.value,
			mode: this.refs.btn_mode.value,
			stage: this.refs.btn_stage.value,
			addr: this.refs.addr.value,
		}
		this.props.api.send(d)
	},
	musicCtrl: function(e) {
		let d = {
			cmd: 'mp3_ctrl',
			addr: this.refs.addr.value,
			music: this.refs.music.value
		}
		this.props.api.send(d)
	},
	lightCtrl: function(e) {
		let d = {
			cmd: 'light_ctrl',
			light_mode: this.refs.light.value,
			addr: this.refs.addr.value,
		}
		this.props.api.send(d)
	},
	modeCtrl: function(e) {
		let d = {
			cmd: 'mode_change',
			mode: this.refs.mode.value,
			addr: this.refs.addr.value,
		}
		this.props.api.send(d)
	},
	scoreCtrl: function(e) {
		let d = {
			cmd: 'init_score',
			addr: this.refs.addr.value,
			score: [{
				status: 'T1',
				'time': this.refs.t1.value
			}, {
				status: 'T2',
				'time': this.refs.t2.value
			}, {
				status: 'T3',
				'time': this.refs.t3.value
			}, {
				status: 'TR',
				'time': this.refs.tr.value
			}, ]
		}
		this.props.api.send(d)
	},
	laserCtrl: function(e) {
		let d = {
			cmd: 'laser_ctrl',
			addr: this.refs.addr.value,
			laser: [{
				laser_n: this.refs.laser_n.value,
				laser_s: this.refs.laser_s.value,
			}]
		}
		this.props.api.send(d)
	}
})), styles)

var api = new Api()

render(<ApiView api={api} />, document.getElementById('api'))
