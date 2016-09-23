//
//  HallController.swift
//  admin
//
//  Created by tassar on 4/20/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import SwiftyJSON
import SWTableViewCell
import ObjectMapper
import PKHUD

class HallController: PLViewController {
	private static let controllerButtonTagStart = 100

	@IBOutlet weak var teamtableView: UITableView!
	@IBOutlet weak var teamIDLabel: UILabel!
	@IBOutlet var controllerButtons: [DeviceButton]!

	@IBOutlet weak var modeImageView: UIImageView!
	@IBOutlet weak var modeLabel: UILabel!
	@IBOutlet weak var playerNumberLabel: UILabel!
	@IBOutlet weak var readyButton: UIButton!
	@IBOutlet weak var startButton: UIButton!
	@IBOutlet weak var addPlayerButton: UIButton!
	@IBOutlet weak var removePlayerButton: UIButton!
	@IBOutlet var changeModeTGR: UITapGestureRecognizer!
	@IBOutlet weak var callButton: UIButton!
	@IBOutlet weak var delayButton: UIButton!

	var refreshControl: UIRefreshControl!
	var teams: [Team]?
	var topTeam: Team?
	var controllers: [PlayerController]?
	var hasPlayingTeam = false

	override func viewDidLoad() {
		super.viewDidLoad()
		refreshControl = UIRefreshControl()
		refreshControl.addTarget(self, action: #selector(HallController.refreshTeamData), forControlEvents: UIControlEvents.ValueChanged)
		teamtableView.addSubview(refreshControl)
	}
	override func viewWillAppear(animated: Bool) {
		super.viewWillAppear(animated)
		DataManager.singleton.subscribeData([.HallData, .ControllerData, .NewMatch, .Error], receiver: self)
	}

	override func viewDidDisappear(animated: Bool) {
		super.viewDidDisappear(animated)
		DataManager.singleton.unsubscribe(self)
	}
	func refreshTeamData() {
		DataManager.singleton.queryData(.HallData)
	}
	@IBAction func changeMode() {
		guard let team = topTeam else {
			return
		}
		let mode = team.mode == "g" ? "s" : "g"
		let json = JSON([
			"cmd": "teamChangeMode",
			"teamID": topTeam!.id,
			"mode": mode
		])
		WsClient.singleton.sendJSON(json)
	}
	@IBAction func callTeam(sender: UIButton) {
		guard let team = topTeam else {
			return
		}
		let json = JSON([
			"cmd": "teamCall",
			"teamID": team.id,
		])
		WsClient.singleton.sendJSON(json)
	}
	@IBAction func delayTeam(sender: UIButton) {
		guard let team = topTeam else {
			return
		}
		guard team.status == .Waiting else {
			return
		}
		let json = JSON([
			"cmd": "teamDelay",
			"teamID": team.id,
		])
		WsClient.singleton.sendJSON(json)
	}
	@IBAction func addPlayer(sender: UIButton) {
		guard let team = topTeam where team.size < PLConstants.maxTeamSize else {
			return
		}
		let json = JSON([
			"cmd": "teamAddPlayer",
			"teamID": team.id,
		])
		WsClient.singleton.sendJSON(json)
		sender.enabled = false
	}
	@IBAction func removePlayer(sender: UIButton) {
		guard topTeam != nil && topTeam!.size > 1 else {
			return
		}
		let json = JSON([
			"cmd": "teamRemovePlayer",
			"teamID": topTeam!.id,
		])
		WsClient.singleton.sendJSON(json)
		sender.enabled = false
	}
	@IBAction func ready(sender: UIButton) {
		guard topTeam != nil else {
			return
		}
		if topTeam!.status == .Prepare {
			let json = JSON([
				"cmd": "teamCancelPrepare",
				"teamID": topTeam!.id,
			])
			WsClient.singleton.sendJSON(json)
		} else {
			let json = JSON([
				"cmd": "teamPrepare",
				"teamID": topTeam!.id,
			])
			WsClient.singleton.sendJSON(json)
		}
	}
	@IBAction func start(sender: AnyObject) {
		guard topTeam != nil else {
			return
		}
		var selectedControllerIds = [String]()
		for btn in controllerButtons {
			guard let pc = btn.controller where btn.selected else {
				continue
			}
			selectedControllerIds.append(pc.id)
		}

		let json = JSON([
			"cmd": "teamStart",
			"teamID": topTeam!.id,
			"mode": topTeam!.mode,
			"ids": selectedControllerIds.joinWithSeparator(",")
		])
		WsClient.singleton.sendJSON(json)
		HUD.show(.Progress)
	}

	@IBAction func toggleControllerButton(sender: UIButton) {
		sender.selected = !sender.selected
		startButton.enabled = canStart()
	}

	private func renderTopWaitingTeam() {
		guard topTeam != nil else {
			callButton.enabled = false
			delayButton.enabled = false
			changeModeTGR.enabled = false
			addPlayerButton.enabled = false
			removePlayerButton.enabled = false
			startButton.enabled = false
			readyButton.enabled = false
			return
		}
		readyButton.enabled = true
		teamIDLabel.text = topTeam!.id
		playerNumberLabel.text = "\(topTeam!.size)"
		if topTeam!.mode == "g" {
			modeImageView.image = UIImage(named: "FunIcon")
			modeLabel.text = "[赏金模式]"
		} else {
			modeImageView.image = UIImage(named: "SurvivalIcon")
			modeLabel.text = "[生存模式]"
		}
		if topTeam!.status == .Waiting {
			readyButton.setBackgroundImage(UIImage(named: "PrepareButton"), forState: .Normal)
			callButton.enabled = true
			delayButton.enabled = true
			changeModeTGR.enabled = true
			addPlayerButton.enabled = topTeam!.size < PLConstants.maxTeamSize
			removePlayerButton.enabled = topTeam!.size > 1
		} else if topTeam!.status == .Prepare {
			readyButton.setBackgroundImage(UIImage(named: "CancelPrepare"), forState: .Normal)
			callButton.enabled = false
			delayButton.enabled = false
			changeModeTGR.enabled = false
			addPlayerButton.enabled = false
			removePlayerButton.enabled = false
		}
		startButton.enabled = canStart()
	}

	private func canStart() -> Bool {
		var count = 0
		for btn in controllerButtons {
			if btn.selected {
				count += 1
			}
		}
		return topTeam != nil && topTeam!.status == .Prepare && topTeam!.size == count && !hasPlayingTeam
	}

	private func getBtn(idx: Int) -> UIButton {
		return view.viewWithTag(idx + HallController.controllerButtonTagStart) as! UIButton
	}
}

extension HallController: DataReceiver {
	func onReceivedData(json: [String: AnyObject], type: DataType) {
		if type == .HallData {
			topTeam = nil
			teams = Mapper<Team>().mapArray(json["data"])
			if teams != nil {
				var topTeamSet = false
				hasPlayingTeam = false
				for team in teams! {
					if (team.status == .Waiting || team.status == .Prepare) && !topTeamSet {
						topTeam = team
						topTeamSet = true
					}
					if team.status == .Playing {
						hasPlayingTeam = true
					}
				}
			}
			renderTopWaitingTeam()
			teamtableView.reloadData()
			refreshControl.endRefreshing()
		} else if type == .ControllerData {
			guard let controllers = Mapper<PlayerController>().mapArray(json["data"]) else {
				return
			}
			var controllerMap = [String: PlayerController]()
			for c in controllers {
				if c.online! {
					controllerMap[c.id] = c
				}
			}
			for btn in controllerButtons {
				guard let btnC = btn.controller, let c = controllerMap[btnC.id] else {
					btn.controller = nil
					continue
				}
				btn.controller = c
				controllerMap.removeValueForKey(btnC.id)
			}
			for (_, v) in controllerMap {
				for btn in controllerButtons {
					if btn.controller == nil {
						btn.controller = v
						break
					}
				}
			}
			self.controllers = controllers
		} else if type == .NewMatch {
			HUD.hide()
			for btn in controllerButtons {
				btn.selected = false
			}
		} else if type == .Error {
			HUD.flash(.LabeledError(title: json["msg"] as? String, subtitle: nil), delay: 1)
		}
	}
}

// MARK: swipe function
extension HallController: SWTableViewCellDelegate {
	private var rightButtons: [AnyObject] {
		let jumpButton = UIButton()
		jumpButton.setImage(UIImage(named: "CutLineButton"), forState: .Normal)
		jumpButton.backgroundColor = UIColor.clearColor()
		let removeButton = UIButton()
		removeButton.setImage(UIImage(named: "RemoveTeamButton"), forState: .Normal)
		removeButton.backgroundColor = UIColor.clearColor()
		return [jumpButton, removeButton]
	}

