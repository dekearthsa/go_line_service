package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	controller "service_line_furk/controller"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

// func MqttConnection() {
// 	broker := "tcp://192.168.121.160:1883"
// 	clientID := "go_publisher"
// 	opts := mqtt.NewClientOptions()
// 	opts.AddBroker(broker)
// 	opts.SetClientID(clientID)

// 	// Create a new MQTT client
// 	client := mqtt.NewClient(opts)
// 	mqttClient = client
// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
// 		fmt.Printf("Failed to connect to broker: %v\n", token.Error())
// 		return
// 	}
// }

func main() {
	// controller.MqttConnection()
	router := gin.New()

	router.Static("/static", "./image")
	router.GET("/debug", controller.ControllerDebug)
	router.POST("/api/line", func(c *gin.Context) {
		broker := "tcp://192.168.1.43:1883"
		clientID := "go_publisher"

		// Create a new MQTT client options and configure the broker address
		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker)
		opts.SetClientID(clientID)

		// Create a new MQTT client
		client_mqtt := mqtt.NewClient(opts)
		if token := client_mqtt.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("Failed to connect to broker: %v\n", token.Error())
			return
		}

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
						token := client_mqtt.Publish(topic, 0, false, commands)
						token.Wait()
						client_mqtt.Disconnect(250)
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Koyuki กำลังเปิดไฟ LED ค่ะ")).Do(); err != nil {
							log.Print(err)
						}
					} else if message.Text == "LED-OFF" {
						topic := "sensor/control"
						commands := "{'command': 0}"
						token := client_mqtt.Publish(topic, 0, false, commands)
						token.Wait()
						client_mqtt.Disconnect(250)
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Koyuki กำลังปิดไฟ LED ค่ะ")).Do(); err != nil {
							log.Print(err)
						}
					}

				case *linebot.ImageMessage:
					// fmt.Println("msg img => ", message.ID)

					err := controller.DownloadImage(client, message.ID)
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

					err := controller.DownloadVideo(client, message.ID)
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
					err := controller.Downloadaudio(client, message.ID)
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
					desc, err := controller.DownloadFile(client, message.ID, message.FileName)
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
	})

	err := router.Run(":1111")
	if err != nil {
		fmt.Println("error => ", err)
	}
}
