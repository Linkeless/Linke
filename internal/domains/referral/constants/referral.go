package constants

// Referral Status constants
const (
	ReferralStatusPending   = "pending"
	ReferralStatusConfirmed = "confirmed"
	ReferralStatusRewarded  = "rewarded"
	ReferralStatusCancelled = "cancelled"
)

// Referee Status constants
const (
	RefereeStatusRegistered = "registered"
	RefereeStatusActivated  = "activated"
	RefereeStatusSubscribed = "subscribed"
	RefereeStatusChurned    = "churned"
)

// Reward Status constants
const (
	RewardStatusPending   = "pending"
	RewardStatusEarned    = "earned"
	RewardStatusPaid      = "paid"
	RewardStatusCancelled = "cancelled"
	RewardStatusExpired   = "expired"
	RewardStatusRejected  = "rejected"
)

// Referral Source constants
const (
	ReferralSourceInviteCode = "invite_code"
	ReferralSourceLink       = "link"
	ReferralSourceEmail      = "email"
	ReferralSourceSocial     = "social"
	ReferralSourceOrganic    = "organic"
)

// Event Type constants
const (
	EventTypeClick         = "click"
	EventTypeView          = "view"
	EventTypeRegistration  = "registration"
	EventTypeActivation    = "activation"
	EventTypeFirstPurchase = "first_purchase"
	EventTypeSubscription  = "subscription"
	EventTypeRenewal       = "renewal"
	EventTypeCancellation  = "cancellation"
	EventTypeRefund        = "refund"
	EventTypeReward        = "reward"
	EventTypeConversion    = "conversion"
	EventTypeExpired       = "expired"
	EventTypeBlocked       = "blocked"
)

// Campaign Type constants
const (
	CampaignTypeStandard   = "standard"
	CampaignTypeBonus      = "bonus"
	CampaignTypeSeasonal   = "seasonal"
	CampaignTypeInfluencer = "influencer"
	CampaignTypePartner    = "partner"
)

// Campaign Status constants
const (
	CampaignStatusActive = "active"
	CampaignStatusPaused = "paused"
	CampaignStatusEnded  = "ended"
)

// Reward Type constants (for campaigns)
const (
	RewardTypeFixed        = "fixed"
	RewardTypePercentage   = "percentage"
	RewardTypeTiered       = "tiered"
	RewardTypeDiscount     = "discount"
	RewardTypeCash         = "cash"
	RewardTypeCredit       = "credit"
	RewardTypeBonus        = "bonus"
	RewardTypeCoupon       = "coupon"
	RewardTypeSubscription = "subscription"
	RewardTypeProduct      = "product"
	RewardTypeService      = "service"
)

// Reward Trigger constants
const (
	RewardTriggerRegistration  = "registration"
	RewardTriggerFirstPurchase = "first_purchase"
	RewardTriggerSubscription  = "subscription"
	RewardTriggerActivation    = "activation"
)

// Invite Code Status constants
const (
	InviteCodeStatusActive   = "active"
	InviteCodeStatusUsed     = "used"
	InviteCodeStatusDisabled = "disabled"
)

// Payment Method constants
const (
	PaymentMethodPayPal       = "paypal"
	PaymentMethodStripe       = "stripe"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCredit       = "credit"
	PaymentMethodCrypto       = "crypto"
	PaymentMethodCheck        = "check"
)

// Default values
const (
	DefaultCurrency       = "CNY"
	DefaultMaxUsesPerCode = 10
	DefaultRewardAmount   = 0.0
	DefaultMaxReferrals   = 0
	DefaultMaxRewards     = 0
)
