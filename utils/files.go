package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func DownloadFileTo(url string, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}

	if resp.StatusCode != http.StatusOK {
		out.Close()
		resp.Body.Close()
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		resp.Body.Close()
		return err
	}

	err = out.Close()
	if err != nil {
		resp.Body.Close()
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
