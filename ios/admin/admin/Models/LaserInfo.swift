//
//  LaserInfo.swift
//  admin
//
//  Created by tassar on 5/21/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

class LaserInfo: Mappable {
	var id: String!
	var ur: String!
	var idx: Int!
	var err: Int!
	required init?(_ map: Map) {
	}
	func mapping(map: Map) {
		id <- map["id"]
		ur <- map["ur"]
		idx <- map["idx"]
		err <- map["error"]
	}
}
