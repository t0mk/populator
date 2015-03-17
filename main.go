package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

const localDirKey = "localDir"
const flagRequired = "REQUIRED"

var once sync.Once
var cacheCredentialsPtr *bool

var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func run(args []string) {
	darkGreen := color.New(color.FgGreen, color.Bold)
	darkGreen.Printf("About to run\n")
	color.Red("$ " + strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func build(name string, path string) {

	finfo, _ := os.Stat(path)
	if finfo == nil {
		fmt.Println("dir ", path, " does not exist => not building", name)
		return
	}
	if !finfo.IsDir() {
		fmt.Println("file ", path, " is not a directory => not building", name)
		return
	}
	run([]string{"docker", "build", "-t", name, path})

}

func expandTilde(ppath string) string {
	if ppath[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		return strings.Replace(ppath, "~", dir, 1)
	}
	return ppath
}

func get(repo map[string]string, repoUrl string) error {
	fmt.Println("getting ", repoUrl)
	if _, ok := repo[localDirKey]; !ok {
		fmt.Println("no LocalDir specified for repo ", repoUrl, " will clone it to random dir")
		sliced := strings.Split(strings.TrimSuffix(repoUrl, ".git"), "/")
		shortName := sliced[len(sliced)-1]
		repo[localDirKey] = path.Join(os.Getenv("HOME"), nextSuffix()+"_"+shortName)
		fmt.Println("random dir for repo ", repoUrl, " is ", repo[localDirKey])
	}
	gitcmd := "clone"
	repo[localDirKey] = expandTilde(repo[localDirKey])
	finfo, _ := os.Stat(repo[localDirKey])
	if finfo != nil {
		if finfo.IsDir() {
			gitcmd = "pull"
		} else {
			return errors.New(repo[localDirKey] + " exists and is not a directory")
		}
	}
	if *cacheCredentialsPtr {
		once.Do(func() { run([]string{"git", "config", "--global", "credentials.helper", "cache"}) })
	}
	if gitcmd == "pull" {
		fmt.Println("Pulling repo", repoUrl, "to dir", repo[localDirKey])
		os.Chdir(repo[localDirKey])
		run([]string{"git", gitcmd})
	}
	if gitcmd == "clone" {
		fmt.Println("Cloning repo", repoUrl, "to dir", repo[localDirKey])
		run([]string{"git", gitcmd, "--depth", "1", repoUrl, repo[localDirKey]})
	}
	return nil
}

func main() {
	cacheCredentialsPtr = flag.Bool("credcache", true, "should git cache credentials?")
	configFilePtr := flag.String("config", flagRequired, "configuration file containing repo urls and image names and locations")
	buildOnlyPtr := flag.Bool("onlybuild", false, "do not git pull/clone, only docker build")
	onlyPtr := flag.String("only", "", "only download repos matching this substring | only build images matching this substring")
	downloadOnlyPtr := flag.Bool("onlydownload", false, "only git pull/clone, do not docker build")
	flag.Parse()

	if *configFilePtr == flagRequired {
		log.Fatal("You must supply the config file in the -config flag")
	}
	data, err := ioutil.ReadFile(*configFilePtr)
	if err != nil {
		log.Fatal("Cant open file ", *configFilePtr, " to read config")
	}
	m := make(map[string]map[string]string)
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(m)
	if !(*buildOnlyPtr) {
		for repoUrl, repoMap := range m {
			if strings.Contains(repoUrl, *onlyPtr) {
				err = get(repoMap, repoUrl)
				if err != nil {
					log.Fatalf("error while getting %s: %v", repoUrl, err)
				}
				fmt.Println("\n")
			}
		}
	}
	if !(*downloadOnlyPtr) {
		for _, repoMap := range m {
			localRepoDir := repoMap[localDirKey]
			fmt.Println(localRepoDir)
			for repoPath, imageName := range repoMap {
				if repoPath != localDirKey && strings.Contains(imageName, *onlyPtr) {
					build(imageName, path.Join(localRepoDir, repoPath))
					fmt.Println("\n")
				}
			}
		}
	}
}
