package main //@包主

import ( //@进口
	"context" //@语境
	"testing" //@测试
	"time" //@时间
)


func TestRetentionMap_VerifyOTP(t *testing.T) { //@func test retention map verify ot p t 测试 t
	ctx := context.Background() //@ctx上下文背景
	ctx, cancel := context.WithCancel(ctx) //@ctx 使用 cancel ctx 取消上下文


	rm := NewRetentionMap(ctx, 1*time.Second) //@rm new retention map ctx 时间秒

	otp := rm.NewOTP() //@otp rm 新的 ot p

	if ok := rm.VerifyOTP(otp.Key); !ok{ //@如果正常 rm 验证 ot p otp 密钥正常
		t.Error("failed to verify otp key that exists") //@t 错误无法验证存在的 otp 密钥
	}
	if ok := rm.VerifyOTP(otp.Key); ok{ //@如果正常 rm 验证 ot p otp 密钥正常
		t.Error("Reusing a OTP should not succeed") //@重复使用 otp 的错误不应该成功
	}


	cancel() //@取消
}

func TestOTP_Retention(t *testing.T) { //@功能测试 ot p 保留 t 测试 t

	// Create context with cancel to stop goroutine //@使用 cancel 创建上下文以停止 goroutine
	ctx := context.Background() //@ctx上下文背景
	ctx, cancel := context.WithCancel(ctx) //@ctx 使用 cancel ctx 取消上下文

	// Create RM and add a few OTP with a few Seconds in between //@创建 rm 并添加几个 otp，中间间隔几秒钟
	rm := NewRetentionMap(ctx, 1*time.Second) //@rm new retention map ctx 时间秒

	rm.NewOTP() //@rm 新 ot p
	rm.NewOTP() //@rm 新 ot p

	time.Sleep(2 * time.Second) //@时间睡眠时间秒

	otp := rm.NewOTP() //@otp rm 新的 ot p

	// Make sure that only 1 password is still left and it matches the latest //@确保只剩下密码并且它与最新的相匹配
	if len(rm) != 1 { //@如果 len rm
		t.Error("Failed to clean up") //@t错误清理失败
	}

	if rm[otp.Key] != otp { //@如果 rm otp 密钥 otp
		t.Error("The key should still be in place") //@t error 钥匙应该还在原位
	}
	cancel() //@取消
}
