package main

import (
        "os"
        "os/exec"
        "github.com/web3-storage/go-w3s-client"
        "fmt"
        "context"
        "database/sql"
        "strings"
        "log"
        "time"
        "io/ioutil"
        "gopkg.in/vansante/go-ffprobe.v2"
        "math/rand"

        _ "github.com/lib/pq"
)

var (
        index string
        episode_index []string
        urls []string
        proxies []string
)


func gettbn()string{
        ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	fileReader, err := os.Open("download.mp4")
	if err != nil {
		log.Panicf("Error opening test file: %v", err)
	}

	data, err := ffprobe.ProbeReader(ctx, fileReader)
	if err != nil {
		log.Panicf("Error getting data: %v", err)
	}

	tbn := strings.Split(data.Streams[0].TimeBase, "/")
	return tbn[1]
}


func main(){

        index = os.Getenv("INDEX")
        episode_index = strings.Split(os.Getenv("EPISODE_INDEX"), ",");
        urls = strings.Split(os.Getenv("URLS"), ";")
        proxies = strings.Split(os.Getenv("PROXIES"), ";")


        for i, v := range urls{

                go func(i int, v string){

                        rand.Seed(time.Now().UnixNano())
	                randIdx := rand.Intn(len(proxies))
	                proxy := proxies[randIdx]

                        cmd := exec.Command("wget", "--limit-rate", "3m", "-O", "download.mp4", `https://`+proxy+`/main/`+v)
                        cmd.Run()

                        cmd = exec.Command("chmod", "+x", "autodelogo.sh")
                        cmd.Run()

                        tbn := gettbn()

                        cmd = exec.Command("bash", "autodelogo.sh", tbn)
                        cmd.Run()


                        dirmain, _ := ioutil.ReadDir("main")
                        if len(dirmain) == 0{
                                return
                        }


                        c, _ := w3s.NewClient(w3s.WithToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkaWQ6ZXRocjoweDMyQWI1NThkQWVCN2Y5MjQ3NzY5ZTM3MGZkYTBGYTNFNmRlM2I2QWMiLCJpc3MiOiJ3ZWIzLXN0b3JhZ2UiLCJpYXQiOjE2NDMxMzIwNDY1MTgsIm5hbWUiOiJzdG9yYWdlIn0.-sEIB2KQ48wP0GeCx53hUKvEqPJ7wFw7Qf1yseY8kUs"))

                        dir, err := os.Open("main")
                        if err!= nil{
                                panic(err)
                        }
                        cid, err := c.Put(context.Background(), dir)
                        if err != nil{
                                panic(err)
                        }

                        cmd = exec.Command("rm", "-rf", "main")
                        cmd.Run()

                        db, err := sql.Open("postgres", `postgres://evaddaucvcbnxo:785c7b60fead46d306ace829c26b00d815ebf12d053f37fb00626fc01945e9e1@ec2-54-75-26-218.eu-west-1.compute.amazonaws.com:5432/d58pvsk1dskehn`)
                        if err != nil{
                                log.Fatal(err)
                        }

                        if _, err := db.Exec(`INSERT INTO detail_table(index, episode_index, episode_url) VALUES($1, $2, $3)`, index, episode_index[i], fmt.Sprintf("%v", cid)); err != nil{
                                log.Fatal(err)
                        }
                }(i, v)

                select{
                        case <- time.After(time.Second*800):
                                continue
                }
        }
}
