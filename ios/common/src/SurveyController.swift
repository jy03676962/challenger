//
//  SurveyController.swift
//  postgame
//
//  Created by tassar on 5/8/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import EasyPeasy
import Alamofire
import PKHUD

class SurveyController: PLViewController {
	var survey: Survey!
	var playerData: PlayerData!

	@IBOutlet weak var scrollView: UIScrollView!
	@IBOutlet weak var pageControl: UIPageControl!
	var vcs: [QuestionViewController]!

	override func viewDidLoad() {
		super.viewDidLoad()
		scrollView.contentSize = CGSize(width: 960 * self.survey.questions.count, height: 430)
		vcs = [QuestionViewController]()
		let sb = UIStoryboard(name: "MatchResult", bundle: nil)
		for (i, q) in survey.questions.enumerate() {
			let vc = sb.instantiateViewControllerWithIdentifier("QuestionViewController") as! QuestionViewController
			vc.question = q
			vc.questionIndex = i
			vc.isLastQuestion = i == survey.questions.count - 1
			addChildViewController(vc)
			scrollView.addSubview(vc.view)
			let size = scrollView.bounds.size
			let left = CGFloat(i) * size.width
			vc.view <- [
				Size(size),
				Left(left),
				Top(0)
			]
			vc.didMoveToParentViewController(self)
			vc.delegate = self
			vcs.append(vc)
		}
		pageControl.currentPage = 0
	}

	override func viewDidLayoutSubviews() {
		super.viewDidLayoutSubviews()
	}
}

extension SurveyController: QuestionViewControllerDelegate {
	func okAction(sender: QuestionViewController, answer: String) {
		let idx = sender.questionIndex
		let p = [
			"pid": String(playerData.id),
			"qid": String(idx + 1),
			"aid": answer,
			"total": String(self.survey.questions.count)
		]
		Alamofire.request(.POST, PLConstants.getHttpAddress("api/answer"), parameters: p, encoding: .URL, headers: nil)
			.validate()
			.responseJSON(completionHandler: { resp in
				if !sender.isLastQuestion {
					var frame = self.scrollView.frame
					frame.origin.x = frame.size.width * CGFloat(idx + 1)
					frame.origin.y = 0
					self.scrollView.scrollRectToVisible(frame, animated: true)
					self.pageControl.currentPage = idx + 1
				} else {
					self.navigationController?.popViewControllerAnimated(true)
				}
		})
	}
}
