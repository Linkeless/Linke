package cache

import (
	"fmt"
	"strings"
)

type DefaultKeyBuilder struct {
	separator string
}

func NewKeyBuilder() CacheKeyBuilder {
	return &DefaultKeyBuilder{
		separator: ":",
	}
}

func (kb *DefaultKeyBuilder) Build(parts ...string) string {
	return strings.Join(parts, kb.separator)
}

func (kb *DefaultKeyBuilder) BuildWithPrefix(prefix string, parts ...string) string {
	allParts := append([]string{prefix}, parts...)
	return kb.Build(allParts...)
}

func (kb *DefaultKeyBuilder) ExtractPattern(key string) string {
	parts := strings.Split(key, kb.separator)
	if len(parts) > 0 {
		return parts[0] + kb.separator + "*"
	}
	return key + "*"
}

type UserCacheKeys struct {
	kb CacheKeyBuilder
}

func NewUserCacheKeys(kb CacheKeyBuilder) *UserCacheKeys {
	return &UserCacheKeys{kb: kb}
}

func (uck *UserCacheKeys) UserByID(userID uint) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, "id", fmt.Sprintf("%d", userID))
}

func (uck *UserCacheKeys) UserByEmail(email string) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, "email", email)
}

func (uck *UserCacheKeys) UserByUsername(username string) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, "username", username)
}

func (uck *UserCacheKeys) UserProfile(userID uint) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, "profile", fmt.Sprintf("%d", userID))
}

func (uck *UserCacheKeys) UserPermissions(userID uint) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, "permissions", fmt.Sprintf("%d", userID))
}

func (uck *UserCacheKeys) UserPattern(userID uint) string {
	return uck.kb.BuildWithPrefix(CachePrefixUser, fmt.Sprintf("%d", userID), "*")
}

type SubscriptionCacheKeys struct {
	kb CacheKeyBuilder
}

func NewSubscriptionCacheKeys(kb CacheKeyBuilder) *SubscriptionCacheKeys {
	return &SubscriptionCacheKeys{kb: kb}
}

func (sck *SubscriptionCacheKeys) PlanByID(planID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixPlan, "id", fmt.Sprintf("%d", planID))
}

func (sck *SubscriptionCacheKeys) PlanList() string {
	return sck.kb.BuildWithPrefix(CachePrefixPlan, "list", "all")
}

func (sck *SubscriptionCacheKeys) ActivePlans() string {
	return sck.kb.BuildWithPrefix(CachePrefixPlan, "active", "all")
}

func (sck *SubscriptionCacheKeys) UserSubscription(userID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixSubscription, "user", fmt.Sprintf("%d", userID))
}

func (sck *SubscriptionCacheKeys) UserActiveSubscriptions(userID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixSubscription, "user", fmt.Sprintf("%d", userID), "active")
}

func (sck *SubscriptionCacheKeys) OrderByID(orderID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixSubscription, "order", fmt.Sprintf("%d", orderID))
}

func (sck *SubscriptionCacheKeys) OrderByNumber(orderNumber string) string {
	return sck.kb.BuildWithPrefix(CachePrefixSubscription, "order", "number", orderNumber)
}

func (sck *SubscriptionCacheKeys) UserOrders(userID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixSubscription, "user", fmt.Sprintf("%d", userID), "orders")
}

type PaymentCacheKeys struct {
	kb CacheKeyBuilder
}

func NewPaymentCacheKeys(kb CacheKeyBuilder) *PaymentCacheKeys {
	return &PaymentCacheKeys{kb: kb}
}

func (pck *PaymentCacheKeys) PaymentByNo(paymentNo string) string {
	return pck.kb.BuildWithPrefix(CachePrefixPayment, "no", paymentNo)
}

func (pck *PaymentCacheKeys) PaymentByTransactionID(txnID string) string {
	return pck.kb.BuildWithPrefix(CachePrefixPayment, "txn", txnID)
}

func (pck *PaymentCacheKeys) UserPayments(userID uint) string {
	return pck.kb.BuildWithPrefix(CachePrefixPayment, "user", fmt.Sprintf("%d", userID))
}

func (pck *PaymentCacheKeys) PaymentMethods() string {
	return pck.kb.BuildWithPrefix(CachePrefixPayment, "methods", "all")
}

func (pck *PaymentCacheKeys) IdempotencyKey(gateway, key string) string {
	return pck.kb.BuildWithPrefix(CachePrefixPayment, "idempotency", gateway, key)
}

type AuthCacheKeys struct {
	kb CacheKeyBuilder
}

func NewAuthCacheKeys(kb CacheKeyBuilder) *AuthCacheKeys {
	return &AuthCacheKeys{kb: kb}
}

func (ack *AuthCacheKeys) SessionToken(token string) string {
	return ack.kb.BuildWithPrefix(CachePrefixSession, token)
}

func (ack *AuthCacheKeys) UserSessions(userID uint) string {
	return ack.kb.BuildWithPrefix(CachePrefixSession, "user", fmt.Sprintf("%d", userID))
}

func (ack *AuthCacheKeys) RefreshToken(token string) string {
	return ack.kb.BuildWithPrefix(CachePrefixAuth, "refresh", token)
}

func (ack *AuthCacheKeys) OAuthState(state string) string {
	return ack.kb.BuildWithPrefix(CachePrefixAuth, "oauth", "state", state)
}

