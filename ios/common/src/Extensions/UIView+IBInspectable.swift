//
//  UIView+IBInspectable.swift
//  postgame
//
//  Created by tassar on 4/17/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import Foundation
import UIKit

/*
 将UIView中一些常用属性暴露到IB中
 */
extension UIView {
	@IBInspectable var cornerRadius: CGFloat {
		get {
			return layer.cornerRadius
		}
		set {
			layer.cornerRadius = newValue
			layer.masksToBounds = newValue > 0
		}
	}
	@IBInspectable var borderColor: UIColor? {
		get {
			guard layer.borderColor != nil else {
				return nil
			}
			return UIColor(CGColor: layer.borderColor!)
		}
		set {
			layer.borderColor = newValue?.CGColor
		}
	}

	@IBInspectable var borderWidth: CGFloat {
		get {
			return layer.borderWidth
		}
		set {
			layer.borderWidth = newValue
		}
	}
}