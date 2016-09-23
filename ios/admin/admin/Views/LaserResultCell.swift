//
//  LaserResultCell.swift
//  admin
//
//  Created by tassar on 5/21/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class LaserResultCell: UITableViewCell {

	@IBOutlet weak var idLabel: UILabel!
	@IBOutlet weak var urLabel: UILabel!
	@IBOutlet weak var errLabel: UILabel!

	func renderData(data: LaserInfo) {
		idLabel.text = data.id
		urLabel.text = data.ur
		if data.err == 1 {
			errLabel.text = "当前设备未连接"
		} else if data.err == 2 {
			errLabel.text = "有多个反馈"
		} else {
			errLabel.text = ""
		}
	}
}
