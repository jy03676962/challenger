//
//  HistoryTableViewCell.swift
//  admin
//
//  Created by tassar on 5/6/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class HistoryTableViewCell: UITableViewCell {
	@IBOutlet weak var teamIDLabel: UILabel!
	@IBOutlet weak var playerCountLabel: UILabel!
	@IBOutlet weak var statusLabel: UILabel!
	@IBOutlet weak var backgroundImageView: UIImageView!
	override func awakeFromNib() {
		super.awakeFromNib()
		backgroundColor = UIColor.clearColor()
	}
	func setData(data: MatchData) {
		teamIDLabel.text = data.teamID
		playerCountLabel.text = "\(data.member.count)"
		var txt: String
		switch data.answerType! {
		case .NotAnswer:
			txt = "尚未答题"
		case .Answering:
			txt = "答题中"
		case .Answered:
			txt = "已答题"
		}
		statusLabel.text = txt
	}
	override func setSelected(selected: Bool, animated: Bool) {
		backgroundImageView.hidden = !selected
	}
}
