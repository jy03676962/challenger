//
//  Team.swift
//  admin
//
//  Created by tassar on 4/26/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

enum TeamStatus: Int {
	case Waiting = 0, Prepare, Playing, After, Finished
}

class Team: Mappable {
	var size: Int!
	var id: String!
	var delayCount: Int!
	var status: TeamStatus!
	var waitTime: Int!
	var mode: String!

	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		size <- map["size"]
		id <- map["id"]
		delayCount <- map["delayCount"]
		status <- map["status"]
		waitTime <- map["waitTime"]
		mode <- map["mode"]
	}
}
