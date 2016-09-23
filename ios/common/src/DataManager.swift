//
//  DataManager.swift
//  admin
//
//  Created by tassar on 4/25/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import Starscream
import SwiftyJSON

public protocol DataReceiver: class {
	func onReceivedData(json: [String: AnyObject], type: DataType)
}

public enum DataType: String {
	case HallData = "HallData"
	case ControllerData = "ControllerData"
	case NewMatch = "newMatch"
	case ArduinoList = "ArduinoList"
	case UpdateMatch = "updateMatch"
	case MatchStop = "matchStop"
	case StartAnswer = "startAnswer"
	case StopAnswer = "stopAnswer"
	case UpdatePlayerData = "updatePlayerData"
	case QuestionCount = "QuestionCount"
	case LaserInfo = "laserInfo"
	case QuickCheckInfo = "QuickCheck"
	case Error = "error"

	var queryCmd: String {
		return "query\(self.rawValue)"
	}

	var shouldQuery: Bool {
		let first: String = self.rawValue[0]
		return first.uppercaseString == first
	}
}

public class DataManager {

	private var receiversMap: [DataType: [DataReceiver]] = [:]

	public static let singleton = DataManager()

	public func subscribeData(types: [DataType], receiver: DataReceiver) {
		for type in types {
			var list = receiversMap[type] ?? [DataReceiver]()
			if !list.contains({ (rcv) -> Bool in rcv === receiver }) {
				list.append(receiver)
				receiversMap[type] = list
			}
			if type.shouldQuery {
				WsClient.singleton.sendCmd(type.queryCmd)
			}
		}
	}

	public func unsubscribe(receiver: DataReceiver) {
		for (t, l) in receiversMap {
			var nl = [DataReceiver]()
			for r in l {
				if r !== receiver {
					nl.append(r)
				}
			}
			receiversMap[t] = nl
		}
	}
	public func unsubscribe(receiver: DataReceiver, type: DataType) {
		for (t, l) in receiversMap {
			if t != type {
				continue
			}
			var nl = [DataReceiver]()
			for r in l {
				if r !== receiver {
					nl.append(r)
				}
			}
			receiversMap[t] = nl
		}
	}

	public func queryData(type: DataType) {
		WsClient.singleton.sendCmd(type.queryCmd)
	}

	private func dispatch() {
	}

	private init() {
		WsClient.singleton.delegate = self
	}
}

// MARK: websocket notificaiton
extension DataManager: WsClientDelegate {

	public func wsClientDidInit(client: WsClient, data: [String: AnyObject]) {
		for (type, _) in receiversMap {
			WsClient.singleton.sendCmd(type.queryCmd)
		}
	}

	public func wsClientDidDisconnect(client: WsClient, error: NSError?) {
	}

	public func wsClientDidReceiveMessage(client: WsClient, cmd: String, data: [String: AnyObject]) {
		for (type, receivers) in receiversMap {
			if type.rawValue == cmd {
				for receiver in receivers {
					receiver.onReceivedData(data, type: type)
				}
			}
		}
	}
}