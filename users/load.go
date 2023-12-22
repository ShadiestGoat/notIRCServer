package users

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/shadiestgoat/log"
	"gopkg.in/yaml.v2"
)

type Perm int

const (
	PERM_READ Perm = 1 << iota
	PERM_READ_LIVE
	PERM_WRITE
	PERM_DELETE
)

const (
	str_perm_READ      = "read"
	str_perm_READ_LIVE = "read_live"
	str_perm_WRITE     = "write"
	str_perm_DELETE    = "delete"
)

var (
	permOrder = []Perm{
		PERM_READ,
		PERM_READ_LIVE,
		PERM_WRITE,
		PERM_DELETE,
	}
	strToInt = map[string]Perm{
		str_perm_READ:      PERM_READ,
		str_perm_READ_LIVE: PERM_READ_LIVE,
		str_perm_WRITE:     PERM_WRITE,
		str_perm_DELETE:    PERM_DELETE,
	}
	// will be created by init
	intToStr = map[Perm]string{}
	allPerms Perm

	defaultPerms Perm = PERM_READ | PERM_READ_LIVE | PERM_WRITE
)

func init() {
	for s, p := range strToInt {
		intToStr[p] = s
		allPerms |= p
	}
}

type User struct {
	Name                  string
	Color                 int
	Perms                 Perm
	AbleToReadAllWhispers bool
	WriteWhispers         map[string]bool
	ReadWhispers          map[string]bool
	Hidden                bool
	Token                 string
}

func (u User) HasPerm(p Perm) bool {
	return BitmaskHas(u.Perms, p)
}

func GetUser(name string) *User {
	return users[name]
}

type confUser struct {
	Color         string    `yaml:"color"`
	Perms         []string  `yaml:"perms"`
	WriteWhispers *[]string `yaml:"writeWhispers"`
	ReadWhispers  *[]string `yaml:"readWhispers"`
	Token         string    `yaml:"token"`
	Hidden        bool      `yaml:"hidden"`
}

// username -> *User
var users = map[string]*User{}

func Init(file string, disableTokenCheck bool) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0755)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("File '%v' does not exist!", file)
		}

		log.Fatal("Err with file '%v': %v", file, err)
	}

	d := yaml.NewDecoder(f)
	d.SetStrict(true)

	inpUsers := map[string]*confUser{}

	log.FatalIfErr(d.Decode(&inpUsers), "loading the users config file")

	for n, c := range inpUsers {
		if n == "" {
			log.Fatal("A 0-len user has been detected. This is not allowed.")
		}
		if n == "*" {
			log.Fatal("User '*' is special - you cannot name your user like this.")
		}

		if strings.HasPrefix(c.Token, "env:") {
			c.Token = os.Getenv(c.Token[4:])
		}

		if !disableTokenCheck {
			if c.Token == "" {
				log.Fatal("User '%v' needs to have a token!", n)
			} else if strings.Contains(c.Token, " ") {
				log.Fatal("Spaces are not allowed in tokens, but user '%v' uses them!", n)
			}
		}

		var col int64 = 0xffffff

		if c.Color != "" {
			if strings.HasPrefix(c.Color, "#") {
				c.Color = c.Color[1:]
			} else if strings.HasPrefix(c.Color, "0x") {
				c.Color = c.Color[2:]
			}

			if len(c.Color) == 3 {
				str := ""
				for i := 0; i < 3; i++ {
					str += string(c.Color[i]) + string(c.Color[i])
				}

				c.Color = str
			} else if len(c.Color) != 6 {
				log.Fatal("User '%v' has an invalid color (bad length: '%v')", n, c.Color)
			}

			col, err = strconv.ParseInt(c.Color, 16, 64)
			log.FatalIfErr(err, "parsing color for '%v'", n)
		}

		if c.ReadWhispers == nil {
			d := []string{n}
			c.ReadWhispers = &d
		}

		if c.WriteWhispers == nil {
			d := []string{"*"}
			c.WriteWhispers = &d
		}

		nc := User{
			Name:          n,
			Color:         int(col),
			Perms:         defaultPerms,
			WriteWhispers: map[string]bool{},
			ReadWhispers:  map[string]bool{},
			Token:         c.Token,
			Hidden:        c.Hidden,
		}

		for _, v := range c.Perms {
			if _, ok := strToInt[v]; !ok {
				log.Fatal("User '%v' has unrecognized perm '%v'", n, v)
			}

			nc.Perms ^= strToInt[v]
		}

		whisperConf := []struct {
			sl []string
			m  map[string]bool
			op string
		}{
			{
				sl: *c.ReadWhispers,
				m:  nc.ReadWhispers,
				op: "read",
			},
			{
				sl: *c.WriteWhispers,
				m:  nc.WriteWhispers,
				op: "write",
			},
		}

		for _, wc := range whisperConf {
			for _, v := range wc.sl {
				if _, ok := inpUsers[v]; !ok && v != "*" {
					log.Fatal("User '%v' tries to %v whispers from user '%v', who doesn't exist!", n, wc.op, v)
				}

				wc.m[v] = !wc.m[v]
			}

			for v := range inpUsers {
				wc.m[v] = wc.m[v]
			}

			if wc.m["*"] {
				for v := range wc.m {
					wc.m[v] = !wc.m[v]
				}
			}

			delete(wc.m, "*")
		}

		nc.AbleToReadAllWhispers = true

		if nc.AbleToReadAllWhispers {
			for _, v := range nc.ReadWhispers {
				if !v {
					nc.AbleToReadAllWhispers = false
					break
				}
			}
		}

		users[n] = &nc
	}

	if len(users) == 0 {
		log.Fatal("No users loaded")
	}

	log.Success("Loaded %v users:", len(users))
	for _, u := range users {
		perms := []string{}

		for _, p := range permOrder {
			if u.HasPerm(p) {
				perms = append(perms, intToStr[p])
			}
		}

		log.Success("\t%v#%X: (%v)", u.Name, u.Color, strings.Join(perms, ", "))
	}
}

func All() map[string]*User {
	return users
}
