package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// only for single request
type CacheRoundTrip struct {
	RoundTripper http.RoundTripper
	Etag         string
	Cache        []byte
}

type ReleaseData struct {
	TagName string `json:"tag_name"`
}

var (
	githubAPI   = "https://api.github.com/repos/"
	ghproxy     = "https://ghproxy.com/"
	retry       int
	interval    int
	extractPath string
	dlFilename  string
	useProxy    bool
	repo        string
)

func initFlag() {
	flag.StringVar(&repo, "repo", "", "github repository name, e.g. lyineee/hypothesis-web")
	flag.StringVar(&dlFilename, "filename", "dist.tar.gz", "download release assert name")
	flag.StringVar(&extractPath, "extract", "/usr/share/nginx/html", "extract path")
	flag.IntVar(&interval, "interval", 2, "get release info interval")
	flag.IntVar(&retry, "retry", 5, "download retry")
	flag.BoolVar(&useProxy, "ghproxy", true, "use ghproxy.com")
	flag.Parse()
	if repo == "" {
		fmt.Println("privide an repo name, e.g. lyineee/hypothesis-web")
		os.Exit(1)
	}

	log.Printf(`msg="show all config" repo=%s filename=%s extract=%s interval=%d retry=%d ghproxy=%t`, repo, dlFilename, extractPath, interval, retry, useProxy)
}

func main() {
	initFlag()
	if !useProxy {
		ghproxy = ""
	}
	client := http.Client{
		Transport: &CacheRoundTrip{
			RoundTripper: http.DefaultTransport,
		},
	}
	latest := ""
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/releases/latest", githubAPI, repo), nil)
		if err != nil {
			log.Println("new request error: ", err)
			continue
		}
		req.Header.Add("Accept", "application/vnd.github")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == 403 {
			log.Println("API rate limit exceeded", " ratelimit reset: ", resp.Header.Get("x-ratelimit-reset"))
			if resetTime := resp.Header.Get("x-ratelimit-reset"); resetTime != "" {
				unixTime, err := strconv.Atoi(resetTime)
				if err != nil {
					log.Println("err parse reset time to int, reset time string: ", resetTime)
				} else {
					resetTime := time.Unix(int64(unixTime), 0)
					sleepTime := resetTime.Sub(time.Now())
					log.Printf("sleep until reset time %s, sleep time: %s", resetTime, sleepTime)
					time.Sleep(sleepTime)
				}

			}
			continue
		} else if resp.StatusCode != 200 {
			log.Println(fmt.Sprintf(`msg="return none ok" status="%s"`, resp.Status))
			continue
		}

		release := ReleaseData{}
		err = json.NewDecoder(resp.Body).Decode(&release)
		if err != nil {
			log.Println("json recode error: ", err)
		}
		if release.TagName == latest {
			continue
		}
		latest = release.TagName
		log.Println("release: ", release.TagName)
		for re := retry; re > 0; re-- {
			dlUrl := fmt.Sprintf("%shttps://github.com/%s/releases/download/%s/%s", ghproxy, repo, release.TagName, dlFilename)
			log.Printf("download from: %s", dlUrl)
			err = downloadAndExtract(extractPath, dlUrl)
			if err != nil {
				log.Println("download and extract file err: ", err)
				continue
			}
			break
		}
	}
}

func (c *CacheRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	req := cloneRequest(r)
	req.Header.Add("If-None-Match", c.Etag)
	req.Method = "HEAD" //TODO remove form and request body?
	resp, err := c.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotModified {
		buf := bytes.NewBuffer(c.Cache)
		return http.ReadResponse(bufio.NewReader(buf), r)
	}
	resp, err = c.RoundTripper.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	c.Etag = resp.Header.Get("Etag")              // cache etag
	buf, err := httputil.DumpResponse(resp, true) // cache body
	if err != nil {
		return nil, err
	}
	c.Cache = buf
	return resp, nil
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
// (This function copyright goauth2 authors: https://code.google.com/p/goauth2)
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}

func downloadAndExtract(dst, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	for {
		hdr, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		// 处理下保存路径，将要保存的目录加上 header 中的 Name
		// 这个变量保存的有可能是目录，有可能是文件，所以就叫 FileDir 了……
		baseLength := len(filepath.Ext(filepath.Ext(dlFilename)))
		dstFileDir := filepath.Join(dst, hdr.Name[baseLength+1:])
		// 根据 header 的 Typeflag 字段，判断文件的类型
		switch hdr.Typeflag {
		case tar.TypeDir: // 如果是目录时候，创建目录
			// 判断下目录是否存在，不存在就创建
			if b := ExistDir(dstFileDir); !b {
				// 使用 MkdirAll 不使用 Mkdir ，就类似 Linux 终端下的 mkdir -p，
				// 可以递归创建每一级目录
				if err := os.MkdirAll(dstFileDir, 0775); err != nil {
					return err
				}
				log.Printf("make directory: %s", dstFileDir)
			}
		case tar.TypeReg: // 如果是文件就写入到磁盘
			// 创建一个可以读写的文件，权限就使用 header 中记录的权限
			// 因为操作系统的 FileMode 是 int32 类型的，hdr 中的是 int64，所以转换下
			file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			n, err := io.Copy(file, tarReader)
			if err != nil {
				return err
			}
			// 将解压结果输出显示
			log.Printf("extract： %s , total %d char\n", dstFileDir, n)

			// 不要忘记关闭打开的文件，因为它是在 for 循环中，不能使用 defer
			// 如果想使用 defer 就放在一个单独的函数中
			file.Close()
		}
	}
}

// 判断目录是否存在
func ExistDir(dirname string) bool {
	fi, err := os.Stat(dirname)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}
