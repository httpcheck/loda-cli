package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/lodastack/loda-cli/setting"
	"github.com/oiooj/cli"
)

var CmdTree = cli.Command{
	Name:        "tree",
	Usage:       "列出指定节点下的资源",
	Description: "列出指定节点下的资源",
	Action:      runTree,
	BashComplete: func(c *cli.Context) {
		// This will complete if no args are passed
		if len(c.Args()) > 0 {
			return
		}
		for _, t := range MachineInit() {
			fmt.Println(t)
		}
	},
}

func runTree(c *cli.Context) {
	if len(c.Args()) > 0 {
		ns := c.Args()[0]
		var serverList ServerList
		for _, server := range serverList.think(ns) {
			fmt.Println(server.IP)
		}
	} else {
		var nsList NameSpaceList
		for _, ns := range nsList.AllNameSpaces() {
			fmt.Println(ns)
		}
	}
}

type ServerList struct {
	Members   []Server `json:"data"`
	NameSpace string
}

type Server struct {
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	LastReport string `json:"lastReport"`
	Status     string `json:"status"`
	Version    string `json:"version"`
}

func (this *ServerList) think(ns string) []Server {
	arr := strings.SplitN(ns, ".", 2)
	switch strings.ToLower(arr[0]) {
	case "machine":
		return this.getServerList(arr[1], arr[0])
	default:
		fmt.Println("Dont support this resource type. Try: mechine.xxx.loda")
	}
	return this.Members
}

func (this *ServerList) getServerList(ns, resType string) []Server {
	url := fmt.Sprintf(setting.API_Res, ns, resType)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Get from loda error: ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read from HTTP error: ", err)
		os.Exit(1)
	}
	json.Unmarshal(body, &this)
	if len(this.Members) == 0 {
		fmt.Println("No resource found, check your NS.")
	}
	for i, s := range this.Members {
		for _, ip := range strings.Split(s.IP, ",") {
			if IsIntranet(ip) {
				s.IP = ip
				this.Members[i] = s
				break
			}
		}
	}
	return this.Members
}

func IsIntranet(ipStr string) bool {
	if strings.HasPrefix(ipStr, "10.") {
		return true
	}

	if strings.HasPrefix(ipStr, "172.") {
		// 172.16.0.0-172.31.255.255
		arr := strings.Split(ipStr, ".")
		if len(arr) != 4 {
			return false
		}

		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}

		if second >= 16 && second <= 31 {
			return true
		}
	}

	return false
}