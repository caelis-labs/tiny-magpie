package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

// Mount the checked DMG privately and copy its signed bundle without changing
// any metadata inside it. Installation still uses upstream's staged swap.
func unpackDMG(ctx context.Context, dmg, out string) (app string, err error) {
	mount := filepath.Join(filepath.Dir(out), "mount")
	if err := os.MkdirAll(mount, 0o700); err != nil {
		return "", err
	}
	if b, err := proc.CommandContext(ctx, "hdiutil", "attach", "-readonly", "-nobrowse", "-noautoopen", "-mountpoint", mount, dmg).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount update: %w: %s", err, b)
	}
	defer func() {
		// Detaching must still run when the download context was cancelled.
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if b, e := proc.CommandContext(cleanup, "hdiutil", "detach", mount).CombinedOutput(); e != nil {
			err = fmt.Errorf("detach update: %w: %s", e, b)
		}
	}()
	source := filepath.Join(mount, "tiny-magpie.app")
	info, err := os.Lstat(source)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("update has no tiny-magpie.app directory")
	}
	app = filepath.Join(out, "tiny-magpie.app")
	if b, err := proc.CommandContext(ctx, "ditto", source, app).CombinedOutput(); err != nil {
		return "", fmt.Errorf("copy update: %w: %s", err, b)
	}
	return app, nil
}
