//go:build linux

package snapshotter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/overlay"
)

func NewSnapshotter(root string, opts ...overlay.Opt) (snapshots.Snapshotter, error) {
	sn, err := overlay.NewSnapshotter(root, opts...)
	if err != nil {
		return nil, err
	}
	return &overlayCustomAddLowerSnapshotter{sn}, nil
}

// overlayCustomAddLowerSnapshotter 继承 overlay Snapshotter，在返回 mounts 的地方进行改造
type overlayCustomAddLowerSnapshotter struct {
	snapshots.Snapshotter
}

// Mounts implements snapshots.Snapshotter.
func (s *overlayCustomAddLowerSnapshotter) Mounts(ctx context.Context, key string) ([]mount.Mount, error) {
	mounts, err := s.Snapshotter.Mounts(ctx, key)
	if err != nil {
		return nil, err
	}
	return s.tryAddLowers(ctx, key, mounts)
}

// Prepare implements snapshots.Snapshotter.
func (s *overlayCustomAddLowerSnapshotter) Prepare(ctx context.Context, key string, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	mounts, err := s.Snapshotter.Prepare(ctx, key, parent, opts...)
	if err != nil {
		return nil, err
	}
	return s.tryAddLowers(ctx, key, mounts)
}

// View implements snapshots.Snapshotter.
func (s *overlayCustomAddLowerSnapshotter) View(ctx context.Context, key string, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	mounts, err := s.Snapshotter.View(ctx, key, parent, opts...)
	if err != nil {
		return nil, err
	}
	return s.tryAddLowers(ctx, key, mounts)
}

// tryAddLowers 所有返回 mounts 的地方，都需要调用该函数，根据 label ，给 lower 选项添加自定义的 lower 路径。
func (s *overlayCustomAddLowerSnapshotter) tryAddLowers(ctx context.Context, key string, mounts []mount.Mount) ([]mount.Mount, error) {
	if len(mounts) != 1 || mounts[0].Type != "overlay" {
		return mounts, nil
	}
	info, err := s.Snapshotter.Stat(ctx, key)
	if err != nil {
		return nil, err
	}
	lowerPathString, ok := info.Labels[LabelCustomAddLowerPaths]
	if !ok || lowerPathString == "" {
		return mounts, nil
	}
	lowerPaths := strings.Split(lowerPathString, ":")
	for _, p := range lowerPaths {
		if p == "" {
			continue
		}
		err = os.MkdirAll(p, 0o755)
		if err != nil {
			return nil, fmt.Errorf("mkdir lower path %s error: %s", p, err)
		}
	}
	for i, o := range mounts[0].Options {
		if strings.HasPrefix(o, "lowerdir=") {
			mounts[0].Options[i] = lowerPathString + ":" + o
			break
		}
	}
	return mounts, nil
}
