package coupon

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/coupon/usecases/implementations"
	"linke/internal/domains/coupon/usecases/interfaces"
)

// Module Coupon 领域模块
// 提供优惠券系统、折扣计算、使用限制管理等功能
var Module = fx.Module("coupon",
	// 注意：目前 coupon 领域还没有 repository 实现
	// 当添加了 repository 时，需要在这里提供
	
	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewCouponService,
			fx.As(new(interfaces.CouponService)),
		),
	),

	// 注意：目前 coupon 领域还没有 handler 实现
	// 当添加了 handler 时，需要在这里提供

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保优惠券相关表存在并且结构正确
		// 可以添加默认优惠券的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供优惠券服务接口
type ServiceProvider struct {
	CouponService interfaces.CouponService
}

// NewServiceProvider 创建优惠券服务提供者
func NewServiceProvider(couponService interfaces.CouponService) *ServiceProvider {
	return &ServiceProvider{
		CouponService: couponService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("coupon-service",
	fx.Provide(NewServiceProvider),
)