func (ack *AuthCacheKeys) PasswordResetToken(token string) string {
	return ack.kb.BuildWithPrefix(CachePrefixAuth, "reset", token)
}

func (ack *AuthCacheKeys) EmailVerificationToken(token string) string {
	return ack.kb.BuildWithPrefix(CachePrefixAuth, "verify", token)
}

type ServerCacheKeys struct {
	kb CacheKeyBuilder
}

func NewServerCacheKeys(kb CacheKeyBuilder) *ServerCacheKeys {
	return &ServerCacheKeys{kb: kb}
}

func (sck *ServerCacheKeys) ServerByID(serverID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixServer, "id", fmt.Sprintf("%d", serverID))
}

func (sck *ServerCacheKeys) ServerList() string {
	return sck.kb.BuildWithPrefix(CachePrefixServer, "list", "all")
}

func (sck *ServerCacheKeys) ActiveServers() string {
	return sck.kb.BuildWithPrefix(CachePrefixServer, "active", "all")
}

func (sck *ServerCacheKeys) ServerStats(serverID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixServer, "stats", fmt.Sprintf("%d", serverID))
}

func (sck *ServerCacheKeys) UserServerAccess(userID uint) string {
	return sck.kb.BuildWithPrefix(CachePrefixServer, "user", fmt.Sprintf("%d", userID), "access")
}

type CouponCacheKeys struct {
	kb CacheKeyBuilder
}

func NewCouponCacheKeys(kb CacheKeyBuilder) *CouponCacheKeys {
	return &CouponCacheKeys{kb: kb}
}

func (cck *CouponCacheKeys) CouponByCode(code string) string {
	return cck.kb.BuildWithPrefix(CachePrefixCoupon, "code", code)
}

func (cck *CouponCacheKeys) CouponByID(couponID uint) string {
	return cck.kb.BuildWithPrefix(CachePrefixCoupon, "id", fmt.Sprintf("%d", couponID))
}

func (cck *CouponCacheKeys) ActiveCoupons() string {
	return cck.kb.BuildWithPrefix(CachePrefixCoupon, "active", "all")
}

func (cck *CouponCacheKeys) UserCoupons(userID uint) string {
	return cck.kb.BuildWithPrefix(CachePrefixCoupon, "user", fmt.Sprintf("%d", userID))
}

type InvoiceCacheKeys struct {
	kb CacheKeyBuilder
}

func NewInvoiceCacheKeys(kb CacheKeyBuilder) *InvoiceCacheKeys {
	return &InvoiceCacheKeys{kb: kb}
}

func (ick *InvoiceCacheKeys) InvoiceByID(invoiceID uint) string {
	return ick.kb.BuildWithPrefix(CachePrefixInvoice, "id", fmt.Sprintf("%d", invoiceID))
}

func (ick *InvoiceCacheKeys) InvoiceByNumber(invoiceNumber string) string {
	return ick.kb.BuildWithPrefix(CachePrefixInvoice, "number", invoiceNumber)
}

func (ick *InvoiceCacheKeys) UserInvoices(userID uint) string {
	return ick.kb.BuildWithPrefix(CachePrefixInvoice, "user", fmt.Sprintf("%d", userID))
}

func (ick *InvoiceCacheKeys) InvoiceByOrderID(orderID uint) string {
	return ick.kb.BuildWithPrefix(CachePrefixInvoice, "order", fmt.Sprintf("%d", orderID))
}

type RateLimitCacheKeys struct {
	kb CacheKeyBuilder
}

func NewRateLimitCacheKeys(kb CacheKeyBuilder) *RateLimitCacheKeys {
	return &RateLimitCacheKeys{kb: kb}
}

func (rlk *RateLimitCacheKeys) IPRateLimit(ip, endpoint string) string {
	return rlk.kb.BuildWithPrefix(CachePrefixRateLimit, "ip", ip, endpoint)
}

func (rlk *RateLimitCacheKeys) UserRateLimit(userID uint, endpoint string) string {
	return rlk.kb.BuildWithPrefix(CachePrefixRateLimit, "user", fmt.Sprintf("%d", userID), endpoint)
}

func (rlk *RateLimitCacheKeys) GlobalRateLimit(endpoint string) string {
	return rlk.kb.BuildWithPrefix(CachePrefixRateLimit, "global", endpoint)
}

type AllCacheKeys struct {
	User         *UserCacheKeys
	Subscription *SubscriptionCacheKeys
	Payment      *PaymentCacheKeys
	Auth         *AuthCacheKeys
	Server       *ServerCacheKeys
	Coupon       *CouponCacheKeys
	Invoice      *InvoiceCacheKeys
	RateLimit    *RateLimitCacheKeys
	Builder      CacheKeyBuilder
}

func NewAllCacheKeys() *AllCacheKeys {
	kb := NewKeyBuilder()
	return &AllCacheKeys{
		User:         NewUserCacheKeys(kb),
		Subscription: NewSubscriptionCacheKeys(kb),
		Payment:      NewPaymentCacheKeys(kb),
		Auth:         NewAuthCacheKeys(kb),
		Server:       NewServerCacheKeys(kb),
		Coupon:       NewCouponCacheKeys(kb),
		Invoice:      NewInvoiceCacheKeys(kb),
		RateLimit:    NewRateLimitCacheKeys(kb),
		Builder:      kb,
	}
}
