package storage

import (
	"encoding/csv"
	"os"
)

type CSVWriter struct {
	f *os.File
	w *csv.Writer
}

func NewCSVWriter(path string) (*CSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	_, _ = f.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(f)
	w.Comma = ';'

	if err := w.Write([]string{"Имя товара", "Цена", "Ссылка"}); err != nil {
		_ = f.Close()
		return nil, err
	}
	w.Flush()

	return &CSVWriter{f: f, w: w}, nil
}

func (c *CSVWriter) WriteRow(name, price, url string) error {
	if err := c.w.Write([]string{name, price, url}); err != nil {
		return err
	}
	c.w.Flush()
	return c.w.Error()
}

func (c *CSVWriter) Close() error {
	c.w.Flush()
	_ = c.w.Error()
	return c.f.Close()
}
