// Package main - the OTP file is used for having a OTP manager //@package main otp 文件用于拥有 otp 管理器
package main //@包主

import ( //@进口
	"context" //@语境
	"time" //@时间

	"github.com/google/uuid" //@github com 谷歌 uuid
)


type OTP struct { //@输入 otp 结构
	Key     string //@关键字符串
	Created time.Time //@创建时间
}

type Verifier interface { //@类型验证器接口
	VerifyOTP(otp string) bool //@验证 ot p otp string bool
}


type RetentionMap map[string]OTP //@type retention map 映射字符串 ot p

// NewRetentionMap will create a new retentionmap and start the retention given the set period //@new retention map 将创建一个新的 retentionmap 并在给定的期限内开始保留
func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap { //@func new retention map ctx context context 保留期 time duration 保留图
	rm := make(RetentionMap) //@rm 制作保留地图

	go rm.Retention(ctx, retentionPeriod) //@go rm 保留 ctx 保留期

	return rm //@返回 rm
}

// NewOTP creates and adds a new otp to the map //@new otp 创建新的 otp 并将其添加到地图
func (rm RetentionMap) NewOTP() OTP { //@func rm 保留映射 new ot p otp
	o := OTP{
		Key:     uuid.NewString(), //@键 uuid 新字符串
		Created: time.Now(), //@现在创建时间
	}

	rm[o.Key] = o //@rm o 键 o
	return o //@回车
}

// VerifyOTP will make sure a OTP exists //@验证 ot p 将确保 otp 存在
// and return true if so //@如果是，则返回 true
// It will also delete the key so it cant be reused //@它还会删除密钥，因此无法重复使用
func (rm RetentionMap) VerifyOTP(otp string) bool { //@func rm retention map verify ot p otp 字符串 bool
	// Verify OTP is existing //@验证 otp 是否存在
	if _, ok := rm[otp]; !ok { //@如果没问题 rm otp 没问题
		// otp does not exist //@otp不存在
		return false //@返回假
	}
	delete(rm, otp) //@删除rm otp
	return true //@返回真
}

// Retention will make sure old OTPs are removed //@保留将确保删除旧的 o tps
// Is Blocking, so run as a Goroutine //@正在阻塞，所以作为 goroutine 运行
func (rm RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) { //@func rm retention map retention ctx context context retention period 持续时间
	ticker := time.NewTicker(400 * time.Millisecond) //@股票行情时间新的股票行情时间毫秒
	for { //@为了
		select { //@选择
		case <-ticker.C: //@案例代码 c
			for _, otp := range rm { //@对于 otp 范围 rm
				// Add Retention to Created and check if it is expired //@将保留添加到创建并检查它是否已过期
				if otp.Created.Add(retentionPeriod).Before(time.Now()) { //@如果 otp 创建时间之前添加保留期
					delete(rm, otp.Key) //@删除 rm otp 密钥
				}
			}
		case <-ctx.Done(): //@案例 ctx 完成
			return //@返回

		}
	}
}
