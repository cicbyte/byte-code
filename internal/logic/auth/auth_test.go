package auth

import "testing"

// loginGuard 是纯内存状态机，可直接单测（不触库）
func TestLoginGuardLock(t *testing.T) {
	u := "test-user-lock"
	defer loginResetFailures(u)

	// 未失败时不锁
	if err := loginLockCheck(u); err != nil {
		t.Fatalf("初始状态不应锁定: %v", err)
	}
	// 连续失败达到阈值后锁定
	for i := 0; i < loginMaxFailures; i++ {
		loginRecordFailure(u)
	}
	if err := loginLockCheck(u); err == nil {
		t.Fatal("达到阈值后应锁定")
	}
	// 锁定提示包含时长信息
	if err := loginLockCheck(u); err != nil && err.Error() == "" {
		t.Fatal("锁定提示不应为空")
	}
}

func TestLoginGuardReset(t *testing.T) {
	u := "test-user-reset"
	defer loginResetFailures(u)

	for i := 0; i < loginMaxFailures-1; i++ {
		loginRecordFailure(u)
	}
	loginResetFailures(u)
	if err := loginLockCheck(u); err != nil {
		t.Fatalf("成功登录后应清零计数: %v", err)
	}
	// 清零后再失败一次不应立即锁定
	loginRecordFailure(u)
	if err := loginLockCheck(u); err != nil {
		t.Fatalf("清零后单次失败不应锁定: %v", err)
	}
}
