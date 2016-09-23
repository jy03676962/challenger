//
//  PLViewController.swift
//  admin
//
//  Created by tassar on 4/23/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import EasyPeasy

class PLViewController: UIViewController {
	private var timeLabel: UILabel!
	override func viewDidLoad() {
		super.viewDidLoad()
		let imageView = UIImageView()
		imageView.image = UIImage(named: "GlobalBackground")
		view.insertSubview(imageView, atIndex: 0)
		imageView <- Edges()
		timeLabel = UILabel()
		timeLabel.font = UIFont(name: PLConstants.usualFont, size: 30)
		imageView.addSubview(timeLabel)
		timeLabel <- [
			CenterX(0),
			Top(10)
		]
		NSTimer.scheduledTimerWithTimeInterval(0.5, target: self, selector: #selector(PLViewController.tickTime), userInfo: nil, repeats: true)
	}

	override func viewWillAppear(animated: Bool) {
		super.viewWillAppear(animated)
		timeLabel.textColor = WsClient.singleton.didInit ? UIColor.whiteColor() : UIColor.redColor()
		NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(onWsInited), name: WsClient.WsInitedNotification, object: nil)
		NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(onWsConnecting), name: WsClient.WsConnectingNotification, object: nil)
		NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(onWsDisconnected), name: WsClient.WsDisconnectedNotification, object: nil)
	}

	override func viewDidDisappear(animated: Bool) {
		super.viewDidDisappear(animated)
		NSNotificationCenter.defaultCenter().removeObserver(self)
	}

	func onWsInited() {
		timeLabel.textColor = UIColor.whiteColor()
	}

	func onWsConnecting() {
		timeLabel.textColor = UIColor.greenColor()
	}

	func onWsDisconnected() {
		timeLabel.textColor = UIColor.redColor()
	}

	func tickTime() {
		let now = NSDate()
		let fmt = NSDateFormatter()
		fmt.dateFormat = "HH:mm"
		let str = fmt.stringFromDate(now)
		timeLabel.text = "TIME \(str)"
	}
	override func prefersStatusBarHidden() -> Bool {
		return true
	}
}
