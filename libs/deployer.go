package libs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jarvanstack/mysqldump"
)

var filePath string

func Start(c *Container) {
	templates := GetTemp()
	for _, mainFest := range c.ManiFests {
		filePath = filepath.Join(c.WorkDir, "."+mainFest)
		ApplyTemplate(filePath, templates[mainFest], c)
		args := "-f " + filePath + " up -d"
		cmd := exec.Command("docker-compose", strings.Split(args, " ")...)
		cmd.Env = os.Environ()
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	if Exist(filePath) {
		os.Remove(filePath)
	}
}

func Stop(c *Container) {
	templates := GetTemp()
	for _, mainFest := range c.ManiFests {
		filePath = filepath.Join(c.WorkDir, "."+mainFest)
		ApplyTemplate(filePath, templates[mainFest], c)
		args := "-f " + filePath + " down"
		cmd := exec.Command("docker-compose", strings.Split(args, " ")...)
		cmd.Env = os.Environ()
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	if Exist(filePath) {
		os.Remove(filePath)
	}
}

func MakeConfig(c *Container) error {
	configMap := map[string]string{GptConfig: "config.json", NgxConfig: "nginx/nginx.conf"}
	for config, file := range configMap {
		fileName := filepath.Join(c.WorkDir, c.DataDir, file)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(fileName), 0755)
			data := []byte(config)
			if err := os.WriteFile(fileName, data, 0644); err != nil {
				fmt.Println("Error config:", err)
			} else {
				fmt.Println("Config successfully!")
			}
		}
	}
	return nil
}

func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func Backup(c *Container) {
	currentTime := time.Now().Format("20060102-150405")
	args := "exec mongo mongodump -u " + c.DbUser + " -p " + c.DbPass + " -d fastgpt --authenticationDatabase=admin --gzip --out /backup/" + currentTime + "/"
	cmd := exec.Command("docker", strings.Split(args, " ")...)
	cmd.Env = os.Environ()
	//cmd.Stderr = os.Stderr
	cmd.Run()
	fmt.Printf("-> mongo backup finished: %s\n", c.WorkDir+"/backup/mongo/"+currentTime)

	argspg := "exec pg pg_dump -U " + c.DbUser + " postgres -f /backup/postgres-" + currentTime + ".sql"
	cmdpg := exec.Command("docker", strings.Split(argspg, " ")...)
	cmdpg.Env = append(os.Environ(), "PGPASSWORD="+c.DbPass)
	//cmdpg.Stderr = os.Stderr
	cmdpg.Run()
	fmt.Printf("-> postgres backup finished: %s\n", c.WorkDir+"/backup/pg/postgres-"+currentTime+".sql")

	mysqlFile, err := os.Create(c.WorkDir + "/backup/mysql/oneapi-" + currentTime + ".sql")
	if err != nil {
		fmt.Printf("** Error: %s\n", err)
	} else {
		mysqldump.Dump("root:"+c.DbPass+"@tcp(localhost:3306)/oneapi?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
			mysqldump.WithWriter(mysqlFile),
			mysqldump.WithDropTable(),
			mysqldump.WithData(),
		)
		fmt.Printf("-> mysql backup finished: %s\n", c.WorkDir+"/backup/mysql/oneapi-"+currentTime+".sql")
	}
}
