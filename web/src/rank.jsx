import React from 'react'
import { render } from 'react-dom'
import { observable, computed } from 'mobx'
import Enum from 'es6-enum'

import SetIntervalMixin from '~/js/mixins/SetIntervalMixin.jsx'
import * as util from '~/js/util.jsx'
import styles from '~/styles/rank.css'

const ChangeModeInterval = 10
const PullDataInterval = 20

class Rank {
	@ observable survivalModeData
	@ observable generalModeData
	@ observable mode
	@ computed get array() {
		if (this.mode == 's') {
			let data = this.survivalModeData
			return [data['1p'], data['2p'], data['3p'], data['4p']]
		} else {
			let g = this.generalModeData.mode0
			let s = this.generalModeData.mode1
			return [s['1p'], s['2p'], s['3p'], s['4p'], g['1p'], g['2p'], g['3p'], g['4p']]
		}
	}
	constructor() {
		this.mode = 'g'
	}
}

const RankCellView = util.cssCreate(styles, {
	render() {
		let idx = this.props.idx
		switch (idx) {
			case 0:
				var iconSrc = require('./assets/icon_1st.png')
				break
			case 1:
				var iconSrc = require('./assets/icon_2nd.png')
				break
			case 2:
				var iconSrc = require('./assets/3rd.png')
				break
			case 3:
				var iconSrc = require('./assets/icon_4.png')
				break
			case 4:
				var iconSrc = require('./assets/icon_5.png')
				break
			case 5:
				var iconSrc = require('./assets/icon_10.png')
				break
			case 6:
				var iconSrc = require('./assets/icon_20.png')
				break
			case 7:
				var iconSrc = require('./assets/icon_50.png')
				break
		}
		let top = (35 + idx * 44) / 19.2 + 'vw'
		let style = {
			position: 'absolute',
			top: top,
			width: '100%',
			height: 41 / 19.2 + 'vw',
			left: '0',
		}
		let nameStr = this.props.data.users.map(user => {
			return user.username
		}).join('...')

		if (this.props.data.time) {
			var result = util.timeStr(this.props.data.time, 2)
		} else {
			var result = this.props.data.gold + 'G'
		}
		let date = new Date(this.props.data['match_time'])
		let year = date.getFullYear() + ''
		let month = '0' + date.getMonth()
		let day = '0' + date.getDay()
		let dateStr = year.substr(-2) + '.' + month.substr(-2) + '.' + day.substr(-2)

		return (
			<div style={style}>
				<img styleName='CellImg' src={iconSrc} />
				<div styleName='CellPlayerLabel'>{nameStr}</div>
				<div styleName='CellResult'>{result}</div>
				<div styleName='CellDate'>{dateStr}</div>
			</div>
		)
	}
})

const SmallRankListView = util.cssCreate(styles, {
	render() {
		let idx = this.props.idx
		let top = (182 + (idx % 4) * 200) / 19.2 + 'vw'
		let left = idx < 4 ? '0' : 896 / 19.2 + 'vw'
		let imgSrc = idx < 4 ? require('./assets/hlist_live.png') : require('./assets/hlist_coin.png')
		switch (idx % 4) {
			case 0:
				var title = '单人排行'
				break
			case 1:
				var title = '双人排行'
				break
			case 2:
				var title = '三人排行'
				break
			case 3:
				var title = '四人排行'
				break
		}
		let style = {
			position: 'absolute',
			top: top,
			height: 167 / 19.2 + 'vw',
			left: left,
			width: 870 / 19.2 + 'vw'
		}
		return (
			<div style={style}>
				<img src={imgSrc} />
				<div styleName='RankListTitleLabel'>{title}</div>
				{
					this.props.data.slice(0, 3).map((m, idx) =>{
						return <RankCellView data={m} idx={idx} key={idx} />
					})
				}
			</div>
		)
	}
})

const LargeRankListView = util.cssCreate(styles, {
	render() {
		let idx = this.props.idx
		if (idx < 2) {
			var top = 182 / 19.2 + 'vw'
		} else {
			var top = 610 / 19.2 + 'vw'
		}
		switch (idx % 4) {
			case 0:
				var title = '单人排行'
				var left = '0'
				break
			case 1:
				var title = '双人排行'
				var left = 896 / 19.2 + 'vw'
				break
			case 2:
				var title = '三人排行'
				var left = '0'
				break
			case 3:
				var title = '四人排行'
				var left = 896 / 19.2 + 'vw'
				break
		}
		let style = {
			position: 'absolute',
			top: top,
			height: 388 / 19.2 + 'vw',
			left: left,
			width: 870 / 19.2 + 'vw'
		}
		return (
			<div style={style}>
				<img src={require('./assets/life_list.png')} />
				<div styleName='RankListTitleLabel'>{title}</div>
				{
					this.props.data.slice(0, 8).map((m, idx) =>{
						return <RankCellView data={m} idx={idx} key={idx} />
					})
				}
			</div>
		)
	}
})

const RankHeaderView = util.cssCreate(styles, {
	render() {
		if (this.props.rank.mode == 's') {
			var imgSrc = require('./assets/top_live.png')
			var titleElement = <div styleName='SeasonLabel'>{this.props.rank.survivalModeData.season}</div>
		} else {
			var imgSrc = require('./assets/top_cl.png')
			var titleElement = null
		}
		let src = this.props.rank.mode == 's' ?
			require('./assets/top_cl.png') : require('./assets/top_live.png')
		return (
			<div styleName='RankHeaderView'>
				<img src={imgSrc} />
				{titleElement}
			</div>
		)
	}
})

const RankDataView = util.cssMobxCreate(styles, {
	render() {
		let sMode = this.props.rank.mode == 's'
		let data = sMode ? this.props.rank.survivalModeData : this.props.rank.generalModeData
		if (data == null || data.error.length != 0) {
			return null
		}

		return (
			<div styleName='RankDataView'>
				<RankHeaderView {...this.props} />
				{
					this.props.rank.array.map((d, idx) => {
						return sMode ? <LargeRankListView data={d} idx={idx} key={idx} /> :
							<SmallRankListView data={d} idx={idx} key={idx} />
						})
				}
			</div>
		)
	}
})

const RankView = util.cssMobxCreate(styles, {
	mixins: [SetIntervalMixin],
	componentDidMount: function() {
		this.pullData()
		this.setInterval(this.changeMode, ChangeModeInterval * 1000)
		this.setInterval(this.pullData, PullDataInterval * 1000)
	},
	changeMode: function() {
		let rank = this.props.rank
		if (rank.mode == 's') {
			rank.mode = 'g'
		} else {
			rank.mode = 's'
		}
	},
	pullData: function() {
		let rank = this.props.rank
		$.get('/api/allhistory', function(data) {
			rank.generalModeData = data
		})
		let now = new Date()
		let param = {
			month: now.getMonth() + 1,
			year: now.getFullYear()
		}
		$.post('/api/mode1history', param, function(data) {
			rank.survivalModeData = data
		})
	},
	render() {
		return (
			<div styleName='root'>
				<div styleName='container'>
					<img styleName='rootImg' src={require('./assets/bg.png')} />
					<RankDataView {...this.props} />
				</div>
			</div>
		)
	},
})

var z = require('npm-zepto')
var rank = new Rank()

render(<RankView rank={rank} />, document.getElementById('rank'))
