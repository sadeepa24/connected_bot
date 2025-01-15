package constbot

const TmpConfigInfo string = "confusage"
const TmplOutInfo string = "outinfoc"
const TmpWelcomeBot string = "welcomebot"

const TmpInboundinfo string = "ininfo"
const TmpOutboundinfo string = "outinfo"

// configure
const TmpNameChange string = "confnamechange"      // msg when user change name of the configvia configurecommand, para - {{.Name}}
const TmpConfiConfigure string = "configconfigure" // msg before change config, para - {{.Name}}
const TmplConfigureHome string = "configurehome"

// const Tmp
const TmpInchange string = "confinchange"
const TmpOutchange string = "confoutchange"
const TmpFullInfoConfig string = "fullinfocg"
const TmpUserInfo string = "userinfo"
const TmpConQuota string = "newquota"

// CommandCreate
const (
	TmpCrAlreadyHave string = "cralreadyhave"
	TmpCrAvblQuota   string = "cravblquota"
	TmpCrSendUID     string = "senduuid"
	TmpCrInInfo      string = "crininfo"
	TmpCrOutInfo     string = "croutinfo"
	TmplCrSelect     string = "crselectcreator"
)

// CommandStatus
const (
	TmpStTotal    string = "tmsttota"
	TmpStcallback string = "clbackst"
)

// cap
const (
	TmpcapQuota string = "tmcapquota"
	TmpcapWarn  string = "tmpcapwarn"
	Tmpcapreply string = "tmpcapreply"
)

// Gift
const (
	TmpGifSend  string = "tmpgiftsend"
	TmplRecived string = "tmpgiftrec"
)

// Chatmemupdate
const (
	TmpChatmemLeft         string = "chatmemleft"
	TmpGroupWelcome        string = "grpwelcome"
	TmpWelcomeInbox        string = "dmwelcome"
	TmplInboxVerified      string = "dmverified"
	TmplInboxVerifiedAgain string = "dmverifiedagain"
	TmpGrpComeback         string = "grpcmback"  //user who is not in channel join group
	TmpChanComeback        string = "chancmback" //user who is not in channel join group
	TmpChannelWelcome      string = "chanwelcome"
)

// common
const (
	TmplCommonUnverified string = "unverified"
)

// help
const (
	TmpHelpHome         string = "helphome"
	TmpHelpCmPage       string = "heppage"     // this does not use directly istead TmpPage + pagenum
	TmpHelpInfoPage     string = "helpinfoage" // this does not use directly istead TmpPage + pagenum
	TmplHelpBuilderHelp string = "helpbuilderhelp"
	TmplHelpTuto        string = "helptutorial"
	TmpAbout            string = "botabout"
)

// start
const (
	TmpNewUsers             string = "newusersbot"
	TmpNewUsersVerified     string = "newusersbotverified"
	TmplUserUnverifiedStart string = "newusersbountverified"

	TmpregularVerified string = "regverified"

	TmplMonthLimited string = "monthlimitedstart"
	TmpRemUserst     string = "remuserstart"
)

// refer
const (
	TmpRefHome  string = "refhome"
	TmpRefshare string = "refshare"
)

// distribute
const (
	TmpDisGroup string = "disgroup"
)

// points
const (
	TmplPoints string = "pointsob"
)

// events
const (
	TmplEventHome string = "evehome"
)

const (
	TmplBuilderHome string = "builderhome"
)

const (
	TmplGetinfoHome string = "infohome"
)
