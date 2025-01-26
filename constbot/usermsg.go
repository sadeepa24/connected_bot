package constbot

import (
	"encoding/json"
	"os"
)

var (
	MsgserverErr string = "🚨 Something went wrong on the server. Please try again later."

	// Configure //
	MsgNoconfigstochange string = "⚙️ You don't have any configurations to modify. Please use the /create command to set up your configurations."
	//MsgConfChoose        string = "🛠️ Please choose the configuration you want to edit."
	MsgNewName          string = "✏️ Enter a new name for the configuration."
	MsgInAlredSelected  string = "Already This Inbound selected"
	MsgOutAlredSelected string = "Already This OUtbound selected"
	// Name change //
	MsgNamechangeSuc    string = "✅ Configuration name changed successfully!"
	MsgNameChangeFailed string = "❌ Failed to change the configuration name."

	// Change inbound //
	MsgInsel          string = "📥 Select the inbound configuration to change."
	MsgInchanGeWarn   string = "⚠️ Warning: Changing the inbound configuration will interrupt your connection. Ensure you update your settings accordingly. Proceed only if you know what you're doing."
	MsgInchangesucses string = "✅ Inbound configuration successfully changed!"

	// Change outbound //
	Msgoutsel         string = "📤 Select the outbound configuration to change."
	MsgOutchangeWar   string = "⚠️ Warning: Changing the outbound configuration will modify your IP address."
	MsOutchangesucses string = "✅ Outbound configuration successfully changed!"

	// Delete configuration //
	MsgdelConnWarn string = "⚠️ Your connection will be closed."
	MsgSure        string = "❓ Are you sure? Your connection will be lost."
	MsgdelFail     string = "❌ Failed to delete the configuration."
	MsgdelSuccses  string = "✅ Configuration deleted successfully!"

	// Quota //
	MsgCoQuota   string = "✅ Configuration quota successfully updated!"
	MsgQuotawarn string = "⚠️ Your quota must be within the specified range."
	MsgQuotawarnlow string = "⚠️ Your quota must be greater than you'r current usage"
	MsgQuotawarnzero string = "⚠️ Value should be greater than zero"
)

var (
	// Common
	MsgSessionExcist   string = "⚠️ A session already exists. Please close it first."
	Msgwrong           string = "❌ Something went wrong. Please try again."
	MsgwrongtAdmmin    string = "❌ Something went wrong. Please retry, or contact the admin with the error."
	MsgConfUnfoun      string = "⚙️ Configuration not found. Please retry."
	MsgSessionOver     string = "⏳ Your session has ended. Please try again."
	MsgSessionFail     string = "❌ Session creation failed. Please try again later."
	MsgContextDead     string = "❌ Context canceled. Please try again."
	MsgValidName       string = "🔑 Please send a valid name, no commands allowed."
	MsgValidInt        string = "🔢 Please send a valid integer."
	Msgretryfail       string = "🚫 මෝඩයෙක් බව පෙන්නනන හදන්න එපා"
	MsgRecursionExceed string = "⚠️ Too many button presses. Please do what you need to do. (දකින දකින එක  ඔබන්න එප රිලවෙක් වගේ)"
	MsgDberr           string = "❌ A database error occurred. Please try again later."
	MsgUserNotFoun     string = "👤 User not found. The user may not have registered with the bot."
	MsgTargetcapped    string = "🚫 Target user is a capped user."

	MsgNotCmdDIs		string = ""

	// Others
	ButtonSelectEmjoi   string = " ✅"
	MsgUserMonthLimited string = "🚫 You can't use this service until your punishment period is over."
)

var (
	//MsgInfoStart        string = "👋 Hello! How are you? What would you like to do?"
	MsgInfoNoconfigs    string = "⚙️ You don't have any configurations."
	MsgInfoSelectConfig string = "🔧 Select a configuration to reveal its information."

	// Configs
	Msgconfcannotfind string = "❌ Configuration not found. Something went wrong. You may need to notify the admin if this continues."
	MsgfetchUsage     string = "⏳ Please wait while we fetch your usage history. This may take a moment."
)

