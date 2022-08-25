package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/web3-storage/go-w3s-client"
	"gopkg.in/vansante/go-ffprobe.v2"

	_ "github.com/lib/pq"
)

type response struct {
	URL      string `json:"url"`
	ID       string `json:"id"`
	FROM     string `json:"from"`
	URL_NEXT string `json:"url_next"`
	NID      int    `json:"nid"`
}

func findsubmatch(rege string, body string) []string {

	reg := regexp.MustCompile(rege)
	return reg.FindStringSubmatch(body)
}

var (
	res      response
	videourl string
	start    string
	latest   bool = false
	failed   []string
)

func gettbn() (string, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	fileReader, err := os.Open("tmp2.mp4")
	if err != nil {
		log.Printf("Error opening test file: %v", err)
	}

	data, err := ffprobe.ProbeReader(ctx, fileReader)
	if err != nil {
		log.Printf("Error getting data: %v", err)
	}

	if data == nil {
		return "", fmt.Errorf("empty response")
	}

	tbn := strings.Split(data.Streams[0].TimeBase, "/")
	return tbn[1], nil
}

func deferfunc() {

	cmd := exec.Command("rm", "-rf", "main")
	cmd.Run()
}

func handle(url string) {
	cmd := exec.Command("chmod", "+x", "autodelogo.sh")
	cmd.Run()
	cmd = exec.Command("chmod", "+x", "autoconvert.sh")
	cmd.Run()

	log.Printf("downloading: %s\n", url)
	cmd = exec.Command("wget", "--timeout=30", "-O", "download.mp4", url)
	cmd.Run()

	var position string

	if size := strings.Contains(url, "540P"); size {
		position = "delogo=x=64:y=2:w=560:h=54"
	} else if size := strings.Contains(url, "720P"); size {
		position = "delogo=x=64:y=2:w=650:h=80"
	} else {
		position = "delogo=x=380:y=20:w=1100:h=225"
	}

	cmd = exec.Command("bash", "autodelogo.sh", position)
	cmd.Run()

	tbn, err := gettbn()
	if err != nil {
		failed = append(failed, url)
		return
	}
	log.Printf("the current tbn is: %s", tbn)

	cmd = exec.Command("bash", "autoconvert.sh", tbn)
	cmd.Run()

	dirmain, _ := ioutil.ReadDir("main")
	if len(dirmain) == 0 {
		log.Println("the folder's empty")
		deferfunc()
		return
	}

	c, _ := w3s.NewClient(w3s.WithToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkaWQ6ZXRocjoweDMyQWI1NThkQWVCN2Y5MjQ3NzY5ZTM3MGZkYTBGYTNFNmRlM2I2QWMiLCJpc3MiOiJ3ZWIzLXN0b3JhZ2UiLCJpYXQiOjE2NDMxMzIwNDY1MTgsIm5hbWUiOiJzdG9yYWdlIn0.-sEIB2KQ48wP0GeCx53hUKvEqPJ7wFw7Qf1yseY8kUs"))

	dir, err := os.Open("main")
	if err != nil {
		deferfunc()
		return
	}
	cid, err := c.Put(context.Background(), dir)
	if err != nil {
		deferfunc()
		failed = append(failed, url)
		return
	}

	db, err := sql.Open("postgres", `postgres://evaddaucvcbnxo:785c7b60fead46d306ace829c26b00d815ebf12d053f37fb00626fc01945e9e1@ec2-54-75-26-218.eu-west-1.compute.amazonaws.com:5432/d58pvsk1dskehn`)
	if err != nil {
		deferfunc()
		failed = append(failed, url)
		return
	}
	defer db.Close()

	if _, err := db.Exec(`INSERT INTO detail_table(index, episode_index, episode_url) VALUES($1, $2, $3)`, res.ID, res.NID, fmt.Sprintf("%v", cid)); err != nil {
		deferfunc()
		failed = append(failed, url)
		return
	}
}

func dispatch(url string) {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)

	req.Host = "zxzj.vip"

	resp, err := client.Do(req)
	if err != nil {
		failed = append(failed, url)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		failed = append(failed, url)
		return
	}

	result := findsubmatch(`(?is)player_aaaa=(.*?)<`, string(body))

	if len(result) < 2 {
		failed = append(failed, url)
		return
	}

	if err := json.Unmarshal([]byte(result[1]), &res); err != nil {
		failed = append(failed, url)
		return
	} else {
		fmt.Printf("%+v\n", res)
	}

	if res.URL == "" {
		failed = append(failed, url)
		return
	}

	if res.FROM == "dpp" {
		req, _ = http.NewRequest("GET", "https://jx.zxzj.vip/dplayer.php?url="+res.URL, nil)

		req.Host = "jx.zxzj.vip"

		resp, err = client.Do(req)
		if err != nil {
			failed = append(failed, url)
			return
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			failed = append(failed, url)
			return
		}

		result = findsubmatch(`(?is)var urls = '(.*?)';`, string(body))
		if len(result) < 2 {
			failed = append(failed, url)
			return
		}

		videourl = result[1]
	} else if res.FROM == "ck" {
		req, _ = http.NewRequest("GET", "https://jx.zxzj.vip/ckplayer.php?url="+res.URL, nil)

		req.Host = "jx.zxzj.vip"

		resp, err = client.Do(req)
		if err != nil {
			failed = append(failed, url)
			return
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			failed = append(failed, url)
			return
		}

		result = findsubmatch(`(?is)var urls = '(.*?)';`, string(body))
		if len(result) < 2 {
			failed = append(failed, url)
			return
		}

		videourl = result[1]
	}

	//视频下载处理部分
	handle(videourl)

	//最后一集退出
	if res.URL_NEXT == "G" {
		latest = true
	}
}

func main() {

	start = os.Getenv("START")
	mode := os.Getenv("MODE")

	//判断是全部还是部分
	if mode == `all` {

		count, _ := strconv.Atoi(os.Getenv("COUNT"))

		for {
			if !latest {
				dispatch(start + fmt.Sprintf("%v", count) + `.html`)
				count++
			} else {
				break
			}
		}
	} else if mode == `partly` {

		lists := os.Getenv("LISTS")
		list := strings.Split(lists, ",")

		for _, v := range list {
			dispatch(start + fmt.Sprintf("%v", v) + `.html`)
		}
	} else {
		return
	}

	//再次尝试失败
	f := make([]string, len(failed))
	copy(f, failed)
	for _, v := range f {
		dispatch(v)
	}
}
