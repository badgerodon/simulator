package builder

import (
	"os/exec"

	"github.com/pkg/errors"
)

func brotliCompress(name string) error {
	bs, err := exec.Command("brotli", "-q", "9", "-f", name).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "error compressing `%s`:\n%s", name, bs)
	}
	return nil
}

func gzipCompress(name string) error {
	bs, err := exec.Command("gzip", "-9", "-k", "-f", name).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "error compressing `%s`:\n%s", name, bs)
	}
	return nil
}
