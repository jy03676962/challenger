//
//  MatchResultCell.swift
//  admin
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class MatchResultCell: UITableViewCell {
	@IBOutlet weak var idLabel: UILabel!
	@IBOutlet weak var levelLabel: UILabel!
	@IBOutlet weak var gradeLabel: UILabel!
	@IBOutlet weak var goldLabel: UILabel!
	@IBOutlet weak var energyLabel: UILabel!
	@IBOutlet weak var comboLabel: UILabel!
	@IBOutlet weak var energyIcon: UIImageView!

	override func awakeFromNib() {
		super.awakeFromNib()
		backgroundColor = UIColor.clearColor()
	}

	func setData(data: PlayerData?, current: Bool) {
		var c: UIColor
		if current {
			energyIcon.image = UIImage(named: "PowerIcon")
			c = UIColor.whiteColor()
		} else {
			energyIcon.image = UIImage(named: "PowerIconOff")
			c = UIColor(red: 113 / 255.0, green: 146 / 255.0, blue: 191 / 255.0, alpha: 1)
		}
		idLabel.textColor = c
		levelLabel.textColor = c
		gradeLabel.textColor = c
		goldLabel.textColor = c
		energyLabel.textColor = c
		comboLabel.textColor = c
		if let d = data {
			idLabel.text = d.getName()
			if let lvl = d.level {
				levelLabel.text = "LEVEL.\(lvl)"
			} else {
				levelLabel.text = "--"
			}
			gradeLabel.text = d.grade.characters.count > 0 ? d.grade.uppercaseString : "-"
			goldLabel.text = "\(d.gold)/\(d.lostGold)"
			energyLabel.text = "\(Int(d.energy))"
			comboLabel.text = "\(d.combo)"
		} else {
			idLabel.text = "--"
			levelLabel.text = "--"
			gradeLabel.text = "-"
			goldLabel.text = "--"
			energyLabel.text = "--"
			comboLabel.text = "--"
		}
	}
}
