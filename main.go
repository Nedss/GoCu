package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Global var from command line
var (
	Token    string
	DictPath string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.StringVar(&DictPath, "d", "", "Path of the dictionary file")
	flag.Parse()
}

func main() {
	discordBot, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("error creating Discord session", err)
		return
	}

	discordBot.AddHandler(commandCreate)
	discordBot.AddHandler(interactionCreate)

	// Only listen receiving message events
	discordBot.Identify.Intents = discordgo.IntentsGuildMessages

	err = discordBot.Open()
	if err != nil {
		log.Fatal("error opening connection", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discordBot.Close()
}

func randNumber(max int) (int, error) {
	rand.Seed(time.Now().UnixNano())
	min := 0
	number := rand.Intn(max-min) + min
	return number, nil
}

func getRandomWord(wordList []string) (string, error) {
	wordNumber := len(wordList)
	number, err := randNumber(wordNumber)
	if err != nil {
		log.Fatal("Cannot get max number ", err)
	}

	word := wordList[number]
	return word, nil
}

func parseDict(dictionaryPath string) ([]string, error) {
	dictFile, err := os.Open(dictionaryPath)
	if err != nil {
		log.Fatal("cannot open dictionary file : ", err)
	}

	defer dictFile.Close()

	scanner := bufio.NewScanner(dictFile)

	scanner.Split(bufio.ScanLines)

	var lines []string

	for scanner.Scan() {
		word := strings.Split(scanner.Text(), "\t")[0]
		lines = append(lines, word)
	}

	return lines, nil
}

func commandCreate(s *discordgo.Session, event *discordgo.Ready) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "cul",
		Description: "Fais lire aléatoirement le dictionnaire du cul au bot.",
	})
	if err != nil {
		log.Fatal("Cannot create command:", err)
		return
	}
}

func messageCreate() (string, error) {
	wordList, err := parseDict(DictPath)
	if err != nil {
		return "", fmt.Errorf("error parsing dictionary file %d\n", err)
	}

	word, err := getRandomWord(wordList)
	if err != nil {
		return "", fmt.Errorf("cannot get a word from dictionary %d\n", err)
	}
	fmt.Println("Returning : ", word)
	completeSentence := word + " du cul !"
	// s.ChannelMessageSendReply(m.ChannelID, completeSentence, m.Reference())
	return completeSentence, nil
}

func interactionCreate(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	if i.ApplicationCommandData().Name != "cul" {
		return
	}

	var response string
	var err error
	response, err = messageCreate()

	fmt.Println("Recieve a message.")
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Fatal("error sending interaction response", err)
		return
	}
}
