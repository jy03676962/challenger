//
//  Constants.swift
//  postgame
//
//  Created by tassar on 4/7/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import SwiftyUserDefaults

struct PLConstants {
	static let usualFont = "Alien League Bold"
	static let defaultWebsiteHost = "gameapi.pulupulu.com"
	static let maxTeamSize = 4
	static func getHost() -> String {
		return Defaults[.host]
	}
	static func getWsAddress() -> String {
		return "ws://" + getHost() + "/ws"
	}
	static func getHttpAddress(path: String) -> String {
		let p = path.hasPrefix("/") ? path : "/" + path
		return "http://" + getHost() + p
	}
	static func getWebsiteAddress(path: String) -> String {
		let p = path.hasPrefix("/") ? path : "/" + path
		return "http://" + Defaults[.websiteHost] + p
	}
	static func getCid() -> String {
		return "\(Defaults[.socketType]):\(Defaults[.deviceID])"
	}
}

enum GameMode: Int {
	case Fun = 1, Survival
}