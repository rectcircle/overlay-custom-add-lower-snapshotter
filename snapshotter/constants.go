//go:build linux

package snapshotter

const (
	DefaultRootDir           = "/var/lib/containerd/cn.rectcircle.containerd.overlay-custom-add-lower-snapshotter"
	SocksFileName            = "grpc.socks"
	LabelCustomAddLowerPaths = "cn.rectcircle.containerd/overlay-custom-add-lower.paths"
)
