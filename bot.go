package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/bwmarrin/discordgo"
	"github.com/otium/ytdl"
)

var (
	Token string
	APIKey1 string
	APIKey2 string
	APIKey3 string
	APIKey4 string
	APIKey5 string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&APIKey1, "k1", "", "API Key 1 - Rebrandly")
	flag.StringVar(&APIKey2, "k2", "", "API Key 2 - Rebrandly")
	flag.StringVar(&APIKey3, "k3", "", "API Key 3 - OMDB")
	flag.StringVar(&APIKey4, "k4", "", "API Key 4 - Watson Language Translator (Username)")
	flag.StringVar(&APIKey5, "k5", "", "API Key 5 - Watson Language Translator (Password)")
	flag.Parse()
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn((max-min) + 1) + min
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	args := strings.Split(m.Content, " ")

	if args[0] == "!p" {

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!p <data>`")
		} else {
			url := "https://hastebin.com/documents"
			t := strings.TrimPrefix(m.Content, "!p ")
			if strings.HasPrefix(t, "```") && strings.HasSuffix(t, "```") {
				t = strings.TrimPrefix(t, "```")
				t = strings.TrimPrefix(t, "\n")
				t = strings.TrimSuffix(t, "```")
				if strings.HasSuffix(t, "\n") {
					t = strings.TrimSuffix(t, "\n")
				}
			}
			d := strings.NewReader(t)
			rsp, err := http.Post(url, "application/json", d)
			if err != nil {
				log.Println(err)
				return
			}
			defer rsp.Body.Close()
			body, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				log.Println(err)
				return
			}
			v, err := jason.NewObjectFromBytes([]byte(body))
			if err != nil {
				log.Println(err)
				return
			}

			key, err := v.GetString("key")
			if err != nil {
				log.Println(err)
				return
			}

			paste := "<https://hastebin.com/" + key + ">"

			s.ChannelMessageSend(m.ChannelID, paste)
		}
	} else if args[0] == "!sh" {

		if len(args) >= 2 {
			if strings.HasPrefix(args[1], "<") && strings.HasSuffix(args[1], ">") {
				args[1] = strings.TrimPrefix(args[1], "<")
				args[1] = strings.TrimSuffix(args[1], ">")
			}
		}

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!sh https://www.google.com/`")
		} else if strings.HasPrefix(args[1], "http://") || strings.HasPrefix(args[1], "https://") {
			type Payload struct {
					Destination string `json:"destination"`
					Domain	struct {
						FullName string `json:"fullName"`
					} `json:"domain"`
				}

				data := Payload{}
				data.Destination = args[1]
				data.Domain.FullName = "rebrand.ly"

				payloadBytes, err := json.Marshal(data)
				if err != nil {
					log.Println(err)
					return
				}

				body := bytes.NewReader(payloadBytes)

				req, err := http.NewRequest("POST", "https://api.rebrandly.com/v1/links", body)
				if err != nil {
					log.Println(err)
					return
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Apikey", APIKey1)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println(err)
					return
				}

				defer resp.Body.Close()

				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}

				v, err := jason.NewObjectFromBytes([]byte(b))
				if err != nil {
					log.Println(err)
					return
				}

				shurl, err := v.GetString("shortUrl")
				if err != nil {
					log.Println(err)
					return
				}

				d := "<https://" + shurl + ">"

				s.ChannelMessageSend(m.ChannelID, d)
		}
	} else if args[0] == "!yt" {

		if len(args) >= 2 {
			if strings.HasPrefix(args[1], "<") && strings.HasSuffix(args[1], ">") {
					args[1] = strings.TrimPrefix(args[1], "<")
					args[1] = strings.TrimSuffix(args[1], ">")
			}
		}

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!yt https://www.youtube.com/watch?v=f0bbDFRYD_A`")
		} else if strings.HasPrefix(args[1], "http://") || strings.HasPrefix(args[1], "https://") {
			vid, err := ytdl.GetVideoInfo(args[1])
			if err != nil {
				log.Println(err)
				return
			}

			format := vid.Formats.Best(ytdl.FormatAudioBitrateKey)[0]

			url, err := vid.GetDownloadURL(format)
			if err != nil {
				log.Println(err)
				return
			}

			type Payload struct {
				Destination string `json:"destination"`
				Domain	struct {
					FullName string `json:"fullName"`
				} `json:"domain"`
			}

			data := Payload{}
			data.Destination = url.String()
			data.Domain.FullName = "rebrand.ly"

			payloadBytes, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return
			}

			body := bytes.NewReader(payloadBytes)

			req, err := http.NewRequest("POST", "https://api.rebrandly.com/v1/links", body)
			if err != nil {
				log.Println(err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Apikey", APIKey2)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				return
			}

			defer resp.Body.Close()

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}

			v, err := jason.NewObjectFromBytes([]byte(b))
			if err != nil {
				log.Println(err)
				return
			}

			shurl, err := v.GetString("shortUrl")
			if err != nil {
				log.Println(err)
				return
			}

			d := "<https://" + shurl + ">"

			s.ChannelMessageSend(m.ChannelID, d)
		}
	} else if args[0] == "!imdb" {

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!imdb Breaking Bad`")
			return
		}

		d := strings.TrimPrefix(m.Content, "!imdb ")
		x := &url.URL{Path: d}
		resp, err := http.Get("https://www.omdbapi.com/?apikey=" + APIKey3 + "&t=" + x.String())

		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		v, err := jason.NewObjectFromBytes([]byte(body))
		if err != nil {
			log.Println(err)
			return
		}

		title, err := v.GetString("Title")
		if err != nil {
			log.Println(err)
			return
		}

		imdbRating, err := v.GetString("imdbRating")
		if err != nil {
			log.Println(err)
			return
		}

		genre, err := v.GetString("Genre")
		if err != nil {
			log.Println(err)
			return
		}

		plot, err := v.GetString("Plot")
		if err != nil {
			log.Println(err)
			return
		}

		poster, err := v.GetString("Poster")
		if err != nil {
			log.Println(err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Author:	&discordgo.MessageEmbedAuthor{},
			Color:	0x36393E,
			Description: plot,
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:	"Genre",
					Value:	genre,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:	"IMDB Rating",
					Value:	imdbRating,
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: poster,
			},
			Title: title,
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	} else if args[0] == "!til" {

		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://www.reddit.com/r/todayilearned/top.json?limit=100", nil)
		if err != nil {
				log.Println(err)
				return
		}

		req.Header.Set("User-Agent", "KAPPA")

		resp, err := client.Do(req)
		if err != nil {
				log.Println(err)
				return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
				log.Println(err)
				return
		}

		v, err := jason.NewObjectFromBytes([]byte(body))
		if err != nil {
				log.Println(err)
				return
		}

		data, err := v.GetObjectArray("data", "children")
		if err != nil {
				log.Println(err)
				return
		}

		r := random(0, 99)
		title, err := data[r].GetString("data", "title")
		if err != nil {
			log.Println(err)
			return
		}

		s.ChannelMessageSend(m.ChannelID, title)
	} else if args[0] == "!shower" {

		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://www.reddit.com/r/Showerthoughts/top.json?limit=100", nil)
		if err != nil {
				log.Println(err)
				return
		}

		req.Header.Set("User-Agent", "KAPPA")

		resp, err := client.Do(req)
		if err != nil {
				log.Println(err)
				return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
				log.Println(err)
				return
		}

		v, err := jason.NewObjectFromBytes([]byte(body))
		if err != nil {
				log.Println(err)
				return
		}

		data, err := v.GetObjectArray("data", "children")
		if err != nil {
				log.Println(err)
				return
		}

		r := random(0, 99)
		title, err := data[r].GetString("data", "title")
		if err != nil {
			log.Println(err)
			return
		}

		s.ChannelMessageSend(m.ChannelID, title)
	} else if args[0] == "!earth" {

		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://www.reddit.com/r/EarthPorn/top.json?limit=100", nil)
		if err != nil {
				log.Println(err)
				return
		}

		req.Header.Set("User-Agent", "KAPPA")

		resp, err := client.Do(req)
		if err != nil {
				log.Println(err)
				return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
				log.Println(err)
				return
		}

		v, err := jason.NewObjectFromBytes([]byte(body))
		if err != nil {
				log.Println(err)
				return
		}

		data, err := v.GetObjectArray("data", "children")
		if err != nil {
				log.Println(err)
				return
		}

		for {
			r := random(0, 99)
			url, err := data[r].GetString("data", "url")
			if err != nil {
				log.Println(err)
				return
			}
			if strings.HasPrefix(url, "https://i.redd.it/") {
				embed := &discordgo.MessageEmbed{
					Author:	&discordgo.MessageEmbedAuthor{},
					Color:	0x36393E,
					Image: &discordgo.MessageEmbedImage{
						URL: url,
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
				break
			}
		}
	} else if args[0] == "!t" {

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!t es <text>`")
			return
		} else {
			t := strings.TrimPrefix(m.Content, m.Content[0:6])
			d := strings.NewReader(t)
			reqID, err := http.NewRequest("POST", "https://gateway.watsonplatform.net/language-translator/api/v3/identify?version=2018-05-01", d)
			if err != nil {
				log.Println(err)
				return
			}

			reqID.SetBasicAuth(APIKey4, APIKey5)
			reqID.Header.Set("Content-Type", "text/plain")

			respID, err := http.DefaultClient.Do(reqID)

			if err != nil {
				log.Println(err)
				return
			}

			defer respID.Body.Close()

			bodyID, err := ioutil.ReadAll(respID.Body)
			if err != nil {
				log.Println(err)
				return
			}

			vID, err := jason.NewObjectFromBytes([]byte(bodyID))
			if err != nil {
				log.Println(err)
				return
			}

			dataID, err := vID.GetObjectArray("languages")
			if err != nil {
				log.Println(err)
				return
			}

			id, err := dataID[0].GetString("language")
			if err != nil {
				log.Println(err)
				return
			}

			model := id + "-" + args[1]

			type Payload struct {
				Text	string	`json:"text"`
				ModelID	string	`json:"model_id"`
			}

			payload := Payload{}
			payload.Text = t
			payload.ModelID = model

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Println(err)
				return
			}

			body := bytes.NewReader(payloadBytes)

			reqT, err := http.NewRequest("POST", "https://gateway.watsonplatform.net/language-translator/api/v3/translate?version=2018-05-01", body)
			if err != nil {
				log.Println(err)
				return

			}

			reqT.SetBasicAuth(APIKey4, APIKey5)
			reqT.Header.Set("Content-Type", "application/json")

			respT, err := http.DefaultClient.Do(reqT)
			if err != nil {
				log.Println(err)
				return
			}

			defer respT.Body.Close()

			bodyT, err := ioutil.ReadAll(respT.Body)
			if err != nil {
				log.Println(err)
				return
			}

			vT, err := jason.NewObjectFromBytes([]byte(bodyT))
			if err != nil {
				log.Println(err)
				return
			}

			translations, err := vT.GetObjectArray("translations")
			if err != nil {
				log.Println(err)
				return
			}

			translation, err := translations[0].GetString("translation")
			if err != nil {
				log.Println(err)
				return
			}

			s.ChannelMessageSend(m.ChannelID, translation)
		}
	} else if args[0] == "!wb" {

		if len(args) == 1 || args[1] == "help" || args[1] == "h" {
			s.ChannelMessageSend(m.ChannelID, "Usage : `!wb example.com`")
			return
		} else {
			resp, err := http.Get("https://archive.org/wayback/available?url=" + args[1])

			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}

			v, err := jason.NewObjectFromBytes([]byte(body))
			if err != nil {
				log.Println(err)
				return
			}

			url, err := v.GetString("archived_snapshots", "closest", "url")
			if err != nil {
				log.Println(err)
				return
			}

			s.ChannelMessageSend(m.ChannelID, url)
		}
	} else if args[0] == "!help" || args[0] == "!h" {
		embed := &discordgo.MessageEmbed{
			Author:	&discordgo.MessageEmbedAuthor{},
			Color:	0x36393E,
			Description: ":white_check_mark: Check https://github.com/perfekto1337/discord_bot for a list of commands",
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}

}