// Command Create
var (
	MsgVerifiedUser string = "🔒 You need to be a verified user first. Please join our group and channel, then try again."
	MsgUsageExceed  string = "❌ You can create configurations, but you will not be able to use them as your quota is exceeded."
	MsgselectIn     string = "📥 Select an inbound configuration to create. You can change the inbound later."
	MsgselectOut    string = "📤 Select an outbound configuration to create. You can change the outbound later. Tip: Users should choose the default outbound unless they have special cases. Use the /help command for more information."
	MsgnoQuota      string = "⚠️ You don't have available quota to add to this configuration. Please change the quota of another configuration."
	MsgGetName      string = "🔑 Please provide a name for the configuration."
	MsgCrFailed     string = "❌ Configuration creation failed. Please try again later."
	MsgInternalErr  string = "⚠️ Internal VPN server error. You may need to contact the admin."
	MsgGetSni       string = "🔑 Please provide your SNI (you can change this later yourself)."
	MsgCrsuccsess   string = "✅ You have successfully created the configuration."
	MsgCrLogin      string = "🔑 Please specify how many users can log in at once (max 5). Example: If you select 1, only one IP address can connect at a time."
	MsgCrLoginwarn  string = "⚠️ Login limit should be between 0 and 5."
	MsgSnifail      string = "❌ SNI reception failed while creating the configuration. You can compile your config into multiple blocks using /confblocks."
	MsgCrConfigIn   string = "🛠️ You can create configuration blocks using the /" + CmdConfigBlocks + " command."
	MsgCrdisuser    string = "⚠️ You can't create configurations at the moment. You are a restricted user."

	MsgCrInerr  string = "❌ The selected inbound configuration has a fatal error."
	MsgCrOuterr string = "❌ The selected outbound configuration has a fatal error."

	MsgCrQuotaNote string = "⚠️ Your current quota may be higher than what you can add to this configuration. This happens because your total usage for the month doesn't match the usage of existing configurations, which could be due to the deletion of a configuration."

	MsgNoQuota = "You Don't Have Available Quota To create a new config, if you want to create a new config, you can change the quota of the existing config or delete"
)

// Command Status
var (
	MsgStVerify   string = "🔒 You are not a verified user. In order to see the status or access any Singbox services, you need to be a verified user."
	MsgStNoconfig string = "⚙️ You don't have any configurations. To get usage, please create configs using the /create command."
)

var (
	Msgxrayuse string = "🔒 You need to be a verified user to use Xray services."
)

// User Cap
var (
	Msgcapverify  string = "🔒 You need to be a verified user in order to distribute your bandwidth."
	MsgcapAlready string = "⚠️ You are already capped. You can't cap again. Please wait until your cap time limit is over."
	Msgcapexced   string = "❌ You can't cap your quota because you've already used all of your quota."
	Msgcapzerod   string = "⚠️ The cap cannot be zero. If you want to share your total bandwidth, please use the /distribute command."
	MsgcapThan    string = "⚠️ Please enter a value lower than your available quota."
	MsgcapConform string = "⚠️ You are about to cap your quota. This quota will be valid for the next 30 days."
	MsgcapCancle  string = "❌ Capping your quota has been canceled."
	MsgcapSuccses string = "✅ You have successfully capped your quota."

	MsgcapUsage string = "⚠️ You have already used the cap you entered. Please enter a cap higher than your usage."

	MsgcapRecalFail string = "⚠️ Recalculation Failed, You'r quota will update in next db refresh"
)

// User Distribute
var (
	MsgDisAlready    string = "⚠️ You are already a distributed user. You can't distribute again."
	MsgDisneedVerify string = "🔒 You need to be a verified user in order to distribute your bandwidth."
	MsgDisConform    string = "⚠️ Are you sure? You are about to distribute all of your quota."
	MsgDisSucsess    string = "✅ You have successfully distributed your quota. Thank you!"
	MsgDisCapped     string = "⚠️ You are capped user. You can't distribute"
)

// User Conform
var (
	MsgGifVerify      string = "🔒 You need to be a verified user in order to send gifts."
	MsgGifUsercap     string = "⚠️ You can't send gifts as you are a capped user. Please wait until your cap is over."
	MsgGifrec         string = "🎁 You have received a gift from someone, so you can't send one right now."
	MsgGifsend        string = "⚠️ You have already sent a gift. You can't send any more gifts until 30 days have passed."
	Msggifterr        string = "❌ You can't gift more than what you have."
	MsgGifreciver     string = "🎁 Alright, now send me the recipient's Telegram ID or username. If you want to cancel, send /cancel."
	MsgGiftcancle     string = "❌ Sending the gift has been canceled."
	MsgGifRecnOconfig string = "⚠️ The target user does not have any created config. They need to have configs in order to receive gifts."
)

