package stage2

import (
	"reflect"
	"testing"
)

func TestYieldResumesAtOriginalPosition(t *testing.T) {
	s := New()
	var got []string

	s.Spawn(func(ctx *Context) {
		local := "局部变量仍然存在"
		got = append(got, "第一步")
		ctx.Yield()
		got = append(got, local)
		ctx.Yield()
		got = append(got, "第三步")
	})

	s.Tick()
	assertStrings(t, got, []string{"第一步"})
	s.Tick()
	assertStrings(t, got, []string{"第一步", "局部变量仍然存在"})
	s.Tick()
	assertStrings(t, got, []string{"第一步", "局部变量仍然存在", "第三步"})

	if s.Pending() != 0 {
		t.Fatalf("未完成任务数 = %d，期望 0", s.Pending())
	}
}

func TestTasksRunRoundRobinInCreationOrder(t *testing.T) {
	s := New()
	var got []string

	s.Spawn(func(ctx *Context) {
		got = append(got, "A1")
		ctx.Yield()
		got = append(got, "A2")
	})
	s.Spawn(func(ctx *Context) {
		got = append(got, "B1")
		ctx.Yield()
		got = append(got, "B2")
	})

	s.Tick()
	assertStrings(t, got, []string{"A1", "B1"})
	s.Tick()
	assertStrings(t, got, []string{"A1", "B1", "A2", "B2"})
}

func TestYieldPreservesNestedCallStack(t *testing.T) {
	s := New()
	var got string

	s.Spawn(func(ctx *Context) {
		got = nestedCall(ctx)
	})

	s.Tick()
	if got != "" {
		t.Fatalf("第一次 Tick 后结果 = %q，期望任务仍在暂停", got)
	}
	s.Tick()
	if got != "调用栈已恢复" {
		t.Fatalf("第二次 Tick 后结果 = %q，期望 %q", got, "调用栈已恢复")
	}
}

func TestTaskSpawnedDuringTickStartsOnNextTick(t *testing.T) {
	s := New()
	childRan := false

	s.Spawn(func(*Context) {
		s.Spawn(func(*Context) {
			childRan = true
		})
	})

	s.Tick()
	if childRan {
		t.Fatal("子任务在创建它的 Tick 中运行了")
	}
	s.Tick()
	if !childRan {
		t.Fatal("子任务没有在下一次 Tick 中运行")
	}
}

func nestedCall(ctx *Context) string {
	local := "调用栈已恢复"
	ctx.Yield()
	return local
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("执行顺序 = %v，期望 %v", got, want)
	}
}
