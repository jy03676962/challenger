//
//  LoginTextField.swift
//  postgame
//
//  Created by tassar on 4/17/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

/*
 登陆用的TextField，在IB里面设置了边框相关属性，placeholder颜色属性
 自定了placeholderColor IBInspectable
 在commonInit设置了leftView和clearButton的样式
 总之目前是个通用性很低的类，单独成类是为了代码可读性(比较时候将某些feature抽到通用类)
 */
@IBDesignable
class LoginTextField: UITextField {
	required init?(coder aDecoder: NSCoder) {
		super.init(coder: aDecoder)
		commonInit()
	}

	override init(frame: CGRect) {
		super.init(frame: frame)
		commonInit()
	}

	@IBInspectable var placeholderColor: UIColor? {
		didSet {
			guard placeholderColor != nil else {
				return
			}
			guard placeholder != nil else {
				return
			}
			attributedPlaceholder = NSAttributedString(string: placeholder!, attributes: [NSForegroundColorAttributeName: placeholderColor!])
		}
	}

	override var placeholder: String? {
		didSet {
			guard placeholder != nil else {
				return
			}
			guard placeholderColor != nil else {
				return
			}
			attributedPlaceholder = NSAttributedString(string: placeholder!, attributes: [NSForegroundColorAttributeName: placeholder!])
		}
	}

	func commonInit() {
		leftView = UIView(frame: CGRect(x: 0, y: 0, width: 10, height: 0))
		leftViewMode = .Always
		let btn = valueForKey("_clearButton")
		btn?.setImage(UIImage(named: "TextFieldClear"), forState: .Normal)
	}
}
