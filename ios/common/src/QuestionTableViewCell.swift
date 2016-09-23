//
//  QuestionTableViewCell.swift
//  postgame
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class QuestionTableViewCell: UITableViewCell {

	@IBOutlet weak var iconImageView: UIImageView!
	@IBOutlet weak var contentLabel: UILabel!
	override func awakeFromNib() {
		super.awakeFromNib()
		backgroundColor = UIColor.clearColor()
		selectionStyle = .None
	}

	override func setSelected(selected: Bool, animated: Bool) {
		super.setSelected(selected, animated: animated)
		iconImageView.image = selected ? UIImage(named: "OptionOn") : UIImage(named: "OptionOff")
	}

	func setData(data: String?) {
		contentLabel.text = data
	}
}
