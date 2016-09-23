//
//  MainArduinoInfo.swift
//  admin
//
//  Created by tassar on 5/21/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

class MainArduinoInfo: Mappable {
	var id: String!
	var dir: Int!
	var x: Int!
	var y: Int!
	var type: String!
	var laserNum: Int!
	required init?(_ map: Map) {
	}

	func mapping(map: Map) {
		id <- map["id"]
		dir <- map["dir"]
		x <- map["x"]
		y <- map["y"]
		type <- map["type"]
		laserNum <- map["laserNum"]
	}
}