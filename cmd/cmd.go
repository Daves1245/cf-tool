package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Eval evaluates command
func Eval(args map[string]interface{}) error {
	if args["config"].(bool) {
		return Config()
	}
	if args["submit"].(bool) {
		return Submit()
	}
	if args["list"].(bool) {
		return List()
	} else if args["parse"].(bool) {
		return Parse()
	} else if args["gen"].(bool) {
		return Gen()
	} else if args["test"].(bool) {
		return TestConnection()
	} else if args["watch"].(bool) {
		return Watch()
	} else if args["open"].(bool) {
		return Open()
	} else if args["stand"].(bool) {
		return Stand()
	} else if args["sid"].(bool) {
		return Sid()
	} else if args["race"].(bool) {
		return Race()
	} else if args["pull"].(bool) {
		return Pull()
	} else if args["clone"].(bool) {
		return Clone()
	} else if args["upgrade"].(bool) {
		return Upgrade()
	}
	return nil
}

// TestConnection tests if we can connect to Codeforces
func TestConnection() error {
	cln := client.Instance
	if cln == nil {
		return fmt.Errorf("Client is not initialized. Please run 'cf config' first")
	}
	
	// Check if we have login credentials
	if cln.HandleOrEmail == "" || cln.Password == "" {
		color.Yellow("No login credentials found. Please run 'cf config' first to set up your credentials.")
		return cln.ConfigLogin()
	}
	
	return cln.TestConnection()
}

func getSampleID() (samples []string) {
	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`in(\d+).txt`)
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("ans%v.txt", idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

// CodeList Name matches some template suffix, index are template array indexes
type CodeList struct {
	Name  string
	Index []int
}

func getCode(filename string, templates []config.CodeTemplate) (codes []CodeList, err error) {
	mp := make(map[string][]int)
	for i, temp := range templates {
		suffixMap := map[string]bool{}
		for _, suffix := range temp.Suffix {
			if _, ok := suffixMap[suffix]; !ok {
				suffixMap[suffix] = true
				sf := "." + suffix
				mp[sf] = append(mp[sf], i)
			}
		}
	}

	if filename != "" {
		ext := filepath.Ext(filename)
		if idx, ok := mp[ext]; ok {
			return []CodeList{CodeList{filename, idx}}, nil
		}
		return nil, fmt.Errorf("%v can not match any template. You could add a new template by `cf config`", filename)
	}

	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, path := range paths {
		name := path.Name()
		ext := filepath.Ext(name)
		if idx, ok := mp[ext]; ok {
			codes = append(codes, CodeList{name, idx})
		}
	}

	return codes, nil
}

func getOneCode(filename string, templates []config.CodeTemplate) (name string, index int, err error) {
	codes, err := getCode(filename, templates)
	if err != nil {
		return
	}
	if len(codes) < 1 {
		return "", 0, errors.New("Cannot find any code.\nMaybe you should add a new template by `cf config`")
	}
	if len(codes) > 1 {
		color.Cyan("There are multiple files can be selected.")
		for i, code := range codes {
			fmt.Printf("%3v: %v\n", i, code.Name)
		}
		i := util.ChooseIndex(len(codes))
		codes[0] = codes[i]
	}
	if len(codes[0].Index) > 1 {
		color.Cyan("There are multiple languages match the file.")
		for i, idx := range codes[0].Index {
			fmt.Printf("%3v: %v\n", i, client.Langs[templates[idx].Lang])
		}
		i := util.ChooseIndex(len(codes[0].Index))
		codes[0].Index[0] = codes[0].Index[i]
	}
	return codes[0].Name, codes[0].Index[0], nil
}

func loginAgain(cln *client.Client, err error) error {
	if err != nil && err.Error() == client.ErrorNotLogged {
		color.Red("Not logged. Try to login\n")
		err = cln.Login()
	}
	return err
}
