package referral

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/referral/adapters/handlers"
	"linke/internal/domains/referral/adapters/repositories"
	"linke/internal/domains/referral/usecases/implementations"
	"linke/internal/domains/referral/usecases/interfaces"
)

// Module Referral 领域模块
// 提供推荐系统、邀请码管理、奖励计算等功能
var Module = fx.Module("referral",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewReferralRepository,
			fx.As(new(interfaces.ReferralRepository)),
		),
		fx.Annotate(
			repositories.NewReferralCampaignRepository,
			fx.As(new(interfaces.ReferralCampaignRepository)),
		),
		fx.Annotate(
			repositories.NewInviteCodeRepository,
			fx.As(new(interfaces.InviteCodeRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewReferralService,
			fx.As(new(interfaces.ReferralService)),
		),
		fx.Annotate(
			implementations.NewReferralCampaignService,
			fx.As(new(interfaces.ReferralCampaignService)),
		),
		fx.Annotate(
			implementations.NewInviteCodeService,
			fx.As(new(interfaces.InviteCodeService)),
		),
		// 注意：InviteCodeUsageService 接口可能还未定义
		// fx.Annotate(
		//	implementations.NewInviteCodeUsageService,
		//	fx.As(new(interfaces.InviteCodeUsageService)),
		// ),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewAdminReferralHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保推荐相关表存在并且结构正确
		// 可以添加默认推荐活动的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供推荐服务接口
type ServiceProvider struct {
	ReferralService         interfaces.ReferralService
	ReferralCampaignService interfaces.ReferralCampaignService
	InviteCodeService       interfaces.InviteCodeService
}

// NewServiceProvider 创建推荐服务提供者
func NewServiceProvider(
	referralService interfaces.ReferralService,
	referralCampaignService interfaces.ReferralCampaignService,
	inviteCodeService interfaces.InviteCodeService,
) *ServiceProvider {
	return &ServiceProvider{
		ReferralService:         referralService,
		ReferralCampaignService: referralCampaignService,
		InviteCodeService:       inviteCodeService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("referral-service",
	fx.Provide(NewServiceProvider),
)
