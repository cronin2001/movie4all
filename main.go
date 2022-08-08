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

        _ "github.com/lib/pq"
)



func main(){

        for i, v := range urls{
                
                index := os.Getenv("INDEX")
                episode_index := strings.Split(os.Getenv("EPISODE_INDEX"), ",");
                urls := strings.Split(os.Getenv("URLS"), ";")

                go func(i, v){
                        cmd := exec.Command("chmod", "+x", "autodelogo.sh")
                        cmd.Run()

                        cmd = exec.Command("bash", "autodelogo.sh", v)
                        cmd.Run()

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
                        case <- time.After(time.Second*900):
                                continue
                }
        }
}
