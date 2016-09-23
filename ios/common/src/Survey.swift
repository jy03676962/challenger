//
//  Survey.swift
//  postgame
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import ObjectMapper

struct SurveyQuestion: Mappable {
	var q: String!
	var options: [String]!
	init?(_ map: Map) {
	}
	mutating func mapping(map: Map) {
		q <- map["q"]
		options <- map["options"]
	}
}

class Survey: Mappable {
	var questions: [SurveyQuestion]!
	required init?(_ map: Map) {
	}
	func mapping(map: Map) {
		questions <- map["questions"]
	}
}