var (
	MsgChatMemLeft string = ""
	MsgBannedMem   string = "👋 Bye Bye!"
)

// Help
var (
	Msghelpnoverify string = "🔒 This service is only available for verified users."
	MsgCallbackFaq  string = "💬 Ehema pasna na thama." // Assuming this is intentional in a local language; if not, let me know for clarification.
	MsgHeloClosed   string = "❌ Help is closed."
)

// Start
var (
	//MsgUserRemoved   string = "You are a removed user. If you want to use the connected service again, please join with us."
	MsgsttInChan     string = "🔔 You are already in the channel, but to use this bot, you may need to join the group."
	MsgstartGrpin    string = "🔔 You are already in our group. To use this service, you need to join our channel."
	Msgstartmlimited string = "⏳ You didn't use 3/4 of your quota from last month, so you can't use the service for the next 30 days."
	MsgBannedUser    string = "🚫 You are a banned user. Please contact the admin to be unbanned."

	// Referral handling
	MsgSelfRef        string = "❌ You can't be your own referral."
	MsgRefOwenerNFoun string = "⚠️ Something went wrong fetching the owner of the referral. The referral owner may not be registered."
	Msgcanref         string = "📢 You can also refer users and earn rewards. Use /refer for more info."
	MsgRefAlredy      string = "you cant be a reffred user, you are already reffred you are refred from user id %v"
	MsgReferd         string = "🎉 Welcome! You’re now a user who came from %v 's referral. 🌟 Let’s get started on this exciting journey together! 🚀"
)

// Refer
var (
	MsgRefVerify       string = "🔒 You can't use this command unless you're a verified user. Please join the channel and group, then try again."
	MsgRefNousers      string = "⚠️ You don't have any referred users yet. Hurry up and refer users!"
	MsgRefNoUser       string = "❌ You don't have any verified referred users to claim. Please ask your referred user to verify, so you can claim your points."
	MsgRefClaimNote    string = "ℹ️ You can claim referrals as points: Verified user = 2 points, Normal user = 1 point."
	MsgRefClaimConform string = "✅ Confirm to claim your referral points."
	MsgRefClaimError   string = "❌ Error processing claim. Please try again."

	MsgRefClaimed           string = "🎉 You have claimed %v points."
	MSgRefClaimAllunsupport string = "⚠️ Claim all feature is not supported yet."
	MsgClaimCancle          string = "❌ Claim canceled."
	MsgRefLink              string = "🔗 Your referral link is: %v"
	MsgRefNoANyUser         string = "⚠️ You don't have any referred users."
)

// Suggest
var (
	MsgSugess     string = "💡 Please share your suggestion! We’d love to hear your thoughts and ideas. 😊"
	Msgsugessdone string = "✅ Suggestion submitted! 📩 It will be sent to the admin for review. Thank you for your input! 🙏"
)

// Watchman
var (
	MsgQuotanotUsed      string = "⚠️ You have not utilized 75% of your previous quota. As a result, access to the service is suspended for the next 30 days. 🚫 Please plan your usage wisely in the future! 📊"
	MsgDistributeOver    string = "✨ You are no longer part of the distributed users. If you'd like to share your quota again, simply use the /distribute command. 🌟 Thank you for your valuable contribution and support! 🙏"
	MsgwtchErrinnotfound string = "⚠️ Your configuration encountered an error during the database refresh. Please reach out to the developer for assistance. 🛠️ Error: Inbound not found."

	MsgwtchErrtypemiss string = "⚠️ Your configuration encountered an error during the database refresh. Please contact the developer for assistance. 🛠️ Error: Inbound type mismatch."
	MsgwtchErruseradd  string = "⚠️ Your configuration encountered an error during the database refresh. Please contact the developer for assistance. 🛠️ Error: VLESS service error during user addition."

	MsgwtchUsagereset string = "🔄 All your usage has been reset, but any excess usage has been carried over to this month. 📊"
	MsgresetUsage     string = "🔄 All your usage has been successfully reset. ✨"
)

