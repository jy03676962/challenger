//
//  String+PL.swift
//  admin
//
//  Created by tassar on 5/2/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
extension String {

	subscript(i: Int) -> Character {
		return self[self.startIndex.advancedBy(i)]
	}

	subscript(i: Int) -> String {
		return String(self[i] as Character)
	}

	subscript(r: Range<Int>) -> String {
		let start = startIndex.advancedBy(r.startIndex)
		let end = start.advancedBy(r.endIndex - r.startIndex)
		return self[Range(start ..< end)]
	}

	func last() -> String {
		return String(self[self.endIndex.predecessor()])
	}
}