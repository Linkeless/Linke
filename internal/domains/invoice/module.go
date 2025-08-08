package invoice

import (
	"os"
	"path/filepath"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/invoice/adapters/repositories"
	"linke/internal/domains/invoice/adapters/handlers"
	"linke/internal/domains/invoice/usecases/implementations"
	"linke/internal/domains/invoice/usecases/interfaces"
	"linke/internal/shared/logger"
)

// Module Invoice 领域模块
// 提供发票生成、管理、PDF 生成、通知发送等功能
var Module = fx.Module("invoice",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewInvoiceRepository,
			fx.As(new(interfaces.InvoiceRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewInvoiceService,
			fx.As(new(interfaces.InvoiceService)),
		),
		// 提供 PDF 生成服务
		NewPDFGeneratorServiceWithConfig,
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewInvoiceHandler,
		handlers.NewAdminInvoiceHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保发票相关表存在并且结构正确
		// 可以添加发票模板的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供发票服务接口
type ServiceProvider struct {
	InvoiceService interfaces.InvoiceService
}

// NewServiceProvider 创建发票服务提供者
func NewServiceProvider(invoiceService interfaces.InvoiceService) *ServiceProvider {
	return &ServiceProvider{
		InvoiceService: invoiceService,
	}
}

// NewPDFGeneratorServiceWithConfig 创建 PDF 生成服务（带配置）
func NewPDFGeneratorServiceWithConfig(logger logger.Logger) *implementations.PDFGeneratorService {
	// 获取输出目录，默认为 ./data/invoices
	outputDir := os.Getenv("INVOICE_PDF_OUTPUT_DIR")
	if outputDir == "" {
		wd, _ := os.Getwd()
		outputDir = filepath.Join(wd, "data", "invoices")
	}

	return implementations.NewPDFGeneratorService(outputDir, logger)
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("invoice-service",
	fx.Provide(NewServiceProvider),
)