var (
	MsgGiftSent string = "🎁 You have successfully sent a gift of %v to %v. 🌟 Enjoy sharing the love!"
)

// callbackDefaul
var (
	MsgBtnOffline string = "⚠️ It looks like the button is no longer online. Please restart the command 🔌"
)

// Contact
var (
	MsgContactTimeover string = "✨ Your contact time has expired. ⏳ Please feel free to reach out again when you're ready! 😊"
	MsgContactCancle   string = "❌ You have canceled the contact session. If you need assistance later, don't hesitate to reach out! 😊"
)

type UserMsg map[string]string

func GetMsg(inmg string) string {
	//return mg
	mg, ok := AllUserMsg[inmg]
	if !ok {
		return "if you seen this msg please contact admin and tell he missed " + inmg
	}
	return mg
}

var AllUserMsg UserMsg

func LoadUserMsg() error {
	var err error
	overide()
	AllUserMsg, err = newUserMsg("usermsg.json")
	return err
}
func newUserMsg(path string) (UserMsg, error) {
	file, err := os.ReadFile(path)
	var usermg UserMsg
	if err != nil {
		return usermg, err
	}
	err = json.Unmarshal(file, &usermg)
	return usermg, err

}

func overide() {

	MsgserverErr = "MsgserverErr"
	MsgNoconfigstochange = "MsgNoconfigstochange"
	MsgNewName = "MsgNewName"
	MsgInAlredSelected = "MsgInAlredSelected"
	MsgOutAlredSelected = "MsgOutAlredSelected"
	MsgNamechangeSuc = "MsgNamechangeSuc"
	MsgNameChangeFailed = "MsgNameChangeFailed"
	MsgInsel = "MsgInsel"
	MsgInchanGeWarn = "MsgInchanGeWarn"
	MsgInchangesucses = "MsgInchangesucses"
	Msgoutsel = "Msgoutsel"
	MsgOutchangeWar = "MsgOutchangeWar"
	MsOutchangesucses = "MsOutchangesucses"
	MsgdelConnWarn = "MsgdelConnWarn"
	MsgSure = "MsgSure"
	MsgdelFail = "MsgdelFail"
	MsgdelSuccses = "MsgdelSuccses"
	MsgCoQuota = "MsgCoQuota"
	MsgQuotawarn = "MsgQuotawarn"
	MsgSessionExcist = "MsgSessionExcist"
	Msgwrong = "Msgwrong"
	MsgwrongtAdmmin = "MsgwrongtAdmmin"
	MsgConfUnfoun = "MsgConfUnfoun"
	MsgSessionOver = "MsgSessionOver"
	MsgSessionFail = "MsgSessionFail"
	MsgContextDead = "MsgContextDead"
	MsgValidName = "MsgValidName"
	MsgValidInt = "MsgValidInt"
	Msgretryfail = "Msgretryfail"
	MsgRecursionExceed = "MsgRecursionExceed"
	MsgDberr = "MsgDberr"
	MsgUserNotFoun = "MsgUserNotFoun"
	MsgTargetcapped = "MsgTargetcapped"
	ButtonSelectEmjoi = "ButtonSelectEmjoi"
	MsgUserMonthLimited = "MsgUserMonthLimited"
	MsgInfoNoconfigs = "MsgInfoNoconfigs"
	MsgInfoSelectConfig = "MsgInfoSelectConfig"
	Msgconfcannotfind = "Msgconfcannotfind"
	MsgfetchUsage = "MsgfetchUsage"
	MsgVerifiedUser = "MsgVerifiedUser"
	MsgUsageExceed = "MsgUsageExceed"
	MsgselectIn = "MsgselectIn"
	MsgselectOut = "MsgselectOut"
	MsgnoQuota = "MsgnoQuota"
	MsgGetName = "MsgGetName"
	MsgCrFailed = "MsgCrFailed"
	MsgInternalErr = "MsgInternalErr"
	MsgGetSni = "MsgGetSni"
	MsgCrsuccsess = "MsgCrsuccsess"
	MsgCrLogin = "MsgCrLogin"
	MsgCrLoginwarn = "MsgCrLoginwarn"
	MsgSnifail = "MsgSnifail"
	MsgCrConfigIn = "MsgCrConfigIn"
	MsgCrdisuser = "MsgCrdisuser"
	MsgCrInerr = "MsgCrInerr"
	MsgCrOuterr = "MsgCrOuterr"
	MsgCrQuotaNote = "MsgCrQuotaNote"
	MsgStVerify = "MsgStVerify"
	MsgStNoconfig = "MsgStNoconfig"
	Msgxrayuse = "Msgxrayuse"
	Msgcapverify = "Msgcapverify"
	MsgcapAlready = "MsgcapAlready"
	Msgcapexced = "Msgcapexced"
	Msgcapzerod = "Msgcapzerod"
	MsgcapThan = "MsgcapThan"
	MsgcapConform = "MsgcapConform"
	MsgcapCancle = "MsgcapCancle"
	MsgcapSuccses = "MsgcapSuccses"
	MsgcapUsage = "MsgcapUsage"
	MsgcapRecalFail = "MsgcapRecalFail"
	MsgDisAlready = "MsgDisAlready"
	MsgDisneedVerify = "MsgDisneedVerify"
	MsgDisConform = "MsgDisConform"
	MsgDisSucsess = "MsgDisSucsess"
	MsgDisCapped = "MsgDisCapped"
	MsgGifVerify = "MsgGifVerify"
	MsgGifUsercap = "MsgGifUsercap"
	MsgGifrec = "MsgGifrec"
	MsgGifsend = "MsgGifsend"
	Msggifterr = "Msggifterr"
	MsgGifreciver = "MsgGifreciver"
	MsgGiftcancle = "MsgGiftcancle"
	MsgGifRecnOconfig = "MsgGifRecnOconfig"
	MsgChatMemLeft = "MsgChatMemLeft"
	MsgBannedMem = "MsgBannedMem"
	Msghelpnoverify = "Msghelpnoverify"
	MsgCallbackFaq = "MsgCallbackFaq"
	MsgHeloClosed = "MsgHeloClosed"
	MsgsttInChan = "MsgsttInChan"
	MsgstartGrpin = "MsgstartGrpin"
	Msgstartmlimited = "Msgstartmlimited"
	MsgBannedUser = "MsgBannedUser"
	MsgSelfRef = "MsgSelfRef"
	MsgRefOwenerNFoun = "MsgRefOwenerNFoun"
	Msgcanref = "Msgcanref"
	MsgRefAlredy = "MsgRefAlredy"
	MsgReferd = "MsgReferd"
	MsgRefVerify = "MsgRefVerify"
	MsgRefNousers = "MsgRefNousers"
	MsgRefNoUser = "MsgRefNoUser"
	MsgRefClaimNote = "MsgRefClaimNote"
	MsgRefClaimConform = "MsgRefClaimConform"
	MsgRefClaimError = "MsgRefClaimError"
	MsgRefClaimed = "MsgRefClaimed"
	MSgRefClaimAllunsupport = "MSgRefClaimAllunsupport"
	MsgClaimCancle = "MsgClaimCancle"
	MsgRefLink = "MsgRefLink"
	MsgRefNoANyUser = "MsgRefNoANyUser"
	MsgSugess = "MsgSugess"
	Msgsugessdone = "Msgsugessdone"
	MsgQuotanotUsed = "MsgQuotanotUsed"
	MsgDistributeOver = "MsgDistributeOver"
	MsgwtchErrinnotfound = "MsgwtchErrinnotfound"
	MsgwtchErrtypemiss = "MsgwtchErrtypemiss"
	MsgwtchErruseradd = "MsgwtchErruseradd"
	MsgwtchUsagereset = "MsgwtchUsagereset"
	MsgresetUsage = "MsgresetUsage"
	MsgGiftSent = "MsgGiftSent"
	MsgBtnOffline = "MsgBtnOffline"
	MsgContactCancle = "MsgContactCancle"
	MsgContactTimeover = "MsgContactTimeover"
	MsgNotCmdDIs = "MsgNotCmdDIs"

	MsgQuotawarnlow = "MsgQuotawarnlow"
	MsgQuotawarnzero = "MsgQuotawarnzero"
	MsgNoQuota = "MsgNoQuota"

}