	func swipeableTableViewCell(cell: SWTableViewCell!, didTriggerRightUtilityButtonWithIndex index: Int) {
		guard let team = teamFromCell(cell) else {
			return
		}
		if index == 0 {
			let json = JSON([
				"cmd": "teamCutLine",
				"teamID": team.id
			])
			WsClient.singleton.sendJSON(json)
		} else if index == 1 {
			let json = JSON([
				"cmd": "teamRemove",
				"teamID": team.id
			])
			WsClient.singleton.sendJSON(json)
		}
	}
	func swipeableTableViewCellShouldHideUtilityButtonsOnSwipe(cell: SWTableViewCell!) -> Bool {
		return true
	}
	func swipeableTableViewCell(cell: SWTableViewCell!, canSwipeToState state: SWCellState) -> Bool {
		guard let team = teamFromCell(cell) else {
			return false
		}
		if team.status != .Waiting {
			return false
		}
		return true
	}

	private func teamFromCell(cell: SWTableViewCell) -> Team? {
		if let cellIndex = teamtableView.indexPathForCell(cell), let tms = self.teams {
			return tms[cellIndex.row]
		}
		return nil
	}
}

extension HallController: UITableViewDataSource, UITableViewDelegate {
	func tableView(tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
		return teams != nil ? teams!.count : 0
	}
	func numberOfSectionsInTableView(tableView: UITableView) -> Int {
		return 1
	}
	func tableView(tableView: UITableView, cellForRowAtIndexPath indexPath: NSIndexPath) -> UITableViewCell {
		let cell = tableView.dequeueReusableCellWithIdentifier("HallTableViewCell")! as! HallTableViewCell
		let team = teams![indexPath.row]
		cell.setData(team, number: indexPath.row, active: team.id == topTeam?.id)
		cell.delegate = self
		cell.rightUtilityButtons = rightButtons
		return cell
	}
}
