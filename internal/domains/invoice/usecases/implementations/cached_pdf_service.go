package implementations

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"linke/internal/domains/invoice/entities"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"

	"go.uber.org/zap"
)

// CachedPDFService wraps PDFGeneratorService with caching capabilities
type CachedPDFService struct {
	pdfGenerator *PDFGeneratorService
	cache        cache.Cache
	logger       logger.Logger
	cacheTTL     time.Duration
}

// NewCachedPDFService creates a new cached PDF service
func NewCachedPDFService(pdfGenerator *PDFGeneratorService, cache cache.Cache, logger logger.Logger) *CachedPDFService {
	return &CachedPDFService{
		pdfGenerator: pdfGenerator,
		cache:        cache,
		logger:       logger,
		cacheTTL:     24 * time.Hour, // Cache PDFs for 24 hours
	}
}

// GeneratePDFCached generates a PDF with caching
func (cps *CachedPDFService) GeneratePDFCached(ctx context.Context, invoice *entities.Invoice, options *PDFGenerationOptions) ([]byte, string, error) {
	// Generate cache key based on invoice and options
	cacheKey := cps.generateCacheKey(invoice, options)

	// Try to get from cache first
	if cachedPDF, err := cps.cache.Get(ctx, cacheKey); err == nil {
		cps.logger.Info("PDF served from cache",
			zap.String("invoice_number", invoice.InvoiceNumber),
			zap.String("cache_key", cacheKey))
		return cachedPDF, "", nil
	}

	// Generate PDF if not in cache
	pdfBytes, filePath, err := cps.pdfGenerator.GeneratePDF(ctx, invoice, options)
	if err != nil {
		return nil, "", err
	}

	// Cache the PDF (only if not saving to disk to avoid redundancy)
	if !options.SaveToDisk {
		if err := cps.cache.Set(ctx, cacheKey, pdfBytes, cps.cacheTTL); err != nil {
			cps.logger.Warn("Failed to cache PDF",
				zap.Error(err),
				zap.String("cache_key", cacheKey))
		} else {
			cps.logger.Info("PDF cached successfully",
				zap.String("invoice_number", invoice.InvoiceNumber),
				zap.String("cache_key", cacheKey),
				zap.Int("size_bytes", len(pdfBytes)))
		}
	}

	return pdfBytes, filePath, nil
}

// InvalidateInvoiceCache invalidates cache for a specific invoice
func (cps *CachedPDFService) InvalidateInvoiceCache(ctx context.Context, invoiceID uint) error {
	// Generate pattern to match all cache keys for this invoice
	pattern := fmt.Sprintf("pdf:invoice:%d:*", invoiceID)

	// Delete all matching keys
	if err := cps.cache.DeleteByPattern(ctx, pattern); err != nil {
		cps.logger.Error("Failed to invalidate invoice PDF cache",
			zap.Error(err),
			zap.Uint("invoice_id", invoiceID))
		return err
	}

	cps.logger.Info("Invoice PDF cache invalidated",
		zap.Uint("invoice_id", invoiceID))
	return nil
}

// WarmCache pre-generates and caches PDFs for frequently accessed invoices
func (cps *CachedPDFService) WarmCache(ctx context.Context, invoices []*entities.Invoice, templates []string) error {
	cps.logger.Info("Starting PDF cache warm-up",
		zap.Int("invoice_count", len(invoices)),
		zap.Strings("templates", templates))

	var warmed, failed int
	for _, invoice := range invoices {
		for _, template := range templates {
			options := &PDFGenerationOptions{
				Template:   template,
				Language:   invoice.Language,
				SaveToDisk: false,
			}

			if options.Language == "" {
				options.Language = "en"
			}

			_, _, err := cps.GeneratePDFCached(ctx, invoice, options)
			if err != nil {
				failed++
				cps.logger.Error("Failed to warm cache for invoice",
					zap.Error(err),
					zap.Uint("invoice_id", invoice.ID),
					zap.String("template", template))
			} else {
				warmed++
			}
		}
	}

	cps.logger.Info("PDF cache warm-up completed",
		zap.Int("warmed", warmed),
		zap.Int("failed", failed))

	return nil
}

// GetCacheStats returns cache statistics for PDF generation
func (cps *CachedPDFService) GetCacheStats(ctx context.Context) map[string]any {
	// This would depend on the cache implementation
	// For now, return basic info
	return map[string]any{
		"cache_ttl_hours": cps.cacheTTL.Hours(),
		"service_type":    "cached_pdf_service",
	}
}

// generateCacheKey creates a unique cache key for PDF generation
func (cps *CachedPDFService) generateCacheKey(invoice *entities.Invoice, options *PDFGenerationOptions) string {
	// Create a string that uniquely identifies this PDF configuration
	keyData := fmt.Sprintf("invoice:%d:updated:%d:template:%s:lang:%s:watermark:%s:qr:%t",
		invoice.ID,
		invoice.UpdatedAt.Unix(),
		options.Template,
		options.Language,
		options.Watermark,
		options.IncludeQR,
	)

	// Add company info if present
	if options.CompanyInfo != nil {
		keyData += fmt.Sprintf(":company:%s:%s", options.CompanyInfo.Name, options.CompanyInfo.TaxID)
	}

	// Add custom fields if present
	if len(options.CustomFields) > 0 {
		for k, v := range options.CustomFields {
			keyData += fmt.Sprintf(":%s:%s", k, v)
		}
	}

	// Generate MD5 hash for cache key
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("pdf:invoice:%d:%x", invoice.ID, hash)
}

// SetCacheTTL updates the cache TTL
func (cps *CachedPDFService) SetCacheTTL(ttl time.Duration) {
	cps.cacheTTL = ttl
	cps.logger.Info("PDF cache TTL updated", zap.Duration("ttl", ttl))
}

// ClearAllCache clears all PDF cache entries
func (cps *CachedPDFService) ClearAllCache(ctx context.Context) error {
	pattern := "pdf:invoice:*"
	if err := cps.cache.DeleteByPattern(ctx, pattern); err != nil {
		cps.logger.Error("Failed to clear all PDF cache", zap.Error(err))
		return err
	}

	cps.logger.Info("All PDF cache cleared")
	return nil
}
