package vindinium

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MoveTimeout  = 15
	StartTimeout = 10 * 60
)

type Client struct {
	Server    string
	Key       string
	Mode      string
	BotName   string
	Turns     string
	RandomMap bool
	Debug     bool
	Bot       Bot
	State     *State
	Url       string
}

func NewClient(server, key, mode, botName, turns string, randomMap bool, debug bool) (client *Client) {
	client = &Client{
		Server:    server,
		Key:       key,
		Mode:      mode,
		BotName:   botName,
		Turns:     turns,
		RandomMap: randomMap,
		Debug:     debug,
	}
	client.Setup()
	return
}

func (c *Client) Setup() {
	c.Url = c.Server + "/api/" + c.Mode
	switch c.BotName {
	case "cash":
		c.Bot = &CashBot{}
	case "fighter":
		c.Bot = &FighterBot{}
	default:
		c.Bot = &RandomBot{}
	}
}

func (c *Client) finished() bool {
	return c.State.Game.Finished
}

func (c *Client) move(dir Direction) error {
	values := make(url.Values)
	values.Set("dir", string(dir))
	return c.post(c.State.PlayUrl, values, MoveTimeout)
}

func (c *Client) post(uri string, values url.Values, seconds int) error {
	if c.Debug {
		fmt.Printf("Making request to to: %s\n", uri)
	}
	timeout := time.Duration(seconds) * time.Second
	dial := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{Dial: dial}
	client := http.Client{Transport: &transport}

	response, err := client.PostForm(uri, values)
	if err != nil {
		fmt.Println("PostForm error: ", values)
		return err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	if err := json.Unmarshal(data, &c.State); err != nil {
		fmt.Println("Unmarshal Error: ", string(data), "state: ", c.State)
		return err
	}

	if c.Debug {
		fmt.Printf("Setting data to:\n%s\n", string(data))
	}

	return nil
}

func (c *Client) Start() error {
	values := make(url.Values)
	values.Set("key", c.Key)
	if c.Mode == "training" {
		values.Set("turns", c.Turns)
		if !c.RandomMap {
			values.Set("map", "m1")
		}
	}

	fmt.Println("Connecting and waiting for other players to join...")
	return c.post(c.Url, values, StartTimeout)
}

func (c *Client) Play() error {
	fmt.Printf("Playing at: %s\n", c.State.ViewUrl)
	move := 1
	for c.State.Game.Finished == false {
		fmt.Printf("Making move: %d\n", move)

		if c.Debug {
			fmt.Printf("\nclient: %+v\n", c)
			fmt.Printf("bot: %+v\n", c.Bot)
			fmt.Printf("state: %+v\n", c.State)
		}

		size := c.State.Game.Board.Size * 2
		fmt.Println(strings.Repeat("=", size))
		for i := 0; i < size*(size/2); i = i + size {
			fmt.Println(c.State.Game.Board.Tiles[i : i+size])
		}
		c.State.Game.Board.parseTiles()
		fmt.Printf("%+v\n", c.State.Game.Board.Tileset)

		startPlaying := time.Now()
		dir := c.Bot.Move(c.State)
		fmt.Println("Time taken to play: ", time.Since(startPlaying))
		if err := c.move(dir); err != nil {
			return err
		}

		move++
	}

	fmt.Println("\nFinished game.", c.State.ViewUrl)
	return nil
}
