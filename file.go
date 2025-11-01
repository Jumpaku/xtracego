package goxe

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func TransformFile(src, dst string, transform func(r io.Reader, w io.Writer) (err error)) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to find %s: %w", src, err)
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dst, err)
	}

	writer, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dst, err)
	}
	defer writer.Close()

	if err := transform(reader, writer); err != nil {
		return err
	}

	if err := os.Chmod(dst, sourceFileStat.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", dst, err)
	}
	return nil
}
