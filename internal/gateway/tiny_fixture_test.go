package gateway

import (
	"context"
	"os"
	"testing"

	"github.com/yetone/magpie/internal/budget"
	"github.com/yetone/magpie/internal/gwbudget"
	"github.com/yetone/magpie/internal/gwmw"
	"github.com/yetone/magpie/internal/middleware"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/routes"
	"github.com/yetone/magpie/internal/testenv"
)

// Only retained gateway services run in these isolated boundary fixtures.
func TestMain(m *testing.M) {
	for name, strategy := range routes.Builtin {
		routes.Register(name, strategy)
	}
	provider.ServeGroups()
	gwbudget.Limit(budget.Limiter{})
	gwmw.Use(middleware.Chain{})
	claudeTUIRuns = false
	keepRoutes = false
	codexSocketOff.Store(true)
	provider.CodexZstdForTests(false)
	awaitAllowance = func(context.Context, string, string) bool { return false }
	os.Exit(testenv.RunIn(m, func(string) { testenv.Offline() }))
}
