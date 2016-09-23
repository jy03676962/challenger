//
//  ReplaceSeuge.swift
//  postgame
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

class ReplaceSeuge: UIStoryboardSegue {

	override func perform() {
		let navigationController: UINavigationController = sourceViewController.navigationController!;

		var controllerStack = navigationController.viewControllers;
		let index = controllerStack.indexOf(sourceViewController);
		controllerStack[index!] = destinationViewController

		navigationController.setViewControllers(controllerStack, animated: true);
	}
}
