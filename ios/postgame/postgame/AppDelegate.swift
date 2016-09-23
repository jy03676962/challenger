//
//  AppDelegate.swift
//  postgame
//
//  Created by tassar on 3/29/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import XCGLogger
import SwiftyUserDefaults
import ObjectMapper
import PKHUD

let log = XCGLogger.defaultInstance()

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {

	var window: UIWindow?
	var navi: UINavigationController? {
		return window?.rootViewController as? UINavigationController
	}

	func application(application: UIApplication, didFinishLaunchingWithOptions launchOptions: [NSObject: AnyObject]?) -> Bool {
		Instabug.startWithToken("c9b33e734887212b949d3f9944652f22", invocationEvent: IBGInvocationEvent.Shake)
		#if DEBUG
			log.setup(.Debug, showThreadName: true, showLogLevel: true, showFileNames: true, showLineNumbers: true, writeToFile: nil)
		#else
			log.setup(.Severe, showThreadName: true, showLogLevel: true, showFileNames: true, showLineNumbers: true, writeToFile: nil)
		#endif
		if Defaults[.host] == "" {
			Defaults[.host] = "192.168.1.5:3000"
		}
		if Defaults[.deviceID] == "" {
			Defaults[.deviceID] = "1"
		}
		if Defaults[.websiteHost] == "" {
			Defaults[.websiteHost] = PLConstants.defaultWebsiteHost
		}
		Defaults[.socketType] = "4"
		Defaults[.matchID] = 0
		DataManager.singleton.subscribeData([.StopAnswer], receiver: self)
		WsClient.singleton.connect(PLConstants.getWsAddress())
		return true
	}
}

extension AppDelegate: DataReceiver {
	func onReceivedData(json: [String: AnyObject], type: DataType) {
		if type == .StopAnswer {
			guard navi?.visibleViewController as? LoginViewController == nil else {
				return
			}
			HUD.hide()
			let sb = UIStoryboard(name: "Main", bundle: nil)
			let login = sb.instantiateViewControllerWithIdentifier("LoginViewController")
			navi?.setViewControllers([login], animated: true)
		}
	}
}
