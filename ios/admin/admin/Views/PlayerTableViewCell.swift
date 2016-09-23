//
//  PlayerTableViewCell.swift
//  admin
//
//  Created by tassar on 5/3/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class PlayerTableViewCell: UITableViewCell {
	@IBOutlet weak var playerIDLabel: UILabel!
	@IBOutlet weak var goldLabel: UILabel!
	@IBOutlet weak var energyLabel: UILabel!
	@IBOutlet weak var comboLabel: UILabel!

	override func awakeFromNib() {
		super.awakeFromNib()
		backgroundColor = UIColor.clearColor()
	}

	func setData(player: Player) {
		playerIDLabel.text = player.displayName
		goldLabel.text = "\(player.gold)/\(player.lostGold)"
		energyLabel.text = String(format: "%.f", player.energy)
		comboLabel.text = "\(player.combo)"
	}
}
