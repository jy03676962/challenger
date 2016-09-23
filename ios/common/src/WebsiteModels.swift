//
//  WebsiteModels.swift
//  postgame
//
//  Created by tassar on 5/10/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

class BaseResult: Mappable {
	var code: Int?
	var error: String?
	let codeTransform = TransformOf<Int, AnyObject>(fromJSON: { (value: AnyObject?) -> Int? in
		if value == nil {
			return 0
		} else if let i = value as? Int {
			return i
		} else if let s = value as? String {
			return Int(s)
		} else {
			return -1
		}
		}, toJSON: { (value: Int?) -> AnyObject? in
		if let value = value {
			return String(value)
		}
		return nil
	})
	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		code <- (map["code"], codeTransform)
		error <- map["error"]
	}
}

class LoginResult: BaseResult {
	var username: String!
	var userID: Int!
	var currentCoin: Int!
	required init?(_ map: Map) {
		super.init(map)
	}

	override func mapping(map: Map) {
		super.mapping(map)
		username <- map["username"]
		userID <- map["user_id"]
		currentCoin <- map["current_coin"]
	}
}

class AddUserResult: BaseResult {
	var level: String?
	var url: String?
	required init?(_ map: Map) {
		super.init(map)
	}

	override func mapping(map: Map) {
		super.mapping(map)
		level <- map["level"]
		url <- map["url"]
	}
}

class AddMatchResult: BaseResult {
	var matchID: Int!
	required init?(_ map: Map) {
		super.init(map)
	}
	override func mapping(map: Map) {
		super.mapping(map)
		matchID <- map["match_id"]
	}
}
