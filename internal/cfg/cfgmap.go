package cfg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	L "team_server/internal/log"
	T "team_server/internal/types"
)

var _ T.ICfg = (*CfgMaps)(nil)

type CfgMaps struct {
	envFname  string
	envVals   map[string]string
	jsonFname string
	jsonUsers []T.User
	jsonAdmin T.User
	log       T.ILog
	mu        sync.Mutex
}

func NewCfgMaps(dir, file string) *CfgMaps {
	envVals := make(map[string]string, 16)
	envVals[T.TS_APP_NAME] = file
	envVals[T.TS_APP_IP] = "localhost"
	envVals[T.TS_LOG_LEVEL] = "INFO" // TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default if empty or mess)
	envVals[T.TG_BOT_PROXY] = "OFF"  // OFF, ON
	envVals[T.TG_BOT_TOKEN] = ""
	return &CfgMaps{
		envFname:  filepath.Join(dir, file+".env"),
		envVals:   envVals,
		jsonFname: filepath.Join(dir, file+".json"),
	}
}

func (c *CfgMaps) Parse() T.ICfg {
	c.log = L.NewLogFprintf(c, 0, 0)
	c.parseIpFromInterface()
	if _, err := os.Stat(c.envFname); err == nil {
		c.parseFileDotEnvVars()
	}
	c.parseOsEnvVars()
	if _, err := os.Stat(c.jsonFname); err == nil {
		c.parseFileJsonVars()
	}
	return c
}

func (c *CfgMaps) GetJsonUsers() []T.User {
	return c.jsonUsers
}
func (c *CfgMaps) GetJsonAdmin() T.User {
	return c.jsonAdmin
}

func (c *CfgMaps) parseFileJsonVars() {
	fileBuf, err := os.ReadFile(c.jsonFname)
	if err != nil {
		c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).GetJsonVals(): error while reading json cfg file", err))
	}
	var vals T.JsonVals
	err = json.Unmarshal(fileBuf, &vals)
	if err != nil {
		c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).GetJsonVals(): error while Unmarshaling json cfg file", err))
	}
	c.jsonUsers = vals.Users
	c.jsonAdmin = vals.Admin
}

func (c *CfgMaps) SetEnvVal(setkey string, setval string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.envVals {
		if key == setkey {
			c.envVals[setkey] = setval
			envFileBuf, err := os.ReadFile(c.envFname)
			if err != nil {
				c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).SetEnvVal(): error while reading cfg file", err))
			}

			regex := regexp.MustCompile(setkey + "=" + ".*\n")
			fixedEnvFileBuf := regex.ReplaceAllString(string(envFileBuf), setkey+"="+setval+"\n")

			err = os.WriteFile(c.envFname, []byte(fixedEnvFileBuf), os.ModePerm)
			if err != nil {
				c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).SetEnvVal(): error while writing cfg file", err))
			}
		}
	}
}

func (c *CfgMaps) GetEnvVal(key string) string {
	val := c.envVals[key]
	return val
}

func (c *CfgMaps) parseOsEnvVars() {
	for key := range c.envVals {
		if v, ok := os.LookupEnv(key); ok && (len(v) > 0) {
			c.envVals[key] = v
			c.log.LogDebug("OSENV %s=%s\n", key, v)
		}
	}
}

func (c *CfgMaps) parseFileDotEnvVars() {
	f, err := os.Open(c.envFname)
	if err != nil {
		c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).parseFileDotEnvVars(): error while opening cfg file", err))
		return
	}
	c.log.LogDebug("load config from file: %s", c.envFname)
	defer f.Close()

	pattern := regexp.MustCompile("^[0-9A-Za-z_]+=[0-9A-Za-z-_:/.]+")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := pattern.FindString(scanner.Text())
		if len(str) > 0 {
			strarr := strings.Split(str, "=")
			if _, ok := c.envVals[strarr[0]]; ok {
				c.envVals[strarr[0]] = strarr[1]
				c.log.LogDebug("CFGFILE %s=%s\n", strarr[0], strarr[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).parseFileDotEnvVars(): error while reading cfg file", err))
	}
}

func (c *CfgMaps) parseIpFromInterface() {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		c.log.LogError(fmt.Errorf("%s: %w", "(CfgMaps).parseIpFromInterface(): error while getting IP interface", err))
		return
	}
	strarr := []string{}
	for _, addr := range addr {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				strarr = append(strarr, ipnet.IP.String())
			}
		}
	}
	ip := strings.Join(strarr, ";")
	if len(ip) > 0 {
		c.SetEnvVal(T.TS_APP_IP, ip)
		c.log.LogDebug("(CfgMaps).parseIpFromInterface(): %s", ip)
	}
}
