package main

import (
	"bytes"
	"github.com/jlaffaye/ftp"
	"io"
	"os"
	"strings"
	"time"
)

func LoadHTMLFromFTP() (string, error) {
	c, err := ftp.Dial(os.Getenv("FTP_HOST"), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return "", err
	}
	defer c.Quit()

	err = c.Login(os.Getenv("FTP_USER"), os.Getenv("FTP_PASSWORD"))
	if err != nil {
		return "", err
	}

	resp, err := c.Retr(os.Getenv("FTP_PATH"))
	if err != nil {
		return "", err
	}
	defer resp.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SaveHTMLToFTP(content string) error {
	c, err := ftp.Dial(os.Getenv("FTP_HOST"), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	defer c.Quit()

	err = c.Login(os.Getenv("FTP_USER"), os.Getenv("FTP_PASSWORD"))
	if err != nil {
		return err
	}

	return c.Stor(os.Getenv("FTP_PATH"), strings.NewReader(content))
}
