import React from 'react'
import { render } from 'react-dom'
import { observable, computed } from 'mobx'
import { observer } from 'mobx-react'
import CSSModules from 'react-css-modules'
import styles from '~/styles/front.css'


class Front {
	@observable number
}

const FrontView = CSSModules(observer(React.createClass({
	render() {
		let number = this.props.front.number
		return (
			<div styleName='root'>
				<div styleName='title'>暴走的金币</div>
				<input type='radio' name='num' ref='p1' value='p1' defaultChecked /><span>1人</span>
				<input type='radio' name='num' ref='p2' value='p2' /><span>2人</span>
				<input type='radio' name='num' ref='p3' value='p3' /><span>3人</span>
				<input type='radio' name='num' ref='p4' value='p4' /><span>4人</span><br/><br/><br/>
				<input type='radio' name='mode' ref='gold' value='gold' defaultChecked /><span>赏金</span>
				<input type='radio' name='mode' ref='survival' value='survival' /><span>生存</span><br/>
				<button styleName='add' onClick={this.addTeam}>取号</button>
				<label>当前号码</label><br/>
				{
					number ? <label ref='currentNumber'>{number}</label> : null
				}
				<button styleName='reset' onClick={this.resetQueue}>重置</button>
			</div>
		)
	},
	addTeam: function(e) {
		let front = this.props.front
		var c = 1
		if (this.refs.p1.checked) {
			c = 1
		} else if (this.refs.p2.checked) {
			c = 2
		} else if (this.refs.p3.checked) {
			c = 3
		} else {
			c = 4
		}
		let param = {
			count: c,
			mode: this.refs.gold.checked ? 'g' : 's'
		}
		$.post('/api/addteam', param, function(data) {
			if (data) {
				front.number = data.id
			}
		})
	},
	resetQueue: function(e) {
		var r = window.confirm('确定要重置吗？')
		if (r == true) {
			let front = this.props.front
			$.post('/api/resetqueue', function(data) {
				front.number = null
			})
		}
	}
})), styles)

var z = require('npm-zepto')
var front = new Front()

render(<FrontView front={front}/>, document.getElementById('front'))
