//
//  Address.swift
//  admin
//
//  Created by tassar on 5/1/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

enum AddressType: Int {
	case Unknown = 0, Admin, Simulator, ArduinoTest, Postgame, Wearable, MainArduino, SubArduino, QueueDevice, IngameDevice, MusicArduino, DoorArduino
}

struct Address: Mappable {
	var type: AddressType!
	var id: String!
	init?(_ map: Map) {
	}
	mutating func mapping(map: Map) {
		type <- map["type"]
		id <- map["id"]
	}
}

class PlayerController: Mappable {
	var address: Address!
	var id: String!
	var matchID: Int!
	var online: Bool!

	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		address <- map["address"]
		id <- map["id"]
		matchID <- map["matchID"]
		online <- map["online"]
	}
}