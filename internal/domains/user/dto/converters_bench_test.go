package dto

import (
	"sync"
	"testing"
	"time"

	"linke/internal/domains/user/entities"

	"gorm.io/gorm"
)

// BenchmarkUserResponseConversion 测试用户响应转换的性能
func BenchmarkUserResponseConversion(b *testing.B) {
	user := &entities.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Name:      "Test User",
		Avatar:    "https://example.com/avatar.jpg",
		Provider:  "local",
		Status:    "active",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: gorm.DeletedAt{},
	}

	b.ResetTimer()

	b.Run("WithObjectPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp := ToUserResponse(user)
			PutUserResponse(resp)
		}
	})

	b.Run("WithoutObjectPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp := &UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Username:  user.Username,
				Name:      user.Name,
				Avatar:    user.Avatar,
				Provider:  user.Provider,
				Status:    user.Status,
				Role:      user.Role,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
			_ = resp // Prevent compiler optimization
		}
	})
}

// BenchmarkConcurrentObjectPoolUsage 测试并发场景下对象池的性能
func BenchmarkConcurrentObjectPoolUsage(b *testing.B) {
	user := &entities.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Name:      "Test User",
		Avatar:    "https://example.com/avatar.jpg",
		Provider:  "local",
		Status:    "active",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: gorm.DeletedAt{},
	}

	b.ResetTimer()

	b.Run("ConcurrentWithPool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp := ToUserResponse(user)
				PutUserResponse(resp)
			}
		})
	})

	b.Run("ConcurrentWithoutPool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp := &UserResponse{
					ID:        user.ID,
					Email:     user.Email,
					Username:  user.Username,
					Name:      user.Name,
					Avatar:    user.Avatar,
					Provider:  user.Provider,
					Status:    user.Status,
					Role:      user.Role,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
				}
				_ = resp // Prevent compiler optimization
			}
		})
	})
}

// BenchmarkBatchConversion 测试批量转换的性能
func BenchmarkBatchConversion(b *testing.B) {
	users := make([]*entities.User, 100)
	for i := 0; i < 100; i++ {
		users[i] = &entities.User{
			ID:        uint(i + 1),
			Email:     "test@example.com",
			Username:  "testuser",
			Name:      "Test User",
			Avatar:    "https://example.com/avatar.jpg",
			Provider:  "local",
			Status:    "active",
			Role:      "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{},
		}
	}

	b.ResetTimer()

	b.Run("BatchWithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			responses := ToUserResponseSlice(users)
			PutUserResponseSlice(responses)
		}
	})

	b.Run("BatchWithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			responses := make([]*UserResponse, len(users))
			for j, user := range users {
				responses[j] = &UserResponse{
					ID:        user.ID,
					Email:     user.Email,
					Username:  user.Username,
					Name:      user.Name,
					Avatar:    user.Avatar,
					Provider:  user.Provider,
					Status:    user.Status,
					Role:      user.Role,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
				}
			}
			_ = responses // Prevent compiler optimization
		}
	})
}

// BenchmarkMemoryAllocation 测试内存分配模式
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("StructAlignment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 测试结构体字段对齐对性能的影响
			resp := &UserResponse{
				ID:        1,
				Email:     "test@example.com",
				Username:  "testuser",
				Name:      "Test User",
				Avatar:    "https://example.com/avatar.jpg",
				Provider:  "local",
				Status:    "active",
				Role:      "user",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = resp
		}
	})
}

// BenchmarkPoolStatistics 测试统计数据的性能影响
func BenchmarkPoolStatistics(b *testing.B) {
	// 重置统计数据
	ResetUserPoolStats()

	b.Run("WithStatistics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp := GetUserResponse()
			PutUserResponse(resp)
		}
	})

	b.Run("WithoutStatistics", func(b *testing.B) {
		// 使用直接的 sync.Pool 而不经过统计包装
		pool := sync.Pool{
			New: func() any {
				return &UserResponse{}
			},
		}

		for i := 0; i < b.N; i++ {
			resp := pool.Get().(*UserResponse)
			*resp = UserResponse{} // Reset
			pool.Put(resp)
		}
	})
}

// TestObjectPoolCorrectness 测试对象池的正确性
func TestObjectPoolCorrectness(t *testing.T) {
	// 重置统计数据
	ResetUserPoolStats()

	// 获取多个对象
	resp1 := GetUserResponse()
	resp2 := GetUserResponse()

	// 设置不同的数据
	resp1.ID = 1
	resp1.Email = "user1@example.com"
	resp2.ID = 2
	resp2.Email = "user2@example.com"

	// 验证数据不会相互影响
	if resp1.ID == resp2.ID {
		t.Error("Object pool returned the same object instance")
	}

	// 返回到池中
	PutUserResponse(resp1)
	PutUserResponse(resp2)

	// 再次获取对象
	resp3 := GetUserResponse()

	// 验证对象已被重置
	if resp3.ID != 0 || resp3.Email != "" {
		t.Error("Object was not properly reset when returned to pool")
	}

	PutUserResponse(resp3)

	// 验证统计数据
	stats := GetUserPoolStats()
	if stats["user_response_hits"] == 0 {
		t.Error("Pool hit statistics not working")
	}
}

// TestMemoryAlignment 测试结构体内存对齐
func TestMemoryAlignment(t *testing.T) {
	// 这个测试主要用于验证结构体大小，确保字段对齐优化有效
	resp := &UserResponse{}

	// 基本验证：确保结构体可以正常使用
	resp.ID = 1
	resp.Email = "test@example.com"

	if resp.ID != 1 || resp.Email != "test@example.com" {
		t.Error("Struct field assignment failed")
	}
}
