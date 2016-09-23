//
//  Match.swift
//  admin
//
//  Created by tassar on 5/5/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

struct Position: Mappable {
	var X: Double!
	var Y: Double!
	init?(_ map: Map) {
	}

	mutating func mapping(map: Map) {
		X <- map["X"]
		Y <- map["Y"]
	}
}

class Laser: Mappable {
	var isPause: Bool!
	var displayP: Position!
	var displayP2: Position!
	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		isPause <- map["isPause"]
		displayP <- map["displayP"]
		displayP2 <- map["displayP2"]
	}
}

class Player: Mappable {
	var pos: Position!
	var dir: String!
	var button: String!
	var buttonTime: Double!
	var buttonLevel: Int!
	var gold: Int!
	var energy: Double!
	var levelData: [Int]!
	var hitCount: Int!
	var lostGold: Int!
	var invincibleTime: Double!
	var combo: Int!
	var comboCount: Int!
	var controllerID: String!
	var displayPos: Position!

	var displayName: String {
		return String(format: "[%02d]", Int(controllerID.componentsSeparatedByString(":")[1])!)
	}

	var displayID: String {
		return String(format: "%02d", Int(controllerID.componentsSeparatedByString(":")[1])!)
	}

	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		pos <- map["pos"]
		displayPos <- map["displayPos"]
		dir <- map["dir"]
		button <- map["button"]
		buttonTime <- map["buttonTime"]
		buttonLevel <- map["buttonLevel"]
		gold <- map["gold"]
		energy <- map["energy"]
		levelData <- map["levelData"]
		hitCount <- map["hitCount"]
		lostGold <- map["lostgold"]
		invincibleTime <- map["invincibleTime"]
		combo <- map["combo"]
		comboCount <- map["comboCount"]
		controllerID <- map["cid"]
	}
}

class Match: Mappable {
	var member: [Player]!
	var stage: String!
	var totalTime: Double!
	var elasped: Double!
	var warmupTime: Double!
	var rampageTime: Double!
	var mode: String!
	var gold: Int!
	var energy: Double!
	var rampageCount: Int!
	var id: Int!
	var teamID: String!
	var maxEnergy: Int!
	var isSimulator: Int!
	var lasers: [Laser]?

	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		member <- map["member"]
		stage <- map["stage"]
		totalTime <- map["totalTime"]
		elasped <- map["elasped"]
		warmupTime <- map["warmupTime"]
		rampageTime <- map["rampageTime"]
		mode <- map["mode"]
		gold <- map["gold"]
		energy <- map["energy"]
		rampageCount <- map["rampageCount"]
		id <- map["id"]
		teamID <- map["teamID"]
		maxEnergy <- map["maxEnergy"]
		isSimulator <- map["isSimulator"]
		lasers <- map["lasers"]
	}
}
