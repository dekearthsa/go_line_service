package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

var mqttClient mqtt.Client

func ControllerDebug(c *gin.Context) {
	c.JSON(200, gin.H{"debug": "ok"})
}

func MqttConnection() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())

	}
}

func DownloadImage(bot *linebot.Client, messageID string) error {
	content, err := bot.GetMessageContent(messageID).Do()
	if err != nil {
		// log.Print(err)
		return err
	}
	defer content.Content.Close()

	// Save the image to a file (this example saves it as "received_image.jpg")
	filePath := filepath.Join("./image", messageID+"_received_image.jpg")
	file, err := os.Create(filePath)
	if err != nil {
		// log.Print(err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content.Content)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func DownloadVideo(bot *linebot.Client, messageID string) error {
	content, err := bot.GetMessageContent(messageID).Do()
	if err != nil {
		return err
	}
	defer content.Content.Close()
	filepath := filepath.Join("./video", messageID+"_video.mp4")
	file, err := os.Create(filepath)
	defer file.Close()
	_, err = io.Copy(file, content.Content)
	if err != nil {
		return err
	}
	return nil
}

func Downloadaudio(bot *linebot.Client, messageID string) error {
	content, err := bot.GetMessageContent(messageID).Do()
	if err != nil {
		return err
	}
	defer content.Content.Close()
	filepath := filepath.Join("./audio", messageID+"_audio.m4a")
	file, err := os.Create(filepath)
	defer file.Close()
	_, err = io.Copy(file, content.Content)
	if err != nil {
		return err
	}
	return nil
}

func PrefixString(fileName string) (string, string) {
	listData := strings.Split(fileName, ".")
	log.Println("listData => ", listData)
	return listData[0], listData[1]
}

func CheckTypeFile(ext string) bool {
	condition := []string{
		"pdf",
		"m4a",
	}
	for _, el := range condition {
		if el == ext {
			return true
		}
	}
	return false
}

func DownloadFile(bot *linebot.Client, messageID string, fileName string) (string, error) {
	name, ext := PrefixString(fileName)
	invalidFileStatus := CheckTypeFile(ext)
	if !invalidFileStatus {
		return "Invalid file.", nil
	}
	content, err := bot.GetMessageContent(messageID).Do()
	if err != nil {
		return "", err
	}
	defer content.Content.Close()
	filepath := filepath.Join("./files", messageID+"_"+name+"."+ext)
	file, err := os.Create(filepath)
	defer file.Close()
	_, err = io.Copy(file, content.Content)
	if err != nil {
		return "", err
	}
	return "", nil
}

func ControllerLineReplyMsg(c *gin.Context) {
	MqttConnection()
	const LOCALIMAGE_URL = "https://da2b-49-229-126-85.ngrok-free.app/static"
	const CH_SECRET = "5ef71f88c9f2b51ef476624c4386d4a3"
	const TOKEN = "QMgC0s/UrLHAwAOUvCRkGQyFMbnMX4p+P3MdfwyY3NDY8P4Q7XWNAo30ibED9XUbQ61kj9QFka9Lc3YvxV+fGSnec7h+cuYBrKXVVQhPxSWNGUx0iR+HQu9hKak+FcvP9dz1w3OWHSNopDYqO4i3MgdB04t89/1O/w1cDnyilFU="
	client, err := linebot.New(CH_SECRET, TOKEN)
	if err != nil {
		log.Fatal(err)
	}
	events, err := client.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.AbortWithStatus(http.StatusBadRequest)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}

	_, err = json.Marshal(events)
	// fmt.Println("jsonEvt => ", jsonEvt)
	// log.Printf("Event: %s", jsonEvt)
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			/// type text ///
			case *linebot.TextMessage:
				fmt.Println("msg text => ", event.Message)
				if message.Text == "debug" {
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "LED-ON" {
					topic := "sensor/control"
					commands := "{'command': 1}"
					token := mqttClient.Publish(topic, 0, false, commands)
					fmt.Println("token => ", token)
					token.Wait()
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "LED-OFF" {
					topic := "sensor/control"
					commands := "{'command': 0}"
					token := mqttClient.Publish(topic, 0, false, commands)
					token.Wait()
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}

			case *linebot.ImageMessage:
				// fmt.Println("msg img => ", message.ID)

				err := DownloadImage(client, message.ID)
				if err != nil {
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("can't receive image!")).Do(); err != nil {
						log.Print(err)
					}
				} else {
					replayMultiMessage := []linebot.SendingMessage{
						linebot.NewTextMessage("Received image!"),
						linebot.NewImageMessage(LOCALIMAGE_URL+"/test.jpg", LOCALIMAGE_URL+"/test.jpg"),
					}
					if _, err = client.ReplyMessage(event.ReplyToken, replayMultiMessage...).Do(); err != nil {
						log.Print(err)
					}
					// if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Received image!")).Do(); err != nil {
					// 	log.Print(err)
					// }
					// if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewImageMessage(LOCALIMAGE_URL+"/test.jpg", LOCALIMAGE_URL+"/test.jpg")).Do(); err != nil {
					// 	log.Print(err)
					// }
				}

			case *linebot.VideoMessage:
				// log.Printf("jsonEvt.message.id video => %s", jsonEvt.message.id)

				err := DownloadVideo(client, message.ID)
				if err != nil {
					log.Println("error in download video => ", err)
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Can't save video!")).Do(); err != nil {
						log.Print(err)
					}
				} else {
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Received your video!")).Do(); err != nil {
						log.Print(err)
					}
				}

			case *linebot.AudioMessage:
				err := Downloadaudio(client, message.ID)
				if err != nil {
					log.Println("error in download audio => ", err)
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Can't save audio!")).Do(); err != nil {
						log.Print(err)
					}
				} else {
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Received your audio!")).Do(); err != nil {
						log.Print(err)
					}
				}

			case *linebot.FileMessage:
				desc, err := DownloadFile(client, message.ID, message.FileName)
				if err != nil {
					log.Println("error in file file => ", err)
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Can't save file!")).Do(); err != nil {
						log.Print(err)
					}
				} else {
					if desc != "" {
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(desc)).Do(); err != nil {
							log.Print(err)
						}
					} else if desc == "" {
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Received your file!")).Do(); err != nil {
							log.Print(err)
						}
						// if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewImageMessage(LOCALIMAGE_URL+"/test.jpg", LOCALIMAGE_URL+"/test.jpg")).Do(); err != nil {
						// 	log.Print(err)
						// }
					}

				}
			}
		}
	}
	c.Status(http.StatusOK)
}
