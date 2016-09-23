//
//  SecondViewController.swift
//  admin
//
//  Created by tassar on 4/20/16.
//  Copyright © 2016 pulupulu. All rights reserved.
//

import UIKit
import Alamofire
import AlamofireImage
import EasyPeasy
import ObjectMapper
import SwiftyJSON
import SwiftyUserDefaults

let cellSize = 45
let cellBorder = 10

class MatchController: PLViewController {
	@IBOutlet weak var groupIDLabel: UILabel!
	@IBOutlet weak var matchStatusLabel: UILabel!
	@IBOutlet weak var playerCountLabel: UILabel!
	@IBOutlet weak var totalCoinLabel: UILabel!
	@IBOutlet weak var energyLabel: UILabel!
	@IBOutlet weak var matchTimeLabel: UILabel!
	@IBOutlet weak var matchModeImageView: UIImageView!
	@IBOutlet weak var mapContainerView: UIView!
	@IBOutlet weak var playerTableView: UITableView!

	var match: Match?
	var playerViews: [UIButton]!
	var laserViews: [UIView]!

	var mapView: UIImageView = UIImageView()

	@IBAction func forceEnd() {
		let json = JSON([
			"cmd": "stopMatch",
			"matchID": Defaults[.matchID]
		])
		WsClient.singleton.sendJSON(json)
	}

	override func viewDidLoad() {
		super.viewDidLoad()
		playerTableView.backgroundColor = UIColor.clearColor()
		playerViews = [UIButton]()
		laserViews = [UIView]()
		for _ in 1 ... 4 {
			let btn = UIButton()
			btn.frame = CGRect(x: 0, y: 0, width: 30, height: 30)
			btn.setBackgroundImage(UIImage(named: "PlayerIcon"), forState: .Normal)
			btn.setTitleColor(UIColor.whiteColor(), forState: .Normal)
			btn.titleLabel?.font = UIFont.systemFontOfSize(6)
			btn.hidden = true
			playerViews.append(btn)
			mapView.addSubview(btn)
		}
	}

	override func viewWillAppear(animated: Bool) {
		super.viewWillAppear(animated)
		DataManager.singleton.subscribeData([.UpdateMatch, .MatchStop], receiver: self)
		if mapView.image == nil {
			Alamofire.request(.GET, PLConstants.getHttpAddress("api/asset/map.png"))
				.validate()
				.responseImage(completionHandler: { response in
					if let image = response.result.value {
						self.mapView.image = image
						self.mapContainerView.addSubview(self.mapView)
						self.mapView <- [
							Size(image.size),
							Center()
						]
					}
			})
		}
	}

	override func viewDidDisappear(animated: Bool) {
		super.viewDidDisappear(animated)
		DataManager.singleton.unsubscribe(self)
	}

	func renderMatch() {
		if match != nil && match!.id == Defaults[.matchID] {
			groupIDLabel.text = match!.teamID
			matchModeImageView.image = match!.mode == "g" ? UIImage(named: "FunIcon") : UIImage(named: "SurvivalIcon")
			let min = Int(match!.elasped) / 60
			let sec = Int(match!.elasped) % 60
			matchTimeLabel.text = String(format: "%02d:%02d", min, sec)
			matchStatusLabel.text = "实时状态: 进行中"
			playerCountLabel.text = "玩家人数:\(match!.member.count)"
			totalCoinLabel.text = "总金币:\(match!.gold)G"
			energyLabel.text = String(format: "%.1f/%d", match!.energy, match!.maxEnergy)
			playerTableView.reloadData()
			if match!.isSimulator == 0 {
				for view in laserViews {
					view.removeFromSuperview()
				}
				laserViews.removeAll()
				if match!.lasers != nil {
					for laser in match!.lasers! {
						let view = UIView()
						view.bounds = CGRect(x: 0, y: 0, width: 40, height: 40)
						view.center = CGPoint(x: laser.displayP.X / 3, y: laser.displayP.Y / 3)
						view.backgroundColor = UIColor.greenColor()
						mapView.addSubview(view)
						laserViews.append(view)
						if laser.displayP2.X >= 0 {
							let view = UIView()
							view.bounds = CGRect(x: 0, y: 0, width: 40, height: 40)
							view.center = CGPoint(x: laser.displayP2.X / 3, y: laser.displayP2.Y / 3)
							view.backgroundColor = UIColor.greenColor()
							mapView.addSubview(view)
							laserViews.append(view)
						}
					}
				}
			}
			for (i, btn) in playerViews.enumerate() {
				if i < match!.member.count {
					let player = match!.member[i]
					btn.hidden = false
					btn.center = CGPoint(x: player.displayPos.X / 3, y: player.displayPos.Y / 3)
					btn.setTitle(player.displayID, forState: .Normal)
					mapView.bringSubviewToFront(btn)
				} else {
					btn.hidden = true
				}
			}
		} else {
			matchTimeLabel.text = "00: 00"
			matchStatusLabel.text = "实时状态: 未进行"
			playerCountLabel.text = "玩家人数: 0"
			totalCoinLabel.text = "总金币:0G"
			energyLabel.text = ""
			playerTableView.reloadData()
			for btn in playerViews {
				btn.hidden = true
			}
			for view in laserViews {
				view.removeFromSuperview()
			}
			laserViews.removeAll()
		}
		for btn in playerViews {
			log.debug("\(btn)")
		}
	}
}

extension MatchController: DataReceiver {
	func onReceivedData(json: [String: AnyObject], type: DataType) {
		if type == .UpdateMatch {
			match = Mapper<Match>().map(json["data"] as! String)
			if match != nil && match?.id == Defaults[.matchID] {
				renderMatch()
			}
		} else if type == .MatchStop {
			let matchResult = Mapper<MatchResult>().map(json["data"])
			if matchResult != nil {
				if matchResult?.matchID == Defaults[.matchID] {
					match = nil
					renderMatch()
				}
			}
		}
	}
}

extension MatchController: UITableViewDelegate, UITableViewDataSource {
	func tableView(tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
		if match == nil {
			return 0
		} else {
			return match!.member.count
		}
	}

	func numberOfSectionsInTableView(tableView: UITableView) -> Int {
		return 1
	}

	func tableView(tableView: UITableView, cellForRowAtIndexPath indexPath: NSIndexPath) -> UITableViewCell {
		let cell = tableView.dequeueReusableCellWithIdentifier("PlayerTableViewCell") as! PlayerTableViewCell
		cell.setData(match!.member[indexPath.row])
		return cell
	}
}
