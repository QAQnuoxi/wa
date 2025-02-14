// 版权 @2021 凹语言 作者。保留所有权利。

package loader

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	"wa-lang.org/wa/internal/config"
	"wa-lang.org/wa/internal/logger"
	"wa-lang.org/wa/internal/waroot"
)

// 根据路径加载需要的 vfs 和 manifest
func loadProgramFileMeta(cfg *config.Config, filename string, src interface{}) (
	vfs *config.PkgVFS,
	manifest *config.Manifest,
	err error,
) {
	logger.Tracef(&config.EnableTrace_loader, "cfg: %+v", cfg)
	logger.Tracef(&config.EnableTrace_loader, "filename: %s", filename)

	// 读取代码
	srcData, err := readSource(filename, src)
	if err != nil {
		return nil, nil, err
	}

	// 尝试加载本地的 manifest
	manifest, err = config.LoadManifest(nil, filepath.Dir(filename))
	if err != nil {
		err = nil // 忽略错误
	}

	// 重新构造 manifest
	if manifest == nil {
		manifest = &config.Manifest{
			Root:    "__main__",
			MainPkg: "__main__",
			Pkg: config.Manifest_package{
				Name:    filename,
				Pkgpath: "__main__",
			},
		}
	}

	logger.Tracef(&config.EnableTrace_loader, "manifest: %s", manifest.JSONString())

	// 构造入口文件
	vfs = new(config.PkgVFS)
	if vfs.App == nil {
		vfs.App = fstest.MapFS{
			filename: &fstest.MapFile{
				Data: srcData,
			},
		}
	}

	if vfs.Std == nil {
		if cfg.WaRoot != "" {
			vfs.Std = os.DirFS(filepath.Join(cfg.WaRoot, "src"))
		} else {
			vfs.Std = waroot.GetFS()
		}
	}
	if vfs.Vendor == nil {
		if src == nil {
			vfs.Vendor = os.DirFS(filepath.Join(manifest.Root, "vendor"))
		}
		if vfs.Vendor == nil {
			vfs.Vendor = make(fstest.MapFS) // empty fs
		}
	}

	return
}

// 根据路径加载需要的 vfs 和 manifest
func loadProgramMeta(cfg *config.Config, appPath string) (
	vfs *config.PkgVFS,
	manifest *config.Manifest,
	err error,
) {
	logger.Tracef(&config.EnableTrace_loader, "cfg: %+v", cfg)
	logger.Tracef(&config.EnableTrace_loader, "appPath: %s", appPath)

	manifest, err = config.LoadManifest(nil, appPath)
	if err != nil {
		logger.Tracef(&config.EnableTrace_loader, "err: %v", err)
		return nil, nil, err
	}

	logger.Tracef(&config.EnableTrace_loader, "manifest: %s", manifest.JSONString())

	vfs = new(config.PkgVFS)
	if vfs.App == nil {
		vfs.App = os.DirFS(filepath.Join(manifest.Root, "src"))
	}

	if vfs.Std == nil {
		if cfg.WaRoot != "" {
			vfs.Std = os.DirFS(filepath.Join(cfg.WaRoot, "src"))
		} else {
			vfs.Std = waroot.GetFS()
		}
	}
	if vfs.Vendor == nil {
		vfs.Vendor = os.DirFS(filepath.Join(manifest.Root, "vendor"))
		if vfs.Vendor == nil {
			vfs.Vendor = make(fstest.MapFS) // empty fs
		}
	}

	return
}

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			d, err := io.ReadAll(s)
			return d, err
		}
		return nil, errors.New("invalid source")
	}

	d, err := os.ReadFile(filename)
	return d, err
}

func isWaFile(path string) bool {
	if fi, err := os.Lstat(path); err == nil && fi.Mode().IsRegular() {
		return strings.HasSuffix(strings.ToLower(path), ".wa")
	}
	return false
}

func isWzFile(path string) bool {
	if fi, err := os.Lstat(path); err == nil && fi.Mode().IsRegular() {
		return strings.HasSuffix(strings.ToLower(path), ".wz")
	}
	return false
}
