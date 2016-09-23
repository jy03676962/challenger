//
//  HallTableViewCell.swift
//  admin
//
//  Created by tassar on 4/23/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import SwiftyJSON
import SWTableViewCell

class HallTableViewCell: SWTableViewCell {

	@IBOutlet weak var teamIDLabel: UILabel!
	@IBOutlet weak var teamSizeLabel: UILabel!
	@IBOutlet weak var delayCountLabel: UILabel!
	@IBOutlet weak var waitTimeLabel: UILabel!
	@IBOutlet weak var delayCountImageView: UIImageView!
	@IBOutlet weak var numberLabel: UILabel!
	@IBOutlet weak var activeImageView: UIImageView!
	override func awakeFromNib() {
		super.awakeFromNib()
		backgroundColor = UIColor.clearColor()
	}

	override func setSelected(selected: Bool, animated: Bool) {
		super.setSelected(selected, animated: animated)

		// Configure the view for the selected state
	}

	func setData(team: Team, number: Int, active: Bool) {
		teamIDLabel.text = team.id
		teamSizeLabel.text = String(team.size)
		delayCountLabel.text = "- \(team.delayCount) -"
		if team.status == .Prepare {
			waitTimeLabel.text = "准备中..."
		} else if team.status == .Playing {
			waitTimeLabel.text = "游戏中..."
		} else if team.status == .After {
			waitTimeLabel.text = "答题中..."
		} else if team.status == .Waiting {
			waitTimeLabel.text = String(format: "预计等待 %d分钟", team.waitTime / 60)
		}
		let delayImageName = "IconLate\(team.delayCount)"
		let delayImage = UIImage(named: delayImageName)
		delayCountImageView.image = delayImage
		numberLabel.text = String(number + 1)
		activeImageView.hidden = !active
	}
}