//
//  WebSocketClient.swift
//  postgame
//
//  Created by tassar on 4/14/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import Starscream
import SwiftyUserDefaults
import SwiftyJSON

public protocol WsClientDelegate: class {
	func wsClientDidInit(client: WsClient, data: [String: AnyObject])
	func wsClientDidReceiveMessage(client: WsClient, cmd: String, data: [String: AnyObject])
	func wsClientDidDisconnect(client: WsClient, error: NSError?)
}

public class WsClient {
	// notification
	public static let WsInitedNotification = "WsInited"
	public static let WsDisconnectedNotification = "WsDisconnected"
	public static let WsConnectingNotification = "WsConnecting"

	public static let singleton = WsClient()
	public weak var delegate: WsClientDelegate?
	public var didInit: Bool = false

	private static let ERROR_WAIT_SECOND: UInt64 = 10
	private var socket: WebSocket?
	private var address: String?

	public func sendCmd(cmd: String) {
		if !didInit {
			return
		}
		let json = JSON([
			"cmd": cmd
		])
		sendJSON(json)
	}

	public func sendJSON(json: JSON) {
		let str = json.rawString(NSUTF8StringEncoding, options: [])!
		socket!.writeString(str)
	}

	public func connect(addr: String) {
		if address == addr && socket != nil && socket!.isConnected {
			return
		}
		address = addr
		if socket == nil {
			initSocket()
			doConnect()
		} else if socket!.isConnected {
			socket!.disconnect()
		}
	}

	@objc func appDidEnterBackground() {
		if socket != nil && socket!.isConnected {
			socket!.disconnect()
		}
	}

	@objc func appWillEnterForeground() {
		if socket != nil && socket!.isConnected {
			return
		}
		guard address != nil else {
			return
		}
		initSocket()
		doConnect()
	}

	private init() {
		NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(WsClient.appDidEnterBackground), name: UIApplicationDidEnterBackgroundNotification, object: nil)
		NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(WsClient.appWillEnterForeground), name: UIApplicationWillEnterForegroundNotification, object: nil)
	}

	deinit {
		NSNotificationCenter.defaultCenter().removeObserver(self)
	}

	private func initSocket() {
		socket = WebSocket(url: NSURL(string: address!)!)
		socket?.delegate = self
	}

	private func doConnect() {
		socket!.connect()
		NSNotificationCenter.defaultCenter().postNotificationName(WsClient.WsConnectingNotification, object: nil)
	}
}

// MARK: websocket回调方法
extension WsClient: WebSocketDelegate {

	public func websocketDidConnect(socket: WebSocket) {
		log.debug("socket connected")
		let json = JSON([
			"cmd": "init",
			"ID": Defaults[.deviceID],
			"TYPE": Defaults[.socketType],
		])
		self.sendJSON(json)
	}

	public func websocketDidReceiveData(socket: WebSocket, data: NSData) {
	}

	public func websocketDidDisconnect(socket: WebSocket, error: NSError?) {
		self.didInit = false
		delegate?.wsClientDidDisconnect(self, error: error)
		NSNotificationCenter.defaultCenter().postNotificationName(WsClient.WsDisconnectedNotification, object: nil)
		if UIApplication.sharedApplication().applicationState == .Background {
			return
		}
		if socket.currentURL.absoluteString != address {
			initSocket()
		}
		if error == nil {
			doConnect()
		} else {
			dispatch_after(dispatch_time(DISPATCH_TIME_NOW, Int64(WsClient.ERROR_WAIT_SECOND * NSEC_PER_SEC)), dispatch_get_main_queue(), {
				self.doConnect()
			})
		}
	}

	public func websocketDidReceiveMessage(socket: WebSocket, text: String) {
		log.debug("socket got:\(text)")
		let dataFromString = text.dataUsingEncoding(NSUTF8StringEncoding, allowLossyConversion: false)
		guard dataFromString != nil else {
			return
		}
		let json = JSON(data: dataFromString!)
		guard json.type == .Dictionary else {
			return
		}
		let cmd = json["cmd"].string
		guard cmd != nil else {
			return
		}
		if cmd == "init" {
			self.didInit = true
			NSNotificationCenter.defaultCenter().postNotificationName(WsClient.WsInitedNotification, object: nil)
			self.delegate?.wsClientDidInit(self, data: json.dictionaryObject!)
		} else if self.didInit {
			self.delegate?.wsClientDidReceiveMessage(self, cmd: cmd!, data: json.dictionaryObject!)
		}
	}
}
