import { observable, computed } from 'mobx'
import { wsAddressWithPath } from '~/js/util.jsx'

class Game {
  @observable match
  @observable playerName
  @observable options
  @observable controllers
  @observable matchData

  @computed get stage() {
    if (!this.playerName) {
      return 'login'
    }
    if (!this.match && !this.matchData) {
      return 'hall'
    }
    if (this.matchID > 0 && this.match != null) {
      return 'arena'
    }
    if (this.matchData != null) {
      return 'board'
    }
  }

  constructor() {
    this._reset()
  }

  _reset() {
    this.playerName = ''
    this.sock = null
    this.match = null
    this.arg = null
    this.options = null
    this.controllers = null
    this.matchData = null
    this.currentKey = 0
    this.matchID = 0
  }

  connectServer(playerName) {
    let uri = wsAddressWithPath('ws')
    let sock = new WebSocket(uri)
    console.log('socket is ' + uri)
    sock.onopen = () => {
      console.log('connected to ' + uri)
      this.login(playerName)
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

  login(playerName) {
    if (this.sock) {
      let data = {
        cmd: 'init',
        ID: playerName,
        TYPE: '2'
      }
      this.sock.send(JSON.stringify(data))
    }
  }

  startMatch(mode) {
    if (this.sock) {
      let data = {
        cmd: 'startMatch',
        mode: mode,
      }
      this.sock.send(JSON.stringify(data))
    }
  }

  resetMatch() {
    if (this.matchID > 0) {
      if (this.sock) {
        let data = {
          cmd: 'stopMatch',
          matchID: this.matchID,
        }
        this.sock.send(JSON.stringify(data))
      }
    } else if (this.matchData != null) {
      this.matchData = null
    }
  }

  onMessage(msg) {
    let json = JSON.parse(msg)
    console.log(msg)
    switch (json.cmd) {
      case 'init':
        this.options = json.data.options
        console.log(json.data.options)
        this.playerName = json.data.ID
        break
      case 'updateMatch':
        let match = JSON.parse(json.data)
        if (match.id == this.matchID) {
          this.match = match
        }
        break
      case 'ControllerData':
        this.controllers = json.data
        break
      case 'newMatch':
        this.matchID = json.data
        break
      case 'matchStop':
        if (this.matchID == json.data.matchID) {
          this.matchData = json.data.matchData
          this.matchID = 0
          this.match = null
        }
    }
  }

  onKeyDown(e) {
    if (this.stage != 'arena') {
      return
    }
    var code = e.keyCode ? e.keyCode : e.which;
    let dir
    switch (code) {
      case 37: //left
        dir = 'left'
        break
      case 38:
        dir = 'up'
        break
      case 39:
        dir = 'right'
        break
      case 40:
        dir = 'down'
        break
    }
    if (dir) {
      this.currentKey = code
      let data = {
        cmd: 'playerMove',
        dir: dir,
        matchID: this.matchID,
      }
      this.sock.send(JSON.stringify(data))
    }
  }

  onKeyUp(e) {
    if (this.stage != 'arena') {
      return
    }
    var code = e.keyCode ? e.keyCode : e.which;
    if (this.currentKey == code) {
      this.currentKey = 0
      let data = {
        cmd: 'playerStop',
        matchID: this.matchID,
      }
      this.sock.send(JSON.stringify(data))
    }
  }
}


export default Game
