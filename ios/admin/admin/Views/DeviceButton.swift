//
//  DeviceButton.swift
//  admin
//
//  Created by tassar on 6/26/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class DeviceButton: UIButton {

	var controller: PlayerController? {
		didSet {
			if let c = controller {
				enabled = true
				if c.matchID == 0 {
					setBackgroundImage(UIImage(named: "PCAvailable"), forState: .Normal)
				} else {
					setBackgroundImage(UIImage(named: "PCGaming"), forState: .Normal)
				}
				let id: String = c.address.id
				var title: String?
				if c.address.type == .Simulator {
					title = id[0]
				} else if c.address.type == .Wearable {
					title = id.last()
				}
				setTitle(title, forState: .Normal)
				setTitle(title, forState: .Selected)
			} else {
				enabled = false
				selected = false
				setTitle(nil, forState: .Normal)
			}
		}
	}
}
