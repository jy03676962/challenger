import React from 'react';
import { observer } from 'mobx-react'
import Scheme from '~/js/scheme.jsx'

const Laser = observer(React.createClass({
  render() {
    let laser = this.props.laser
    let opt = this.props.options
    let scale = opt.webScale
    let size = opt.arenaCellSize * scale
    let style = {
      position: 'absolute',
      width: size + 'px',
      height: size + 'px',
      top: laser.pos.Y * scale - size / 2 + 'px',
      left: laser.pos.X * scale - size / 2 + 'px',
      backgroundColor: laser.isPause ? Scheme.laserPause : Scheme.laserNormal
    }
    return (
      <div style={style}></div>
    );
  }
}))

export default Laser
