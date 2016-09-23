//
//  LoginViewController.swift
//  postgame
//
//  Created by tassar on 3/31/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import Alamofire
import AlamofireObjectMapper
import AutoKeyboardScrollView
import SVProgressHUD
import EasyPeasy
import SwiftyUserDefaults
import PKHUD

let SegueIDShowMatchResult = "ShowMatchResult"

class LoginViewController: PLViewController {

	/*
	 为什么要这个wrapperView看下面
	 @link https://github.com/honghaoz/AutoKeyboardScrollView#work-with-interface-builder
	 */
	@IBOutlet weak var wrapperView: UIView!
	@IBOutlet weak var usernameTextField: LoginTextField!
	@IBOutlet weak var passwordTextField: LoginTextField!
	@IBOutlet weak var deviceIDTextField: LoginTextField!
	@IBOutlet weak var loginButton: UIButton!

	/**
	 双击登陆界面右上角出现配置窗口
	 */
	@IBAction func showConfig(sender: UITapGestureRecognizer) {
		let alert = UIAlertController(title: "设置", message: nil, preferredStyle: .Alert)
		alert.addTextFieldWithConfigurationHandler { (textfield) in
			textfield.placeholder = Defaults[.host]
		}
		alert.addTextFieldWithConfigurationHandler { textfield in
			textfield.placeholder = Defaults[.deviceID]
		}
		alert.addTextFieldWithConfigurationHandler { textfield in
			textfield.placeholder = Defaults[.websiteHost]
		}
		let cancelAction = UIAlertAction(title: "取消", style: .Cancel, handler: nil)
		alert.addAction(cancelAction)
		weak var weakAlert = alert
		let doneAction = UIAlertAction(title: "确定", style: .Default) { (action) in
			if let host = weakAlert?.textFields![0].text where host != "" {
				Defaults[.host] = host
				WsClient.singleton.connect(PLConstants.getWsAddress())
			}
			if let num = weakAlert?.textFields![1].text where num != "" {
				Defaults[.deviceID] = num
			}
			if let host = weakAlert?.textFields![2].text where host != "" {
				Defaults[.websiteHost] = host
			}
		}
		alert.addAction(doneAction)
		presentViewController(alert, animated: true, completion: nil)
	}

	@IBAction func textFieldValueChanged(sender: UITextField) {
		if usernameTextField.text?.characters.count > 0 && passwordTextField.text?.characters.count > 0 && deviceIDTextField.text?.characters.count > 0 {
			self.loginButton.enabled = true
		} else {
			self.loginButton.enabled = false
		}
	}

	@IBAction func login() {
		HUD.show(.Progress)
		let p = [
			"username": self.usernameTextField.text!,
			"password": self.passwordTextField.text!,
		]
		Alamofire.request(.POST, PLConstants.getWebsiteAddress("user/login"), parameters: p, encoding: .URL, headers: nil)
			.validate()
			.responseObject(completionHandler: { (resp: Response<LoginResult, NSError>) in
				HUD.hide()
				if let _ = resp.result.error {
					HUD.flash(.Error, delay: 2)
				} else {
					let m = resp.result.value!
					if m.code != nil && m.code == 0 {
						self.performSegueWithIdentifier(SegueIDShowMatchResult, sender: m)
					} else {
						HUD.flash(.LabeledError(title: m.error, subtitle: nil), delay: 2)
					}
				}
		})
	}
	@IBAction func skip() {
		performSegueWithIdentifier(SegueIDShowMatchResult, sender: nil)
	}

	override func prepareForSegue(segue: UIStoryboardSegue, sender: AnyObject?) {
		if segue.identifier == SegueIDShowMatchResult {
			let vc = segue.destinationViewController as! MatchResultController
			vc.isAdmin = false
			vc.loginInfo = sender as? LoginResult
		}
	}
}

// MARK: UIViewController
extension LoginViewController {

	override func viewDidLoad() {
		super.viewDidLoad()
		let scrollView = AutoKeyboardScrollView()
		view.addSubview(scrollView)
		wrapperView.removeFromSuperview()
		scrollView.contentView.addSubview(wrapperView)
		scrollView.backgroundColor = wrapperView.backgroundColor
		scrollView.userInteractionEnabled = true
		scrollView.bounces = true
		scrollView.scrollEnabled = true
		scrollView <- Edges()
		wrapperView <- Edges()
		scrollView.setTextMargin(175, forTextField: usernameTextField)
		scrollView.setTextMargin(140, forTextField: passwordTextField)
	}
}
