//
//  HistoryController.swift
//  admin
//
//  Created by tassar on 4/20/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import Alamofire
import AlamofireObjectMapper
import PKHUD
import SwiftyJSON

let segueIDPresentMatchResult = "PresentMatchResult"

class HistoryController: PLViewController {
	@IBOutlet weak var tableView: UITableView!
	@IBOutlet weak var startAnswerButton: UIButton!

	var refreshControl: UIRefreshControl!
	var data: [MatchData]?
	var isAnswering: Int? {
		guard let d = data else {
			return nil
		}
		for (i, m) in d.enumerate() {
			if m.answerType == .Answering {
				return i
			}
		}
		return nil
	}

	@IBAction func startAnswer() {
		guard let data = self.data, let indexPaths = tableView.indexPathsForSelectedRows where indexPaths.count == 1 else {
			return
		}
		if let ing = isAnswering {
			if ing != indexPaths[0].row {
				HUD.flash(.LabeledError(title: "有其他正在答题的队伍", subtitle: "请先结束该组答题后重试"), delay: 1)
				return
			}
		}
		let matchData = data[indexPaths[0].row]
		if matchData.eid != nil && matchData.eid != "" {
			self.startAnswerAfterAdd(matchData)
		} else {
			HUD.show(.Progress)
			var playerDataList: [AnyObject] = []
			for player in matchData.member {
				let pd: [String: AnyObject] = [
					"player_id": player.cid,
					"player_score": String(player.gold - player.lostGold),
					"player_catch": String(player.hitCount),
					"player_rank": player.grade,
				]
				playerDataList.append(pd)
			}
			let p: [String: AnyObject] = [
				"mode": matchData.mode == "g" ? "0" : "1",
				"time": String(Int(matchData.elasped * 1000)),
				"gold": String(matchData.gold),
				"player_num": String(matchData.member.count),
				"team_rampage": String(matchData.rampageCount),
				"team_rank": matchData.grade,
				"player_data": JSON(playerDataList).rawString()!,
			]
			Alamofire.request(.POST, PLConstants.getWebsiteAddress("challenger/match"), parameters: p, encoding: .URL, headers: nil)
				.validate()
				.responseObject(completionHandler: { (response: Response<AddMatchResult, NSError>) in
					HUD.hide()
					if let err = response.result.error {
						HUD.flash(.LabeledError(title: err.localizedDescription, subtitle: nil), delay: 2)
					} else {
						if response.result.value?.code != 0 {
							HUD.flash(.LabeledError(title: response.result.value?.error, subtitle: nil), delay: 2)
						} else {
							matchData.eid = String(response.result.value!.matchID)
							self.startAnswerAfterAdd(matchData)
						}
					}
			})
		}
	}

	override func viewDidLoad() {
		super.viewDidLoad()
		refreshControl = UIRefreshControl()
		refreshControl.addTarget(self, action: #selector(refreshHistory), forControlEvents: .ValueChanged)
		tableView.addSubview(refreshControl)
	}

	override func viewWillAppear(animated: Bool) {
		super.viewWillAppear(animated)
		refreshHistory()
	}

	func refreshHistory() {
		Alamofire.request(.GET, PLConstants.getHttpAddress("api/history"))
			.validate()
			.responseArray(completionHandler: { (response: Response<[MatchData], NSError>) in
				self.data = response.result.value
				if self.data != nil {
					self.tableView.reloadData()
				}
				self.refreshControl.endRefreshing()
		})
	}

	func startAnswerAfterAdd(matchData: MatchData) {
		HUD.show(.Progress)
		Alamofire.request(.POST, PLConstants.getHttpAddress("api/start_answer"), parameters: ["mid": matchData.id, "eid": matchData.eid!], encoding: .URL, headers: nil)
			.validate()
			.responseObject(completionHandler: { (response: Response<MatchData, NSError>) in
				HUD.hide()
				if let err = response.result.error {
					HUD.flash(.LabeledError(title: err.localizedDescription, subtitle: nil), delay: 2)
				} else {
					self.performSegueWithIdentifier(segueIDPresentMatchResult, sender: matchData)
				}
		})
	}

	override func prepareForSegue(segue: UIStoryboardSegue, sender: AnyObject?) {
		if segue.identifier == segueIDPresentMatchResult {
			let vc = segue.destinationViewController as! MatchResultController
			vc.isAdmin = true
			vc.matchData = sender as? MatchData
		}
	}
}

extension HistoryController: UITableViewDelegate, UITableViewDataSource {
	func tableView(tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
		return self.data == nil ? 0 : self.data!.count
	}
	func numberOfSectionsInTableView(tableView: UITableView) -> Int {
		return 1
	}
	func tableView(tableView: UITableView, cellForRowAtIndexPath indexPath: NSIndexPath) -> UITableViewCell {
		let cell = tableView.dequeueReusableCellWithIdentifier("HistoryTableViewCell") as! HistoryTableViewCell
		cell.setData(data![indexPath.row])
		return cell
	}
	func tableView(tableView: UITableView, didSelectRowAtIndexPath indexPath: NSIndexPath) {
		startAnswerButton.enabled = true
	}
	func tableView(tableView: UITableView, didDeselectRowAtIndexPath indexPath: NSIndexPath) {
		startAnswerButton.enabled = false
	}
}
