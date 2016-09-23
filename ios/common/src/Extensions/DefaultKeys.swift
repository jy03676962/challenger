//
//  DefaultKeys.swift
//  admin
//
//  Created by tassar on 4/23/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import SwiftyUserDefaults

extension DefaultsKeys {
	static let host = DefaultsKey<String>("host")
	static let deviceID = DefaultsKey<String>("deviceID")
	static let socketType = DefaultsKey<String>("socketType")
	static let matchID = DefaultsKey<Int>("matchID")
	static let qCount = DefaultsKey<Int>("qCount")
	static let websiteHost = DefaultsKey<String>("websiteHost")
}
