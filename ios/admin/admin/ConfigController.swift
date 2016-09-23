//
//  ConfigController.swift
//  admin
//
//  Created by tassar on 4/23/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import AutoKeyboardScrollView
import EasyPeasy
import SwiftyUserDefaults
import SwiftyJSON
import ObjectMapper

class ConfigController: PLViewController {

	@IBOutlet weak var wrapperView: UIView!
	@IBOutlet weak var idTextField: UITextField!
	@IBOutlet weak var hostTextField: UITextField!
	@IBOutlet weak var modeControl: UISegmentedControl!
	@IBOutlet weak var arduinoView: UIView!
	@IBOutlet weak var webHostTextField: UITextField!

	var arduinoViewMap: [String: UILabel] = [String: UILabel]()
	var timer = NSTimer()

	@IBAction func modeChange(sender: UISegmentedControl) {
		WsClient.singleton.sendJSON(JSON([
			"cmd": "arduinoModeChange",
			"mode": sender.selectedSegmentIndex
			]))
	}
	@IBAction func saveID() {
		if idTextField.text != nil && idTextField.text?.characters.count > 0 {
			Defaults[.deviceID] = idTextField.text!
			WsClient.singleton.connect(PLConstants.getWsAddress())
		}
	}

	@IBAction func saveConfig() {
		if hostTextField.text != nil && hostTextField.text?.characters.count > 0 {
			Defaults[.host] = hostTextField.text!
			WsClient.singleton.connect(PLConstants.getWsAddress())
		}
	}

	@IBAction func saveWebHost() {
		if webHostTextField.text != nil && webHostTextField.text?.characters.count > 0 {
			Defaults[.websiteHost] = webHostTextField.text!
		}
	}

	@IBAction func showLaserMenu(sender: UITapGestureRecognizer) {
		let alert = UIAlertController(title: "激光控制", message: nil, preferredStyle: .ActionSheet)
		alert.addAction(UIAlertAction(title: "调试激光", style: .Default, handler: { (action) in
			self.performSegueWithIdentifier("ShowLaserLoop", sender: nil)
			}));
		alert.addAction(UIAlertAction(title: "检测激光", style: .Default, handler: { (action) in
			self.performSegueWithIdentifier("ShowQuickCheck", sender: nil)
			}));
		alert.addAction(UIAlertAction(title: "取消", style: .Cancel, handler: nil));
		alert.popoverPresentationController?.sourceView = sender.view;
		presentViewController(alert, animated: true, completion: nil)
	}

	override func viewDidLoad() {
		super.viewDidLoad()
		idTextField.placeholder = Defaults[.deviceID]
		hostTextField.placeholder = Defaults[.host]
		webHostTextField.placeholder = Defaults[.websiteHost]
		let scrollView = AutoKeyboardScrollView()
		scrollView.backgroundColor = UIColor.clearColor()
		view.addSubview(scrollView)
		wrapperView.removeFromSuperview()
		scrollView.addSubview(wrapperView)
		scrollView <- Edges()
		wrapperView <- Edges()

		modeControl.tintColor = UIColor.whiteColor()
	}

	override func viewWillAppear(animated: Bool) {
		super.viewWillAppear(animated)
		DataManager.singleton.subscribeData([.ArduinoList], receiver: self)
		timer = NSTimer.scheduledTimerWithTimeInterval(2, target: self, selector: #selector(queryArduinoList), userInfo: nil, repeats: true)
	}

	override func viewDidDisappear(animated: Bool) {
		super.viewDidDisappear(animated)
		DataManager.singleton.unsubscribe(self)
		timer.invalidate()
	}

	func queryArduinoList() {
		WsClient.singleton.sendCmd(DataType.ArduinoList.queryCmd)
	}

	func renderArduinoList(list: [ArduinoController]) {
		let margin: CGFloat = 10
		let row = 8
		let width: CGFloat = 100
		let height: CGFloat = 20
		let firstRender = self.arduinoViewMap.count == 0
		for (i, controller) in list.enumerate() {
			let label: UILabel
			if firstRender {
				label = UILabel()
				label.text = controller.address.id
				label.font = UIFont.systemFontOfSize(12)
				let top = CGFloat(i / row) * (margin + height) + margin
				let left = CGFloat(i % row) * (margin + width) + margin
				label.frame = CGRect(x: left, y: top, width: width, height: height)
				label.textAlignment = .Center
				self.arduinoView.addSubview(label)
				arduinoViewMap[controller.address.id] = label
			} else {
				label = arduinoViewMap[controller.address.id]!
			}
			if (controller.online == true) {
				if controller.mode == .On {
					label.textColor = UIColor.blueColor()
				} else if controller.mode == .Off {
					label.textColor = UIColor.blackColor()
				} else if controller.mode == .Free {
					label.textColor = UIColor.greenColor()
				} else if controller.mode == .Scan {
					label.textColor = UIColor.orangeColor()
				} else {
					label.textColor = UIColor.purpleColor()
				}
			} else {
				label.textColor = UIColor.redColor()
			}
			label.borderWidth = 0
			if (controller.address.type == .MainArduino) {
				if !controller.scoreUpdated {
					label.borderWidth = 1
					label.borderColor = UIColor.redColor()
				}
			}
		}
	}
}

// PLViewController
extension ConfigController {
	override func onWsDisconnected() {
		super.onWsDisconnected()
		if self.presentedViewController != nil {
			self.dismissViewControllerAnimated(false, completion: nil)
		}
		for (_, label) in self.arduinoViewMap {
			label.textColor = UIColor.redColor()
		}
		modeControl.selectedSegmentIndex = 0
	}
}

extension ConfigController: DataReceiver {
	func onReceivedData(json: [String: AnyObject], type: DataType) {
		if type == .ArduinoList {
			let arduinoList = Mapper<ArduinoController>().mapArray(json["data"])
			if arduinoList != nil {
				renderArduinoList(arduinoList!)
			}
		}
	}
}
