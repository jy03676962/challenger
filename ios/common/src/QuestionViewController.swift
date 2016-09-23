//
//  QuestionTableViewController.swift
//  postgame
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit

protocol QuestionViewControllerDelegate: class {
	func okAction(sender: QuestionViewController, answer: String)
}

class QuestionViewController: UIViewController {
	@IBOutlet weak var titleLabel: UILabel!
	@IBOutlet weak var okButton: UIButton!
	@IBOutlet weak var tableView: UITableView!
	var question: SurveyQuestion!
	var questionIndex: Int = 0
	var isLastQuestion = false
	weak var delegate: QuestionViewControllerDelegate?

	@IBAction func okAction() {
		let idx = tableView.indexPathsForSelectedRows![0].row
		var ans = ""
		ans.append(Character(UnicodeScalar(idx + 17)))
		delegate?.okAction(self, answer: ans)
	}

	override func viewDidLoad() {
		super.viewDidLoad()
		titleLabel.text = question.q
		tableView.reloadData()
		if isLastQuestion {
			okButton.setBackgroundImage(UIImage(named: "SurveyDone"), forState: .Normal)
		} else {
			okButton.setBackgroundImage(UIImage(named: "SurveyOK"), forState: .Normal)
		}
	}
}

extension QuestionViewController: UITableViewDataSource, UITableViewDelegate {

	func numberOfSectionsInTableView(tableView: UITableView) -> Int {
		return 1
	}

	func tableView(tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
		return question.options.count
	}
	func tableView(tableView: UITableView, cellForRowAtIndexPath indexPath: NSIndexPath) -> UITableViewCell {
		let cell = tableView.dequeueReusableCellWithIdentifier("QuestionTableViewCell") as! QuestionTableViewCell
		cell.setData(question.options[indexPath.row])
		return cell
	}
	func tableView(tableView: UITableView, didSelectRowAtIndexPath indexPath: NSIndexPath) {
		okButton.enabled = true
	}
	func tableView(tableView: UITableView, didDeselectRowAtIndexPath indexPath: NSIndexPath) {
		okButton.enabled = false
	}
}